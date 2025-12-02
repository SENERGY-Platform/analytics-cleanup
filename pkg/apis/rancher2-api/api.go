/*
 * Copyright 2019 InfAI (CC SES)
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

package rancher2_api

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SENERGY-Platform/analytics-cleanup/lib"

	"github.com/parnurzeal/gorequest"
)

type Rancher2 struct {
	url                string
	accessKey          string
	secretKey          string
	servingNamespaceId string
	servingProjectId   string
	pipeNamespaceId    string
	pipeProjectId      string
}

func NewRancher2(url string, accessKey string, secretKey string, servingNamespaceId string, servingProjectId string, pipeNamespaceId string, pipeProjectId string) *Rancher2 {
	return &Rancher2{url, accessKey, secretKey, servingNamespaceId, servingProjectId, pipeNamespaceId, pipeProjectId}
}

func (r *Rancher2) GetServices(collection string) (services []lib.KubeService, err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey).TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	request.Get(r.url + "projects/" + r.pipeProjectId + "/services/?limit=2000&namespaceId=" + r.pipeNamespaceId)
	if collection == "serving" {
		request.Get(r.url + "projects/" + r.servingProjectId + "/services/?limit=2000&namespaceId=" + r.servingNamespaceId)
	}
	resp, body, e := request.End()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("could not get services: ")
		return
	}
	if len(e) > 0 {
		err = errors.New("something went wrong")
		return
	}
	var serviceCollection = ServiceCollection{}
	err = json.Unmarshal([]byte(body), &serviceCollection)
	if len(serviceCollection.Data) > 1 {
		for _, service := range serviceCollection.Data {
			services = append(services, lib.KubeService{
				Id:                service.Id,
				Name:              service.Name,
				BaseType:          service.BaseType,
				TargetWorkloadIds: service.TargetWorkloadIds,
			})
		}
	}
	return
}

func (r *Rancher2) GetWorkloads(collection string) (workloads []lib.Workload, err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey).TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	request.Get(r.url + "projects/" + r.pipeProjectId + "/workloads/?namespaceId=" + r.pipeNamespaceId)
	if collection == "serving" {
		request.Get(r.url + "projects/" + r.servingProjectId + "/workloads/?namespaceId=" + r.servingNamespaceId)
	}
	resp, body, e := request.End()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("could not get workloads: ")
		return
	}
	if len(e) > 0 {
		err = errors.New("something went wrong")
		return
	}

	var workloadCollection = WorkloadCollection{}
	err = json.Unmarshal([]byte(body), &workloadCollection)
	if len(workloadCollection.Data) > 0 {
		for _, workload := range workloadCollection.Data {
			r2Env := map[string]string{}
			for _, v := range workload.Containers[0].Env {
				r2Env[v.Name] = v.Value
			}
			workloads = append(workloads, lib.Workload{
				Id:          workload.Id,
				Name:        workload.Name,
				ImageUuid:   workload.Containers[0].Image,
				Environment: r2Env,
				Labels:      workload.Labels,
			})
		}
	}
	return
}

func (r *Rancher2) GetWorkloadEnvs(collection string) (envs []map[string]string, err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey).TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	request.Get(r.url + "projects/" + r.pipeProjectId + "/workloads/?namespaceId=" + r.pipeNamespaceId)
	if collection == "serving" {
		request.Get(r.url + "projects/" + r.servingProjectId + "/workloads/?namespaceId=" + r.servingNamespaceId)
	}
	resp, body, e := request.End()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("could not get services: ")
		return
	}
	if len(e) > 0 {
		err = errors.New("something went wrong")
		return
	}
	var workloadCollection = WorkloadCollection{}
	err = json.Unmarshal([]byte(body), &workloadCollection)
	if len(workloadCollection.Data) > 1 {
		for _, workload := range workloadCollection.Data {
			for _, container := range workload.Containers {
				r2Env := map[string]string{}
				for _, v := range container.Env {
					r2Env[v.Name] = v.Value
				}
				envs = append(envs, r2Env)
			}
		}
	}
	return
}

func (r *Rancher2) DeleteWorkload(workloadId string, collection string) (err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey).TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	request.Delete(r.url + "projects/" + r.pipeProjectId + "/workloads/deployment:" + r.pipeNamespaceId + ":" + workloadId)
	if collection == "serving" {
		request.Delete(r.url + "projects/" + r.servingProjectId + "/workloads/deployment:" + r.servingNamespaceId + ":" + workloadId)
	}
	resp, body, e := request.End()
	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			return lib.NewNotFoundError(errors.New(body))
		}
		err = errors.New("could not delete operator: " + body)
		return
	}
	if len(e) > 0 {
		err = errors.New("something went wrong")
		return
	}
	return
}

func (r *Rancher2) DeleteService(serviceId string, collection string) (err error) {
	request := gorequest.New().SetBasicAuth(r.accessKey, r.secretKey).TLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	request.Delete(r.url + "project/" + r.pipeProjectId + "/services/" + serviceId)
	if collection == "serving" {
		request.Delete(r.url + "project/" + r.servingProjectId + "/services/" + serviceId)
	}
	resp, body, e := request.End()
	if resp.StatusCode != http.StatusNoContent {
		err = errors.New("could not delete service: " + body)
		return
	}
	if len(e) > 0 {
		err = errors.New("something went wrong")
		return
	}
	return
}

func (r *Rancher2) getServingInstanceName(name string) string {
	return "kafka2influx-" + name
}
