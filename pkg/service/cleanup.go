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
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/SENERGY-Platform/analytics-cleanup/lib"
	"github.com/SENERGY-Platform/analytics-cleanup/pkg/apis"
	"github.com/SENERGY-Platform/analytics-cleanup/pkg/util"

	pipeModels "github.com/SENERGY-Platform/analytics-pipeline/lib"
)

type CleanupService struct {
	keycloak      apis.KeycloakService
	driver        Driver
	pipeline      apis.PipelineService
	logger        util.FileLogger
	kafkaAdmin    *apis.KafkaAdmin
	ctx           context.Context
	deleteCancel  context.CancelFunc
	deleteRunning bool
	deleteStatus  lib.DeleteStatus
	deleteMu      sync.Mutex
}

const DividerString = "++++++++++++++++++++++++++++++++++++++++++++++++++++++++"

func NewCleanupService(keycloak apis.KeycloakService, driver Driver, pipeline apis.PipelineService, logger util.FileLogger, kafkaAdmin *apis.KafkaAdmin, ctx context.Context) *CleanupService {
	return &CleanupService{
		keycloak:   keycloak,
		driver:     driver,
		pipeline:   pipeline,
		logger:     logger,
		kafkaAdmin: kafkaAdmin,
		ctx:        ctx,
	}
}

func (cs *CleanupService) StartCleanupService(recreatePipes bool) (err error) {
	/****************************
	Check analytics pipelines
	****************************/
	if recreatePipes {
		var info *gocloak.UserInfo
		info, err = cs.keycloak.GetUserInfo()
		if err != nil {
			return
		}

		var pipes []pipeModels.Pipeline
		pipes, err = cs.pipeline.GetPipelines(*info.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			return
		}

		workloads, _ := cs.driver.GetWorkloads(lib.PIPELINE)

		err = cs.recreatePipelines(pipes, workloads)
		if err != nil {
			log.Fatal("recreatePipelines failed: " + err.Error())
		}
	}
	return
	/*
		cs.deleteOrphanedPipelineServices()
		cs.deleteOrphanedAnalyticsWorkloads()
		cs.deleteOrphanedKubeServices(PIPELINE)
		cs.deleteOrphanedKafkaTopics()
	*/
}

