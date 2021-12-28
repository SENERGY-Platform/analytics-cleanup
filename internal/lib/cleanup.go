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
	logger     FileLogger
	influx     *Influx
	kafkaAdmin *KafkaAdmin
}

const DividerString = "++++++++++++++++++++++++++++++++++++++++++++++++++++++++"

func NewCleanupService(keycloak KeycloakService, driver Driver, pipeline PipelineService, serving ServingService, logger FileLogger, kafkaAdmin *KafkaAdmin) *CleanupService {
	influx := NewInflux()
	return &CleanupService{keycloak: keycloak, driver: driver, pipeline: pipeline, serving: serving, logger: logger, influx: influx, kafkaAdmin: kafkaAdmin}
}

func (cs CleanupService) StartCleanupService() {
	/****************************
	Check analytics pipelines
	****************************/
	/*
		err = cs.recreatePipelines(pipes, workloads)
		if err != nil {
			log.Fatal("recreatePipelines failed: " + err.Error())
		}
	*/
	cs.deleteOrphanedPipelineServices()
	cs.deleteOrphanedAnalyticsWorkloads()
	cs.checkKafkaTopics()

	/****************************
		Check analytics serving
	****************************/

	//cs.recreateServingServices(servings, services)
	cs.deleteOrphanedServingServices()
	cs.deleteOrphanedServingWorkloads()
	cs.deleteOrphanedServingKubeServices()
	//cs.deleteOrphanedInfluxMeasurements()
}

func (cs CleanupService) getOrphanedPipelineServices() (orphanedPipelineWorkloads []Pipeline) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		pipes, err := cs.pipeline.GetPipelines(*user.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("GetPipelines failed: " + err.Error())
		}
		workloads, err := cs.driver.GetWorkloads("pipelines")
		if err != nil {
			log.Fatal("GetWorkloads for pipelines failed: " + err.Error())
		}
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
					orphanedPipelineWorkloads = append(orphanedPipelineWorkloads, pipe)
				}
			}

		}
	}
	return
}

