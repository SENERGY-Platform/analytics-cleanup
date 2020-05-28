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
	"log"
	"strings"
)

type CleanupService struct {
	keycloak KeycloakService
	driver   Driver
	pipeline PipelineService
	serving  ServingService
	logger   Logger
}

func NewCleanupService(keycloak KeycloakService, driver Driver, pipeline PipelineService, serving ServingService, logger Logger) *CleanupService {
	return &CleanupService{keycloak: keycloak, driver: driver, pipeline: pipeline, serving: serving, logger: logger}
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
		services, err := cs.driver.GetServices("pipelines")
		if err != nil {
			log.Fatal("GetServices failed: " + err.Error())
		}
		cs.checkPipelineServices(pipes, services)
		cs.checkPipes(services, pipes)

		/****************************
			Check analytics serving
		****************************/

		servings, err := cs.serving.GetServingServices(*user.Sub, cs.keycloak.GetAccessToken())
		if err != nil {
			log.Fatal("GetServingServices failed: " + err.Error())
		}
		services, err = cs.driver.GetServices("serving")
		if err != nil {
			log.Fatal("GetServices failed: " + err.Error())
		}
		cs.checkServingServices(servings, services)
		cs.checkServings(services, servings)
	}
}

func (cs CleanupService) checkPipelineServices(pipes []Pipeline, services []Service) {
	cs.logger.Print("********************************************************")
	cs.logger.Print("**************** Orphaned Pipelines ********************")
	cs.logger.Print("********************************************************")
	for _, pipe := range pipes {
		if !pipeInServices(pipe, services) {
			cs.logger.Print("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			cs.logger.Print(pipe.Id)
			cs.logger.Print(pipe.Name)
			for _, operator := range pipe.Operators {
				cs.logger.Print(operator.Name)
				cs.logger.Print(operator.ImageId)
			}
			cs.logger.Print(len(pipe.Operators))
			user, err := cs.keycloak.GetUserByID(pipe.UserId)
			if err != nil {
				log.Fatal("GetUserByID failed:" + err.Error())
			}
			if user != nil {
				cs.logger.Print(*user.Username)
			}
			err = cs.pipeline.DeletePipeline(pipe.Id, pipe.UserId, cs.keycloak.GetAccessToken())
			if err != nil {
				log.Fatal("DeletePipeline failed:" + err.Error())
			}
		}

	}
}

func (cs CleanupService) checkPipes(services []Service, pipes []Pipeline) {
	cs.logger.Print("********************************************************")
	cs.logger.Print("************** Orphaned Pipeline Services **************")
	cs.logger.Print("********************************************************")
	for _, service := range services {
		if !serviceInPipes(service, pipes) {
			cs.logger.Print("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			cs.logger.Print(service.Name)
			cs.logger.Print(service.Id)
			cs.logger.Print(service.ImageUuid)
			cs.logger.Print(service.Labels)
			err := cs.driver.DeleteService(service.Name, "")
			if err != nil {
				log.Fatal("DeleteService failed: " + err.Error())
			}
		}
	}
}

func (cs CleanupService) checkServingServices(serving []ServingInstance, services []Service) {
	cs.logger.Print("********************************************************")
	cs.logger.Print("**************** Orphaned Servings ********************")
	cs.logger.Print("********************************************************")
	for _, serving := range serving {
		if !servingInServices(serving, services) {
			cs.logger.Print("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			cs.logger.Print(serving.ID)
			cs.logger.Print(serving.Name)
			user, err := cs.keycloak.GetUserByID(serving.UserId)
			if err != nil {
				log.Fatal("GetUserByID failed:" + err.Error())
			}
			if user != nil {
				cs.logger.Print(*user.Username)
			}
			//cs.create(serving)
			err = cs.serving.DeleteServingService(serving.ID.String(), serving.UserId, cs.keycloak.GetAccessToken())
			if err != nil {
				log.Fatal("DeleteServingService failed:" + err.Error())
			}
		}

	}
}

func (cs CleanupService) checkServings(services []Service, servings []ServingInstance) {
	cs.logger.Print("********************************************************")
	cs.logger.Print("************** Orphaned Serving Services ***************")
	cs.logger.Print("********************************************************")
	for _, service := range services {
		if !stringInSlice(service.Name, []string{"influxdb", "influx-auth", "serving-lb", "grafana"}) {
			if !serviceInServings(service, servings) {
				cs.logger.Print("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
				cs.logger.Print(service.Name)
				cs.logger.Print(service.Id)
				cs.logger.Print(service.ImageUuid)
				cs.logger.Print(service.Labels)
				err := cs.driver.DeleteService(service.Name, "serving")
				if err != nil {
					log.Fatal("DeleteService failed: " + err.Error())
				}
			}
		}
	}
}

func pipeInServices(pipe Pipeline, services []Service) bool {
	for _, service := range services {
		if strings.Contains(service.Name, pipe.Id) {
			return true
		}
	}
	return false
}

func serviceInPipes(service Service, pipes []Pipeline) bool {
	for _, pipe := range pipes {
		if strings.Contains(service.Name, pipe.Id) {
			return true
		}
	}
	return false
}

func servingInServices(serving ServingInstance, services []Service) bool {
	for _, service := range services {
		if strings.Contains(service.Name, serving.ID.String()) {
			return true
		}
	}
	return false
}

func serviceInServings(service Service, servings []ServingInstance) bool {
	for _, serving := range servings {
		if strings.Contains(service.Name, serving.ID.String()) {
			return true
		}
	}
	return false
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
