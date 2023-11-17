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
	"github.com/Nerzal/gocloak/v13"
	_ "github.com/influxdata/influxdb1-client"
	"log"
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

func (cs CleanupService) StartCleanupService(recreatePipes bool) {
	/****************************
	Check analytics pipelines
	****************************/
	if recreatePipes {
		info, err := cs.keycloak.GetUserInfo()
		if err != nil {
			return
		}

		pipes, errs := cs.pipeline.GetPipelines(*info.Sub, cs.keycloak.GetAccessToken())
		if len(errs) > 0 {
			fmt.Println(errs[0].Error())
			return
		}

		workloads, _ := cs.driver.GetWorkloads(PIPELINE)

		err = cs.recreatePipelines(pipes, workloads)
		if err != nil {
			log.Fatal("recreatePipelines failed: " + err.Error())
		}
	}

	/*
		cs.deleteOrphanedPipelineServices()
		cs.deleteOrphanedAnalyticsWorkloads()
		cs.deleteOrphanedKubeServices(PIPELINE)
		cs.deleteOrphanedKafkaTopics()

		/****************************
			Check analytics serving
		****************************/

	//cs.recreateServingServices()
	/*
		cs.deleteOrphanedServingServices()
		cs.deleteOrphanedServingWorkloads()
		cs.deleteOrphanedKubeServices(SERVING)
	*/
	//cs.deleteOrphanedInfluxMeasurements()
}

func (cs CleanupService) getOrphanedPipelineServices() (orphanedPipelineWorkloads []Pipeline, errs []error) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		pipes, errss := cs.pipeline.GetPipelines(*user.Sub, cs.keycloak.GetAccessToken())
		if len(errss) > 0 {
			errs = append(errs, errss...)
			return
		}
		workloads, e := cs.driver.GetWorkloads(PIPELINE)
		if e != nil {
			errs = append(errs, e)
			return
		}
		if len(errs) < 1 {
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
	return
}

func (cs CleanupService) deleteOrphanedPipelineService(id string, accessToken string) []error {
	return cs.pipeline.DeletePipeline(id, accessToken)
}

func (cs CleanupService) deleteOrphanedPipelineServices() (pipes []Pipeline, errs []error) {
	pipes, errs = cs.getOrphanedPipelineServices()
	if len(errs) < 1 {
		for _, pipe := range pipes {
			errs = cs.pipeline.DeletePipeline(pipe.Id, cs.keycloak.GetAccessToken())
			if len(errs) > 0 {
				return
			}
		}
	}
	return
}

func (cs CleanupService) getOrphanedAnalyticsWorkloads() (orphanedAnalyticsWorkloads []Workload, errs []error) {
	user, err := cs.keycloak.GetUserInfo()
	if err != nil {
		log.Fatal("GetUserInfo failed:" + err.Error())
	}
	if user != nil {
		pipes, errs := cs.pipeline.GetPipelines(*user.Sub, cs.keycloak.GetAccessToken())
		workloads, e := cs.driver.GetWorkloads(PIPELINE)
		if e != nil {
			errs = append(errs, e)
		}
		if len(errs) < 1 {
			for _, workload := range workloads {
				if !workloadInPipes(workload, pipes) {
					orphanedAnalyticsWorkloads = append(orphanedAnalyticsWorkloads, workload)
				}
			}
		}
	}
	return
}

func (cs CleanupService) deleteOrphanedAnalyticsWorkload(name string) error {
	return cs.driver.DeleteWorkload(name, PIPELINE)
}

func (cs CleanupService) deleteOrphanedAnalyticsWorkloads() (workloads []Workload, errs []error) {
	workloads, errs = cs.getOrphanedAnalyticsWorkloads()
	if len(errs) < 1 {
		for _, workload := range workloads {
			err := cs.driver.DeleteWorkload(workload.Name, PIPELINE)
			if err != nil {
				errs = append(errs, err)
				return
			}
		}
	}
	return
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

func (cs CleanupService) deleteOrphanedKafkaTopics() (errs []error) {
	topics, errs := cs.getOrphanedKafkaTopics()
	if len(errs) > 0 {
		return
	}
	for _, topic := range topics {
		err := cs.kafkaAdmin.DeleteTopic(topic)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return
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

func (cs CleanupService) recreatePipelines(pipelines []Pipeline, workloads []Workload) error {
	cs.logger.Print("**************** Recreate Pipelines *********************")
	for _, pipeline := range pipelines {
		if !pipeInWorkloads(pipeline, workloads) {
			cs._logPrint(pipeline.Id, pipeline.Name, pipeline.UserId)
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