func (cs CleanupService) deleteOrphanedPipelineServices() {
	cs.logger.Print("**************** Orphaned Pipelines ********************")
	for _, pipe := range cs.getOrphanedPipelineServices() {
		user := cs._getKeycloakUserById(pipe.UserId)
		cs._logPrint(pipe.Id, pipe.Name, strconv.Itoa(len(pipe.Operators)), *user.Username)
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

func (cs CleanupService) getOrphanedAnalyticsWorkloads() (orphanedAnalyticsWorkloads []Workload) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		pipes, err := cs.pipeline.GetPipelines(*user.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("GetPipelines failed: " + err.Error())
		}
		workloads, err := cs.driver.GetWorkloads("pipelines")
		if err != nil {
			log.Fatal("GetWorkloads for pipelines failed: " + err.Error())
		}
		for _, workload := range workloads {
			if !workloadInPipes(workload, pipes) {
				orphanedAnalyticsWorkloads = append(orphanedAnalyticsWorkloads, workload)
			}
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedAnalyticsWorkloads() {
	cs.logger.Print("************** Orphaned Pipeline Services **************")
	for _, workload := range cs.getOrphanedAnalyticsWorkloads() {
		cs._logPrint(workload.Name, workload.Id, workload.ImageUuid)
		err := cs.driver.DeleteWorkload(workload.Name, "")
		if err != nil {
			log.Fatal("DeleteWorkload failed: " + err.Error())
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

func (cs CleanupService) getOrphanedServingServices() (orphanedServingWorkloads []ServingInstance) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		servings, err := cs.serving.GetServingServices(*user.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("GetServingServices failed: " + err.Error())
		}

		workloads, err := cs.driver.GetWorkloads("serving")
		if err != nil {
			log.Fatal("GetWorkloads for serving instances failed: " + err.Error())
		}
		for _, serving := range servings {
			if !servingInWorkloads(serving, workloads) {
				orphanedServingWorkloads = append(orphanedServingWorkloads, serving)
			}

		}
	}
	return
}

func (cs CleanupService) deleteOrphanedServingServices() {
	cs.logger.Print("**************** Orphaned Servings *********************")
	for _, serving := range cs.getOrphanedServingServices() {
		user := cs._getKeycloakUserById(serving.UserId)
		cs._logPrint(serving.ID.String(), serving.Name, *user.Username)
		err := cs.serving.DeleteServingService(serving.ID.String(), serving.UserId, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("DeleteServingService failed:" + err.Error())
		}
	}
}

func (cs CleanupService) getOrphanedServingWorkloads() (orphanedServingWorkloads []Workload) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		servings, err := cs.serving.GetServingServices(*user.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("GetServingServices failed: " + err.Error())
		}

		workloads, err := cs.driver.GetWorkloads("serving")
		if err != nil {
			log.Fatal("GetWorkloads for serving instances failed: " + err.Error())
		}
		for _, workload := range workloads {
			if strings.Contains(workload.Name, "kafka-influx") || strings.Contains(workload.Name, "kafka2influx") {
				if !workloadInServings(workload, servings) {
					orphanedServingWorkloads = append(orphanedServingWorkloads, workload)
				}
			}
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedServingWorkloads() {
	cs.logger.Print("************** Orphaned Serving Workloads ***************")
	for _, workload := range cs.getOrphanedServingWorkloads() {
		cs._logPrint(workload.Name, workload.Id, workload.ImageUuid)
		err := cs.driver.DeleteWorkload(workload.Name, "serving")
		if err != nil {
			log.Fatal("DeleteWorkload failed: " + err.Error())
		}
	}
}

func (cs CleanupService) getOrphanedServingKubeServices() []KubeService {
	workloads, err := cs.driver.GetWorkloads("serving")
	if err != nil {
		log.Fatal("GetWorkloads for serving instances failed: " + err.Error())
	}

	services, err := cs.driver.GetServices("serving")
	if err != nil {
		log.Fatal("GetServices for serving instances failed: " + err.Error())
	}
	var orphanedServices []KubeService
	for _, service := range services {
		if !serviceInWorkloads(service, workloads) {
			orphanedServices = append(orphanedServices, service)
		}
	}
	return orphanedServices
}

func (cs CleanupService) deleteOrphanedServingKubeServices() {
	cs.logger.Print("************** Orphaned Serving Services ***************")
	for _, service := range cs.getOrphanedServingKubeServices() {
		cs._logPrint(service.Name, service.Id)
		err := cs.driver.DeleteService(service.Id, "serving")
		if err != nil {
			log.Fatal("DeleteService failed: " + err.Error())
		}
	}
}

func (cs CleanupService) getOrphanedInfluxMeasurements() (orphanedInfluxMeasurements map[string][]string, err error) {
	influxData, err := cs._getInfluxData()
	if err != nil {
		return
	}
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		return
	}
	if user != nil {
		servings, e := cs.serving.GetServingServices(*user.Sub, cs.keycloak.GetAccessToken())
		if e != nil {
			return
		}
		for db, measurements := range influxData {
			for _, measurement := range measurements {
				if !influxMeasurementInServings(measurement, servings) {
					if orphanedInfluxMeasurements[db] == nil {
						orphanedInfluxMeasurements[db] = []string{}
					}
					orphanedInfluxMeasurements[db] = append(orphanedInfluxMeasurements[db], measurement)
				}
			}
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedInfluxMeasurements() {
	cs.logger.Print("**************** Orphaned Measurements *****************")
	influxData, err := cs.getOrphanedInfluxMeasurements()
	if err != nil {
		log.Fatal(err)
	}
	for db, measurements := range influxData {
		for _, measurement := range measurements {
			cs.logger.Print(DividerString)
			cs.logger.Print(db + ":" + measurement)
			errors := cs.influx.DropMeasurement(measurement, db)
			if len(errors) > 0 {
				log.Fatal("could not delete measurement: " + errors[0].Error())
			}
		}
	}
}

func (cs CleanupService) recreateServingServices(serving []ServingInstance, workloads []Workload) {
	cs.logger.Print("**************** Recreate Servings *********************")
	for _, serving := range serving {
		if !servingInWorkloads(serving, workloads) {
			cs._logPrint(serving.ID.String(), serving.Name)
			if GetEnv("FORCE_DELETE", "") != "" && serving.Database == GetEnv("FORCE_DELETE", "") {
				cs.influx.forceDeleteMeasurement(serving)
			}
			cs._create(serving)
		}

	}
}

func (cs CleanupService) recreatePipelines(pipelines []Pipeline, workloads []Workload) error {
	cs.logger.Print("**************** Recreate Pipelines *********************")
	for _, pipeline := range pipelines {
		if !pipeInWorkloads(pipeline, workloads) {
			cs._logPrint(pipeline.Id, pipeline.Name)
			request := pipeline.ToRequest()
			userToken, err := cs.keycloak.GetImpersonateToken(pipeline.UserId)
			if err != nil {
				return err
			}
			err = cs.pipeline.CreatePipeline(request, pipeline.UserId, userToken)
			if err != nil {
				cs.logger.Print(err.Error() + ", User: " + pipeline.UserId + ", Pipeline " + pipeline.Id)
			}
		}
	}
	return nil
}

func (cs CleanupService) _getInfluxData() (influxDbs map[string][]string, err error) {
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

func (cs *CleanupService) _create(serving ServingInstance) {
	dataFields, tagFields := servingGetDataAndTagFields(serving.Values)
	cs.driver.CreateServingInstance(&serving, dataFields, tagFields)
}

func (cs *CleanupService) _getKeycloakUserById(id string) (user *gocloak.User) {
	user, err := cs.keycloak.GetUserByID(id)
	if err != nil {
		log.Fatal("GetUserByID failed:" + err.Error())
	}
	if user == nil {
		log.Fatal("could not get user data from keycloak")
	}
	return
}

func (cs *CleanupService) _logPrint(vars ...string) {
	cs.logger.Print(DividerString)
	for _, v := range vars {
		cs.logger.Print(v)
	}
}
