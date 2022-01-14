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
	uuid "github.com/satori/go.uuid"
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
	cs.deleteOrphanedKubeServices(PIPELINE)
	cs.deleteOrphanedKafkaTopics()

	/****************************
		Check analytics serving
	****************************/

	//cs.recreateServingServices()
	cs.deleteOrphanedServingServices()
	cs.deleteOrphanedServingWorkloads()
	cs.deleteOrphanedKubeServices(SERVING)
	//cs.deleteOrphanedInfluxMeasurements()
}

func (cs CleanupService) getOrphanedPipelineServices() (orphanedPipelineWorkloads []Pipeline) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		pipes, errs := cs.pipeline.GetPipelines(*user.Sub, cs.keycloak.GetAccessToken())
		if len(errs) > 0 {
			log.Printf("GetPipelines failed: %s", errs)
			return
		}
		workloads, err := cs.driver.GetWorkloads(PIPELINE)
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

func (cs CleanupService) deleteOrphanedPipelineService(id string, accessToken string) []error {
	return cs.pipeline.DeletePipeline(id, accessToken)
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
		errs := cs.pipeline.DeletePipeline(pipe.Id, cs.keycloak.GetAccessToken())
		if len(errs) > 0 {
			log.Printf("DeletePipeline failed: %s", errs)
		}
	}
}

