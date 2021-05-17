/*
 * Copyright 2018 InfAI (CC SES)
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
	uuid "github.com/satori/go.uuid"
	"time"
)

type Pipeline struct {
	Id          string    `bson:"id" json:"id"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	FlowId      string    `json:"flowId,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
	UserId      string
	Operators   []Operator `json:"operators,omitempty"`
}

type Operator struct {
	Id             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	ImageId        string `json:"imageId,omitempty"`
	DeploymentType string `json:"deploymentType,omitempty"`
	OperatorId     string `json:"operatorId,omitempty"`
}

type ServingInstance struct {
	ID               uuid.UUID              `json:"ID,omitempty"`
	Name             string                 `json:"Name,omitempty"`
	Description      string                 `json:"Description,omitempty"`
	EntityName       string                 `json:"EntityName,omitempty"`
	ServiceName      string                 `json:"ServiceName,omitempty"`
	Topic            string                 `json:"Topic,omitempty"`
	Database         string                 `json:"Database,omitempty"`
	Measurement      string                 `json:"Measurement,omitempty"`
	Filter           string                 `json:"Filter,omitempty"`
	FilterType       string                 `json:"FilterType,omitempty"`
	TimePath         string                 `json:"TimePath,omitempty"`
	UserId           string                 `json:"UserId,omitempty"`
	RancherServiceId string                 `json:"RancherServiceId,omitempty"`
	Offset           string                 `json:"Offset,omitempty"`
	Values           []ServingInstanceValue `json:"Values,omitempty"`
	CreatedAt        time.Time              `json:"CreatedAt,omitempty"`
	UpdatedAt        time.Time              `json:"UpdatedAt,omitempty"`
}

type ServingInstanceValue struct {
	InstanceID uuid.UUID `json:"InstanceID,omitempty"`
	Name       string    `json:"Name,omitempty"`
	Type       string    `json:"Type,omitempty"`
	Path       string    `json:"Path,omitempty"`
}

type Workload struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	ImageUuid   string            `json:"imageUuid,omitempty"`
	Environment map[string]string `json:"environment"`
	Labels      map[string]string `json:"labels"`
}

type Service struct {
	Id       string `json:"id"`
	BaseType string `json:"baseType"`
	Name     string `json:"name"`
}
