/*
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lib

import (
	"fmt"
	"github.com/Nerzal/gocloak/v5"
	_ "github.com/influxdata/influxdb1-client"
	influxClient "github.com/influxdata/influxdb1-client/v2"
	"log"
	"strconv"
	"strings"
)

type CleanupService struct {
	keycloak   KeycloakService
	driver     Driver
	pipeline   PipelineService
	serving    ServingService
	logger     Logger
	influx     *Influx
	kafkaAdmin *KafkaAdmin
}

const DividerString = "++++++++++++++++++++++++++++++++++++++++++++++++++++++++"

func NewCleanupService(keycloak KeycloakService, driver Driver, pipeline PipelineService, serving ServingService, logger Logger, kafkaAdmin *KafkaAdmin) *CleanupService {
	influx := NewInflux()
	return &CleanupService{keycloak: keycloak, driver: driver, pipeline: pipeline, serving: serving, logger: logger, influx: influx, kafkaAdmin: kafkaAdmin}
}

func (cs CleanupService) StartCleanupService() {
	cs.keycloak.Login()
	defer cs.keycloak.Logout()
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		/****************************
			Check analytics pipelines
		****************************/

		pipes, err := cs.pipeline.GetPipelines(*user.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("GetPipelines failed: " + err.Error())
		}
		workloads, err := cs.driver.GetWorkloads("pipelines")
		if err != nil {
			log.Fatal("GetWorkloads for pipelines failed: " + err.Error())
		}
		cs.checkPipelineWorkloads(pipes, workloads)
		cs.checkPipes(workloads, pipes)
		cs.checkKafkaTopics()

		/****************************
			Check analytics serving
		****************************/

		servings, err := cs.serving.GetServingServices(*user.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("GetServingServices failed: " + err.Error())
		}

		workloads, err = cs.driver.GetWorkloads("serving")
		if err != nil {
			log.Fatal("GetWorkloads for serving instances failed: " + err.Error())
		}

		services, err := cs.driver.GetServices("serving")
		if err != nil {
			log.Fatal("GetServices for serving instances failed: " + err.Error())
		}

		_, err = cs.getInfluxData()
		if err != nil {
			log.Fatal("Get Influx data for serving instances failed: " + err.Error())
		}
		//cs.recreateServingServices(servings, services)
		cs.checkServingWorkloads(servings, workloads)
		cs.checkServings(workloads, servings)
		cs.checkServingServices(workloads, services)
		//cs.checkInfluxMeasurements(influxData, servings)
	}
}

func (cs CleanupService) checkServingServices(workloads []Workload, services []Service) {
	cs.logger.Print("************** Orphaned Serving Services ***************")
	for _, service := range services {
		if !serviceInWorkloads(service, workloads) {
			cs.logPrint(service.Name, service.Id)
			err := cs.driver.DeleteService(service.Id, "serving")
			if err != nil {
				log.Fatal("DeleteService failed: " + err.Error())
			}
		}
	}
}

func (cs CleanupService) checkInfluxMeasurements(influxData map[string][]string, servings []ServingInstance) {
	cs.logger.Print("**************** Orphaned Measurements *****************")
	for db, measurements := range influxData {
		for _, measurement := range measurements {
			if !influxMeasurementInServings(measurement, servings) {
				cs.logger.Print(DividerString)
				cs.logger.Print(db + ":" + measurement)
				errors := cs.influx.DropMeasurement(measurement, db)
				if len(errors) > 0 {
					log.Fatal("could not delete measurement: " + errors[0].Error())
				}
			}
		}
	}
}

func (cs CleanupService) checkPipelineWorkloads(pipes []Pipeline, workloads []Workload) {
	cs.logger.Print("**************** Orphaned Pipelines ********************")
	for _, pipe := range pipes {
		if !pipeInWorkloads(pipe, workloads) {
			deletePipe := true

			for _, operator := range pipe.Operators {
				if operator.DeploymentType == "local" {
					deletePipe = false
					break
				}
			}

			if deletePipe {
				user := cs.getKeycloakUserById(pipe.UserId)
				cs.logPrint(pipe.Id, pipe.Name, strconv.Itoa(len(pipe.Operators)), *user.Username)
				for _, operator := range pipe.Operators {
					cs.logger.Print(operator.Name)
					cs.logger.Print(operator.ImageId)
				}
				err := cs.pipeline.DeletePipeline(pipe.Id, pipe.UserId, cs.keycloak.GetAccessToken())
				if err != nil {
					log.Fatal("DeletePipeline failed:" + err.Error())
				}
			}
		}

	}
}

func (cs CleanupService) checkPipes(workloads []Workload, pipes []Pipeline) {
	cs.logger.Print("************** Orphaned Pipeline Services **************")
	for _, workload := range workloads {
		if !workloadInPipes(workload, pipes) {
			cs.logPrint(workload.Name, workload.Id, workload.ImageUuid)
			err := cs.driver.DeleteWorkload(workload.Name, "")
			if err != nil {
				log.Fatal("DeleteWorkload failed: " + err.Error())
			}
		}
	}
}