func (cs CleanupService) getOrphanedAnalyticsWorkloads() (orphanedAnalyticsWorkloads []Workload) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		pipes, errs := cs.pipeline.GetPipelines(*user.Sub, cs.keycloak.GetAccessToken())
		if len(errs) > 0 {
			log.Printf("GetPipelines failed: %s", errs)
			return
		}
		workloads, e := cs.driver.GetWorkloads(PIPELINE)
		if e != nil {
			log.Fatal("GetWorkloads for pipelines failed: " + e.Error())
		}
		for _, workload := range workloads {
			if !workloadInPipes(workload, pipes) {
				orphanedAnalyticsWorkloads = append(orphanedAnalyticsWorkloads, workload)
			}
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedAnalyticsWorkload(name string) error {
	return cs.driver.DeleteWorkload(name, PIPELINE)
}

func (cs CleanupService) deleteOrphanedAnalyticsWorkloads() {
	cs.logger.Print("************** Orphaned Pipeline Services **************")
	for _, workload := range cs.getOrphanedAnalyticsWorkloads() {
		cs._logPrint(workload.Name, workload.Id, workload.ImageUuid)
		err := cs.driver.DeleteWorkload(workload.Name, PIPELINE)
		if err != nil {
			log.Fatal("DeleteWorkload failed: " + err.Error())
		}
	}
}

func (cs CleanupService) getOrphanedKafkaTopics() (orphanedKafkaTopics []string, errs []error) {
	envs, err := cs.driver.GetWorkloadEnvs(PIPELINE)
	if err != nil {
		errs = append(errs, err)
		return
	}
	topics, err := cs.kafkaAdmin.GetTopics()
	if err != nil {
		errs = append(errs, err)
		return
	}
	for _, topic := range topics {
		if isInternalAnalyticsTopic(topic) && !pipelineExists(topic, envs) {
			orphanedKafkaTopics = append(orphanedKafkaTopics, topic)
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedKafkaTopic(topic string) error {
	return cs.kafkaAdmin.DeleteTopic(topic)
}

func (cs CleanupService) deleteOrphanedKafkaTopics() {
	cs.logger.Print("************** Orphaned Kafka Topics ***************")
	topics, errs := cs.getOrphanedKafkaTopics()
	if len(errs) > 0 {
		log.Fatalf("getOrphanedKafkaTopics failed: %s", errs)
	}
	for _, topic := range topics {
		cs.logger.Print(topic)
		err := cs.kafkaAdmin.DeleteTopic(topic)
		if err != nil {
			log.Fatal("DeleteTopic failed", err.Error())
		}
	}
}

func (cs CleanupService) getOrphanedServingServices() (orphanedServingWorkloads []ServingInstance) {
	servings, workloads := cs._getServingInstancesAndWorkloads()
	for _, serving := range servings {
		if !servingInWorkloads(serving, workloads) {
			orphanedServingWorkloads = append(orphanedServingWorkloads, serving)
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedServingService(id string, accessToken string) []error {
	return cs.serving.DeleteServingService(id, accessToken)
}

func (cs CleanupService) deleteOrphanedServingServices() {
	cs.logger.Print("**************** Orphaned Servings *********************")
	for _, serving := range cs.getOrphanedServingServices() {
		user := cs._getKeycloakUserById(serving.UserId)
		cs._logPrint(serving.ID.String(), serving.Name, *user.Username)
		errs := cs.serving.DeleteServingService(serving.ID.String(), cs.keycloak.GetAccessToken())
		if len(errs) > 0 {
			log.Printf("DeleteServingService failed: %s", errs)
		}
	}
}

func (cs CleanupService) getOrphanedServingWorkloads() (orphanedServingWorkloads []Workload) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		servings, errs := cs.serving.GetServingServices(*user.Sub, cs.keycloak.GetAccessToken())
		if len(errs) > 0 {
			log.Printf("GetServingServices failed: %s", errs)
			return
		}

		workloads, e := cs.driver.GetWorkloads(SERVING)
		if e != nil {
			log.Fatal("GetWorkloads for serving instances failed: " + e.Error())
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

func (cs CleanupService) deleteOrphanedServingWorkload(name string) error {
	return cs.driver.DeleteWorkload(name, SERVING)
}

func (cs CleanupService) deleteOrphanedServingWorkloads() {
	cs.logger.Print("************** Orphaned Serving Workloads ***************")
	for _, workload := range cs.getOrphanedServingWorkloads() {
		cs._logPrint(workload.Name, workload.Id, workload.ImageUuid)
		err := cs.driver.DeleteWorkload(workload.Name, SERVING)
		if err != nil {
			log.Fatal("DeleteWorkload failed: " + err.Error())
		}
	}
}

func (cs CleanupService) getOrphanedKubeServices(collection string) (orphanedServices []KubeService, errs []error) {
	workloads, err := cs.driver.GetWorkloads(collection)
	if err != nil {
		errs = append(errs, err)
	}
	services, err := cs.driver.GetServices(collection)
	if err != nil {
		errs = append(errs, err)
	}
	if len(errs) < 1 {
		for _, service := range services {
			if !serviceInWorkloads(service, workloads) {
				orphanedServices = append(orphanedServices, service)
			}
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedKubeService(collection string, id string) error {
	return cs.driver.DeleteService(id, collection)
}

func (cs CleanupService) deleteOrphanedKubeServices(collection string) (deletedKubeServices []KubeService, errs []error) {
	services, errs := cs.getOrphanedKubeServices(collection)
	if len(errs) < 1 {
		for _, service := range services {
			err := cs.driver.DeleteService(service.Id, collection)
			if err != nil {
				errs = append(errs, err)
			} else {
				deletedKubeServices = append(deletedKubeServices, service)
			}
		}
	}
	return
}

func (cs CleanupService) getOrphanedInfluxMeasurements() (orphanedInfluxMeasurements []InfluxMeasurement, err error) {
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
					orphanedInfluxMeasurements = append(orphanedInfluxMeasurements, InfluxMeasurement{Id: measurement, DatabaseId: db})
				}
			}
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedInfluxMeasurement(measurementId string, databaseId string) []error {
	return cs.influx.DropMeasurement(measurementId, databaseId)
}

func (cs CleanupService) deleteOrphanedInfluxMeasurements() {
	cs.logger.Print("**************** Orphaned Measurements *****************")
	influxData, err := cs.getOrphanedInfluxMeasurements()
	if err != nil {
		log.Fatal(err)
	}
	for _, measurement := range influxData {
		cs.logger.Print(DividerString)
		cs.logger.Print(measurement.DatabaseId + ":" + measurement.Id)
		errors := cs.influx.DropMeasurement(measurement.Id, measurement.DatabaseId)
		if len(errors) > 0 {
			log.Fatal("could not delete measurement: " + errors[0].Error())
		}
	}
}

func (cs CleanupService) recreateServingServices() {
	servings, workloads := cs._getServingInstancesAndWorkloads()
	cs.logger.Print("**************** Recreate Servings *********************")
	idMissing := false
	missingUuid, _ := uuid.FromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	for _, serving := range servings {
		idMissing = false
		if serving.ApplicationId == missingUuid {
			serving.ApplicationId = uuid.NewV4()
			idMissing = true
		}
		if idMissing && servingInWorkloads(serving, workloads) {
			err := cs.driver.DeleteWorkload("kafka2influx-"+serving.ID.String(), SERVING)
			if err != nil {
				log.Fatal(err)
			}
		}
		if idMissing || !servingInWorkloads(serving, workloads) {
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

func (cs CleanupService) _getServingInstancesAndWorkloads() (servings []ServingInstance, workloads []Workload) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		var errs []error
		servings, errs = cs.serving.GetServingServices(*user.Sub, cs.keycloak.GetAccessToken())
		if len(errs) > 0 {
			log.Printf("GetServingServices failed: %s", errs)
			return
		}

		workloads, err = cs.driver.GetWorkloads(SERVING)
		if err != nil {
			log.Fatal("GetWorkloads for serving instances failed: " + err.Error())
		}
	}
	return
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
