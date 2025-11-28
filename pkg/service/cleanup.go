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

package service

import (
	"fmt"
	"log"

	"github.com/Nerzal/gocloak/v13"
	"github.com/SENERGY-Platform/analytics-cleanup/lib"
	"github.com/SENERGY-Platform/analytics-cleanup/pkg/apis"
	"github.com/SENERGY-Platform/analytics-cleanup/pkg/util"

	pipeModels "github.com/SENERGY-Platform/analytics-pipeline/lib"
)

type CleanupService struct {
	keycloak   apis.KeycloakService
	driver     Driver
	pipeline   apis.PipelineService
	logger     util.FileLogger
	kafkaAdmin *apis.KafkaAdmin
}

const DividerString = "++++++++++++++++++++++++++++++++++++++++++++++++++++++++"

func NewCleanupService(keycloak apis.KeycloakService, driver Driver, pipeline apis.PipelineService, logger util.FileLogger, kafkaAdmin *apis.KafkaAdmin) *CleanupService {
	return &CleanupService{keycloak: keycloak, driver: driver, pipeline: pipeline, logger: logger, kafkaAdmin: kafkaAdmin}
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

		workloads, _ := cs.driver.GetWorkloads(lib.PIPELINE)

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
	*/
}

func (cs CleanupService) GetOrphanedPipelineServices(userId string, authToken string) (orphanedPipelineWorkloads []pipeModels.Pipeline, errs []error) {
	pipes, errss := cs.pipeline.GetPipelines(userId, authToken)
	if len(errss) > 0 {
		errs = append(errs, errss...)
		return
	}
	workloads, e := cs.driver.GetWorkloads(lib.PIPELINE)
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

func (cs CleanupService) DeleteOrphanedPipelineService(id string, accessToken string) []error {
	return cs.pipeline.DeletePipeline(id, accessToken)
}

func (cs CleanupService) DeleteOrphanedPipelineServices(userId string, authToken string) (pipes []pipeModels.Pipeline, errs []error) {
	pipes, errs = cs.GetOrphanedPipelineServices(userId, authToken)
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

func (cs CleanupService) GetOrphanedAnalyticsWorkloads(userId string, authToken string) (orphanedAnalyticsWorkloads []lib.Workload, errs []error) {
	pipes, errs := cs.pipeline.GetPipelines(userId, authToken)
	workloads, e := cs.driver.GetWorkloads(lib.PIPELINE)
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
	return
}

func (cs CleanupService) DeleteOrphanedAnalyticsWorkload(name string) error {
	return cs.driver.DeleteWorkload(name, lib.PIPELINE)
}

func (cs CleanupService) DeleteOrphanedAnalyticsWorkloads(userId string, authToken string) (workloads []lib.Workload, errs []error) {
	workloads, errs = cs.GetOrphanedAnalyticsWorkloads(userId, authToken)
	if len(errs) < 1 {
		for _, workload := range workloads {
			err := cs.driver.DeleteWorkload(workload.Name, lib.PIPELINE)
			if err != nil {
				errs = append(errs, err)
				return
			}
		}
	}
	return
}

func (cs CleanupService) GetOrphanedKafkaTopics() (orphanedKafkaTopics []string, errs []error) {
	envs, err := cs.driver.GetWorkloadEnvs(lib.PIPELINE)
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

func (cs CleanupService) DeleteOrphanedKafkaTopic(topic string) error {
	return cs.kafkaAdmin.DeleteTopic(topic)
}

func (cs CleanupService) DeleteOrphanedKafkaTopics() (errs []error) {
	topics, errs := cs.GetOrphanedKafkaTopics()
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

func (cs CleanupService) GetOrphanedKubeServices(collection string) (orphanedServices []lib.KubeService, errs []error) {
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

func (cs CleanupService) DeleteOrphanedKubeService(collection string, id string) error {
	return cs.driver.DeleteService(id, collection)
}

func (cs CleanupService) DeleteOrphanedKubeServices(collection string) (deletedKubeServices []lib.KubeService, errs []error) {
	services, errs := cs.GetOrphanedKubeServices(collection)
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

func (cs CleanupService) recreatePipelines(pipelines []pipeModels.Pipeline, workloads []lib.Workload) error {
	cs.logger.Print("**************** Recreate Pipelines *********************")
	for _, pipeline := range pipelines {
		if !pipeInWorkloads(pipeline, workloads) {
			cs._logPrint(pipeline.Id, pipeline.Name, pipeline.UserId)
			pipe := &lib.Pipeline{Pipeline: pipeline}
			request := pipe.ToRequest()
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

func (cs CleanupService) _getKeycloakUserById(id string) (user *gocloak.User) {
	user, err := cs.keycloak.GetUserByID(id)
	if err != nil {
		log.Fatal("GetUserByID failed:" + err.Error())
	}
	if user == nil {
		log.Fatal("could not get user data from keycloak")
	}
	return
}

func (cs CleanupService) _logPrint(vars ...string) {
	cs.logger.Print(DividerString)
	for _, v := range vars {
		cs.logger.Print(v)
	}
}
