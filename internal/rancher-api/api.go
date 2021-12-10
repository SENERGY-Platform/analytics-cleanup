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

package rancher_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/analytics-cleanup/internal/lib"
	"net/http"

	"github.com/parnurzeal/gorequest"
)

type Rancher struct {
	url              string
	accessKey        string
	secretKey        string
	pipelinesStackId string
	servingStackId   string
}

func NewRancher(url string, accessKey string, secretKey string, pipelinesStackId string, servingStackId string) *Rancher {
	return &Rancher{url, accessKey, secretKey, pipelinesStackId, servingStackId}
}

func (r Rancher) CreateServingInstance(instance *lib.ServingInstance, dataFields string, tagFields string) (serviceId string) {
	env := map[string]string{
		"KAFKA_GROUP_ID":      "transfer-" + instance.ApplicationId.String(),
		"KAFKA_BOOTSTRAP":     lib.GetEnv("KAFKA_BOOTSTRAP", "broker.kafka.rancher.internal:9092"),
		"KAFKA_TOPIC":         instance.Topic,
		"DATA_MEASUREMENT":    instance.Measurement,
		"DATA_FIELDS_MAPPING": dataFields,
		"DATA_TAGS_MAPPING":   tagFields,
		"DATA_TIME_MAPPING":   instance.TimePath,
		"DATA_FILTER_ID":      instance.Filter,
		"INFLUX_DB":           instance.Database,
		"INFLUX_HOST":         lib.GetEnv("INFLUX_DB_HOST", "influxdb"),
		"INFLUX_PORT":         lib.GetEnv("INFLUX_DB_PORT", "8086"),
		"INFLUX_USER":         lib.GetEnv("INFLUX_DB_USER", "root"),
		"INFLUX_PW":           lib.GetEnv("INFLUX_DB_PASSWORD", ""),
		"OFFSET_RESET":        instance.Offset,
	}

	if instance.TimePrecision != nil && *instance.TimePrecision != "" {
		env["TIME_PRECISION"] = *instance.TimePrecision
	}

	if instance.FilterType == "operatorId" {
		env["DATA_FILTER_ID_MAPPING"] = "operator_id"
	}

	if instance.FilterType == "import_id" {
		env["DATA_FILTER_ID_MAPPING"] = "import_id"
	}

	labels := map[string]string{
		"io.rancher.container.pull_image":          "always",
		"io.rancher.scheduler.affinity:host_label": "role=worker",
	}
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey)

	reqBody := &Request{
		Type:          "service",
		Name:          "kafka-influx-" + instance.ID.String(),
		StackId:       r.servingStackId,
		Scale:         1,
		StartOnCreate: true,
		LaunchConfig: LaunchConfig{
			ImageUuid: "docker:" + lib.GetEnv("TRANSFER_IMAGE",
				"fgseitsrancher.wifa.intern.uni-leipzig.de:5000/kafka-influx:unstable"),
			Environment: env,
			Labels:      labels,
		},
	}
	resp, body, e := request.Post(r.url + "services").Send(reqBody).End()
	if resp.StatusCode != http.StatusCreated {
		fmt.Println("Something went totally wrong", resp)
	}
	if len(e) > 0 {
		fmt.Println("Something went wrong", e)
	}
	serviceId = lib.ToJson(body)["id"].(string)
	return
}

func (r Rancher) GetServices(collection string) (services []lib.Service, err error) {
	return
}

func (r Rancher) GetWorkloads(collection string) (services []lib.Workload, err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey)
	req := request.Get(r.url + "stacks/" + r.pipelinesStackId + "/services")
	if collection == "serving" {
		req = request.Get(r.url + "stacks/" + r.servingStackId + "/services")
	}
	resp, body, errs := req.End()
	if len(errs) > 0 || resp.StatusCode != http.StatusOK {
		err = errors.New("could not access services")
		return
	}
	var serviceCollection = ServiceCollection{}
	err = json.Unmarshal([]byte(body), &serviceCollection)
	if len(serviceCollection.Data) > 0 {
		for _, service := range serviceCollection.Data {
			services = append(services, lib.Workload{
				Id:          service.Id,
				Name:        service.Name,
				ImageUuid:   service.ImageUuid,
				Environment: service.Environment,
				Labels:      service.Labels,
			})
		}
	}
	return
}

func (r Rancher) GetWorkloadEnvs(collection string) (envs []map[string]string, err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey)
	req := request.Get(r.url + "stacks/" + r.pipelinesStackId + "/services")
	if collection == "serving" {
		req = request.Get(r.url + "stacks/" + r.servingStackId + "/services")
	}
	resp, body, errs := req.End()
	if len(errs) > 0 || resp.StatusCode != http.StatusOK {
		err = errors.New("could not access services")
		return
	}
	var serviceCollection = ServiceCollection{}
	err = json.Unmarshal([]byte(body), &serviceCollection)
	if len(serviceCollection.Data) > 0 {
		for _, service := range serviceCollection.Data {
			envs = append(envs, service.Environment)
		}
	}
	return
}

func (r Rancher) DeleteWorkload(serviceName string, collection string) (err error) {
	service, err := r.getServiceByName(serviceName)
	if err != nil {
		return
	}
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey)
	resp, body, _ := request.Delete(r.url + "services/" + service.Id).End()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("could not delete operator: " + body)
	}
	return
}

func (r Rancher) DeleteService(serviceName string, collection string) (err error) {
	return
}

func (r Rancher) getServiceByName(name string) (service Service, err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey)
	resp, body, errs := request.Get(r.url + "services/?name=" + name).End()
	if len(errs) > 0 || resp.StatusCode != http.StatusOK {
		err = errors.New("could not access service name")
		return
	}
	var serviceCollection = ServiceCollection{}
	err = json.Unmarshal([]byte(body), &serviceCollection)
	if len(serviceCollection.Data) < 1 {
		err = errors.New("could not find corresponding service")
		return
	}
	service = serviceCollection.Data[0]
	return
}
