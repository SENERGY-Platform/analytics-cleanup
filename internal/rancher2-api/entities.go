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

type Request struct {
	Name        string            `json:"name,omitempty"`
	NamespaceId string            `json:"namespaceId,omitempty"`
	Containers  []Container       `json:"containers,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Selector    Selector          `json:"selector,omitempty"`
	Scheduling  Scheduling        `json:"scheduling,omitempty"`
}

type Container struct {
	Image           string    `json:"image,omitempty"`
	Name            string    `json:"name,omitempty"`
	Env             []Env     `json:"env,omitempty"`
	ImagePullPolicy string    `json:"imagePullPolicy,omitempty"`
	Command         []string  `json:"command,omitempty"`
	Resources       Resources `json:"resources,omitempty"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Resources struct {
	Limits Limits `json:"limits,omitempty"`
}

type Limits struct {
	Cpu string `json:"cpu,omitempty"`
}

type Selector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type Scheduling struct {
	Node      Node   `json:"node,omitempty"`
	Scheduler string `json:"scheduler,omitempty"`
}

type Node struct {
	RequireAll []string `json:"requireAll,omitempty"`
}

type WorkloadCollection struct {
	Data []Workload `json:"data"`
}

type Workload struct {
	Id         string            `json:"id,omitempty"`
	Name       string            `json:"name,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Containers []Container       `json:"containers,omitempty"`
}

type ServiceCollection struct {
	Data []Service `json:"data"`
}

type Service struct {
	Id                string   `json:"id,omitempty"`
	Name              string   `json:"name,omitempty"`
	BaseType          string   `json:"baseType,omitempty"`
	TargetWorkloadIds []string `json:"targetWorkloadIds,omitempty"`
}