func (cs CleanupService) checkServingWorkloads(serving []ServingInstance, workloads []Workload) {
	cs.logger.Print("**************** Orphaned Servings *********************")
	for _, serving := range serving {
		if !servingInWorkloads(serving, workloads) {
			user := cs.getKeycloakUserById(serving.UserId)
			cs.logPrint(serving.ID.String(), serving.Name, *user.Username)
			err := cs.serving.DeleteServingService(serving.ID.String(), serving.UserId, cs.keycloak.GetAccessToken())
			if err != nil {
				log.Fatal("DeleteServingService failed:" + err.Error())
			}
		}

	}
}

func (cs CleanupService) checkServings(workloads []Workload, servings []ServingInstance) {
	cs.logger.Print("************** Orphaned Serving Workloads ***************")
	for _, workload := range workloads {
		if strings.Contains(workload.Name, "kafka-influx") || strings.Contains(workload.Name, "kafka2influx") {
			if !workloadInServings(workload, servings) {
				cs.logPrint(workload.Name, workload.Id, workload.ImageUuid)
				err := cs.driver.DeleteWorkload(workload.Name, "serving")
				if err != nil {
					log.Fatal("DeleteWorkload failed: " + err.Error())
				}
			}
		}
	}
}

func (cs CleanupService) checkKafkaTopics() {
	cs.logger.Print("************** Orphaned Kafka Topics ***************")
	envs, err := cs.driver.GetWorkloadEnvs("pipelines")
	if err != nil {
		log.Fatal("GetWorkloads for pipelines failed: " + err.Error())
	}
	topics, err := cs.kafkaAdmin.GetTopics()
	if err != nil {
		log.Fatal("error in GetTopics", err.Error())
	}
	for _, topic := range topics {
		if isInternalAnalyticsTopic(topic) && !pipelineExists(topic, envs) {
			cs.logger.Print(topic)
			err = cs.kafkaAdmin.DeleteTopic(topic)
			if err != nil {
				log.Fatal("DeleteTopic failed", err.Error())
			}
		}
	}
}

func (cs CleanupService) getInfluxData() (influxDbs map[string][]string, err error) {
	q := influxClient.NewQuery("SHOW DATABASES", "", "")
	response, err := cs.influx.client.Query(q)
	if err != nil {
		return
	}
	if response.Error() != nil {
		err = response.Error()
		return
	}
	influxDbs = make(map[string][]string)
	for _, val := range response.Results[0].Series[0].Values {
		str := fmt.Sprint(val)
		influxDbs[strings.Replace(strings.Replace(str, "]", "", -1), "[", "", -1)] = []string{}
	}
	for db := range influxDbs {
		if db != "_internal" {
			q := influxClient.NewQuery("SHOW MEASUREMENTS", db, "")
			response, err = cs.influx.client.Query(q)
			if err != nil {
				return
			}
			if response.Error() != nil {
				err = response.Error()
				return
			}

			if len(response.Results[0].Series) > 0 {
				for _, measurement := range response.Results[0].Series[0].Values {
					influxDbs[db] = append(influxDbs[db], measurement[0].(string))
				}
			}

		}
	}
	return influxDbs, err
}

func (cs CleanupService) recreateServingServices(serving []ServingInstance, workloads []Workload) {
	cs.logger.Print("**************** Recreate Servings *********************")
	for _, serving := range serving {
		if !servingInWorkloads(serving, workloads) {
			cs.logPrint(serving.ID.String(), serving.Name)
			if GetEnv("FORCE_DELETE", "") != "" && serving.Database == GetEnv("FORCE_DELETE", "") {
				cs.influx.forceDeleteMeasurement(serving)
			}
			cs.create(serving)
		}

	}
}

func (cs *CleanupService) create(serving ServingInstance) {
	dataFields := "{"
	for index, value := range serving.Values {
		dataFields = dataFields + "\"" + value.Name + ":" + value.Type + "\":\"" + value.Path + "\""
		if index+1 < len(serving.Values) {
			dataFields = dataFields + ","
		}
	}
	dataFields = dataFields + "}"
	cs.driver.CreateServingInstance(&serving, dataFields)
}

func (cs *CleanupService) getKeycloakUserById(id string) (user *gocloak.User) {
	user, err := cs.keycloak.GetUserByID(id)
	if err != nil {
		log.Fatal("GetUserByID failed:" + err.Error())
	}
	if user == nil {
		log.Fatal("could not get user data from keycloak")
	}
	return
}

func (cs *CleanupService) logPrint(vars ...string) {
	cs.logger.Print(DividerString)
	for _, v := range vars {
		cs.logger.Print(v)
	}
}
