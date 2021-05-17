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
	"strings"
)

type CleanupService struct {
	keycloak KeycloakService
	driver   Driver
	pipeline PipelineService
	serving  ServingService
	logger   Logger
	influx   *Influx
}

const DividerString = "++++++++++++++++++++++++++++++++++++++++++++++++++++++++"
const HeaderString = "********************************************************"

func NewCleanupService(keycloak KeycloakService, driver Driver, pipeline PipelineService, serving ServingService, logger Logger) *CleanupService {
	influx := NewInflux()
	return &CleanupService{keycloak: keycloak, driver: driver, pipeline: pipeline, serving: serving, logger: logger, influx: influx}
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
			log.Fatal("GetServices for pipelines failed: " + err.Error())
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
			log.Fatal("GetServices for serving instances failed: " + err.Error())
		}

		influxData, err := cs.getInfluxData()
		if err != nil {
			log.Fatal("Get Influx data for serving instances failed: " + err.Error())
		}
		//cs.recreateServingServices(servings, services)
		cs.checkServingServices(servings, services)
		cs.checkServings(services, servings)
		cs.checkInfluxMeasurements(influxData, servings)
	}
}

func (cs CleanupService) checkInfluxMeasurements(influxData map[string][]string, servings []ServingInstance) {
	cs.logger.Print(HeaderString)
	cs.logger.Print("**************** Orphaned Measurements *****************")
	cs.logger.Print(HeaderString)
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

func (cs CleanupService) checkPipelineServices(pipes []Pipeline, services []Service) {
	cs.logger.Print(HeaderString)
	cs.logger.Print("**************** Orphaned Pipelines ********************")
	cs.logger.Print(HeaderString)
	for _, pipe := range pipes {
		if !pipeInServices(pipe, services) {
			deletePipe := true

			for _, operator := range pipe.Operators {
				if operator.DeploymentType != "local" {
					cs.logger.Print(operator.Name)
					cs.logger.Print(operator.ImageId)
				} else {
					deletePipe = false
					break
				}
			}

			if deletePipe {
				cs.logger.Print(DividerString)
				cs.logger.Print(pipe.Id)
				cs.logger.Print(pipe.Name)
				cs.logger.Print(len(pipe.Operators))
				user := cs.getKeycloakUserById(pipe.UserId)
				cs.logger.Print(*user.Username)
				err := cs.pipeline.DeletePipeline(pipe.Id, pipe.UserId, cs.keycloak.GetAccessToken())
				if err != nil {
					log.Fatal("DeletePipeline failed:" + err.Error())
				}
			}
		}

	}
}

func (cs CleanupService) checkPipes(services []Service, pipes []Pipeline) {
	cs.logger.Print(HeaderString)
	cs.logger.Print("************** Orphaned Pipeline Services **************")
	cs.logger.Print(HeaderString)
	for _, service := range services {
		if !serviceInPipes(service, pipes) {
			cs.logger.Print(DividerString)
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
	cs.logger.Print(HeaderString)
	cs.logger.Print("**************** Orphaned Servings *********************")
	cs.logger.Print(HeaderString)
	for _, serving := range serving {
		if !servingInServices(serving, services) {
			cs.logger.Print(DividerString)
			cs.logger.Print(serving.ID)
			cs.logger.Print(serving.Name)
			user := cs.getKeycloakUserById(serving.UserId)
			cs.logger.Print(*user.Username)
			err := cs.serving.DeleteServingService(serving.ID.String(), serving.UserId, cs.keycloak.GetAccessToken())
			if err != nil {
				log.Fatal("DeleteServingService failed:" + err.Error())
			}
		}

	}
}

func (cs CleanupService) checkServings(services []Service, servings []ServingInstance) {
	cs.logger.Print(HeaderString)
	cs.logger.Print("************** Orphaned Serving Services ***************")
	cs.logger.Print(HeaderString)
	for _, service := range services {
		if strings.Contains(service.Name, "kafka-influx") || strings.Contains(service.Name, "kafka2influx") {
			if !serviceInServings(service, servings) {
				cs.logger.Print(DividerString)
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

func influxMeasurementInServings(measurement string, servings []ServingInstance) bool {
	for _, serving := range servings {
		if measurement == serving.Measurement {
			return true
		}
	}
	return false
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
	for db, _ := range influxDbs {
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

func (cs CleanupService) recreateServingServices(serving []ServingInstance, services []Service) {
	cs.logger.Print(HeaderString)
	cs.logger.Print("**************** Recreate Servings *********************")
	cs.logger.Print(HeaderString)
	for _, serving := range serving {
		if !servingInServices(serving, services) {
			cs.logger.Print(DividerString)
			cs.logger.Print(serving.ID)
			cs.logger.Print(serving.Name)

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
