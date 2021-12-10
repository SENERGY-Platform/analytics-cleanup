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

type Driver interface {
	GetServices(collection string) (services []Service, err error)
	GetWorkloads(collection string) (workloads []Workload, err error)
	GetWorkloadEnvs(collection string) (envs []map[string]string, err error)
	DeleteWorkload(id string, collection string) error
	DeleteService(id string, collection string) error
	CreateServingInstance(instance *ServingInstance, dataFields string, tagFields string) string
}