func (cs *CleanupService) GetOrphanedPipelineServices(userId string, authToken string) (orphanedPipelineWorkloads []pipeModels.Pipeline, err error) {
	var pipes []pipeModels.Pipeline
	pipes, err = cs.pipeline.GetPipelines(userId, authToken)
	if err != nil {
		return
	}
	workloads, err := cs.driver.GetWorkloads(lib.PIPELINE)
	if err != nil {
		return
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
	return
}

func (cs *CleanupService) DeleteOrphanedPipelineService(id string, accessToken string) error {
	return cs.pipeline.DeletePipeline(id, accessToken)
}

func (cs *CleanupService) DeleteOrphanedPipelineServices(userId string, authToken string) (pipes []pipeModels.Pipeline, err error) {
	pipes, err = cs.GetOrphanedPipelineServices(userId, authToken)
	if err != nil {
		for _, pipe := range pipes {
			err = cs.pipeline.DeletePipeline(pipe.Id, cs.keycloak.GetAccessToken())
			if err != nil {
				return
			}
		}
	}
	return
}

func (cs *CleanupService) GetOrphanedAnalyticsWorkloads(userId string, authToken string) (orphanedAnalyticsWorkloads []lib.Workload, err error) {
	pipes, err := cs.pipeline.GetPipelines(userId, authToken)
	if err != nil {
		return
	}
	workloads, err := cs.driver.GetWorkloads(lib.PIPELINE)
	if err != nil {
		return
	}
	for _, workload := range workloads {
		if !workloadInPipes(workload, pipes) {
			orphanedAnalyticsWorkloads = append(orphanedAnalyticsWorkloads, workload)
		}
	}
	return
}

func (cs *CleanupService) DeleteOrphanedAnalyticsWorkload(name string) error {
	return cs.driver.DeleteWorkload(name, lib.PIPELINE)
}

func (cs *CleanupService) DeleteOrphanedAnalyticsWorkloads(userId string, authToken string) (workloads []lib.Workload, err error) {
	workloads, err = cs.GetOrphanedAnalyticsWorkloads(userId, authToken)
	if err != nil {
		for _, workload := range workloads {
			err = cs.driver.DeleteWorkload(workload.Name, lib.PIPELINE)
			if err != nil {
				return
			}
		}
	}
	return
}

func (cs *CleanupService) GetOrphanedKafkaTopics() (orphanedKafkaTopics []string, err error) {
	envs, err := cs.driver.GetWorkloadEnvs(lib.PIPELINE)
	if err != nil {
		return
	}
	topics, err := cs.kafkaAdmin.GetTopics()
	if err != nil {
		return
	}
	for _, topic := range topics {
		if isInternalAnalyticsTopic(topic) && !pipelineExists(topic, envs) {
			orphanedKafkaTopics = append(orphanedKafkaTopics, topic)
		}
	}
	return
}

func (cs *CleanupService) DeleteOrphanedKafkaTopic(topic string) error {
	return cs.kafkaAdmin.DeleteTopic(topic)
}

func (cs *CleanupService) DeleteOrphanedKafkaTopics() (err error) {
	cs.deleteMu.Lock()
	if cs.deleteRunning {
		cs.deleteMu.Unlock()
		return lib.NewConflictError(errors.New("delete task already running"))
	}
	cs.deleteRunning = true
	ctx, cancelDelete := context.WithCancel(cs.ctx)
	cs.deleteCancel = cancelDelete
	topics, err := cs.GetOrphanedKafkaTopics()
	if err != nil {
		return
	}
	cs.deleteStatus = lib.DeleteStatus{
		Total:     len(topics),
		Remaining: len(topics),
		Running:   true,
		Errors:    nil,
	}
	cs.deleteMu.Unlock()

	go func(ctx context.Context) {
		for index, topic := range topics {
			var errs []error
			select {
			case <-ctx.Done():
				cs.deleteMu.Lock()
				cs.deleteStatus.Running = false
				cs.deleteStatus.Errors = append(cs.deleteStatus.Errors, errors.New("aborted"))
				cs.deleteRunning = false
				util.Logger.Info("aborted delete kafka topics")
				cs.deleteMu.Unlock()
				return
			default:
				err = cs.kafkaAdmin.DeleteTopic(topic)
				if err != nil {
					errs = append(errs, err)
				} else {
					util.Logger.Info("deleted orphaned kafka topic", "topic", topic)
				}
				cs.deleteMu.Lock()
				cs.deleteStatus.Remaining = len(topics) - (index + 1)
				cs.deleteStatus.Errors = errs
				cs.deleteMu.Unlock()
				// give kafka time to breath
				time.Sleep(1 * time.Second)
			}
		}
		cs.deleteMu.Lock()
		cs.deleteStatus.Running = false
		cs.deleteRunning = false
		cs.deleteMu.Unlock()
	}(ctx)
	return
}

func (cs *CleanupService) GetDeleteOrphanedKafkaTopicsStatus() lib.DeleteStatus {
	cs.deleteMu.Lock()
	defer cs.deleteMu.Unlock()
	return cs.deleteStatus
}

func (cs *CleanupService) StopDeleteOrphanedKafkaTopics() (err error) {
	cs.deleteMu.Lock()
	cancel := cs.deleteCancel
	cs.deleteMu.Unlock()

	if cancel != nil {
		cancel()
		util.Logger.Debug("stopped delete orphaned kafka topics")
	} else {
		err = lib.NewConflictError(errors.New("deleteOrphanedKafkaTopics not running"))
	}
	return
}

func (cs *CleanupService) GetOrphanedKubeServices(collection string) (orphanedServices []lib.KubeService, err error) {
	workloads, err := cs.driver.GetWorkloads(collection)
	if err != nil {
		return
	}
	services, err := cs.driver.GetServices(collection)
	if err != nil {
		return
	}
	for _, service := range services {
		if !serviceInWorkloads(service, workloads) {
			orphanedServices = append(orphanedServices, service)
		}
	}
	return
}

func (cs *CleanupService) DeleteOrphanedKubeService(collection string, id string) error {
	return cs.driver.DeleteService(id, collection)
}

func (cs *CleanupService) DeleteOrphanedKubeServices(collection string) (deletedKubeServices []lib.KubeService, err error) {
	services, err := cs.GetOrphanedKubeServices(collection)
	if err != nil {
		return
	}
	for _, service := range services {
		err = cs.driver.DeleteService(service.Id, collection)
		if err != nil {
			return
		}
		deletedKubeServices = append(deletedKubeServices, service)
	}
	return
}

func (cs *CleanupService) recreatePipelines(pipelines []pipeModels.Pipeline, workloads []lib.Workload) error {
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
