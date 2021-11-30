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
	"log"
	"strings"
	"time"
)

type Pipeline struct {
	Id                 string    `bson:"id" json:"id"`
	Name               string    `json:"name,omitempty"`
	Description        string    `json:"description,omitempty"`
	FlowId             string    `json:"flowId,omitempty"`
	Image              string    `json:"image,omitempty"`
	WindowTime         *int      `json:"windowTime,omitempty"` //missing if not set explicitly, defaults to 30
	ConsumeAllMessages bool      `json:"consumeAllMessages,omitempty"`
	Metrics            bool      `json:"metrics,omitempty"`
	CreatedAt          time.Time `json:"createdAt,omitempty"`
	UpdatedAt          time.Time `json:"updatedAt,omitempty"`
	UserId             string
	Operators          []Operator `json:"operators,omitempty"`
}

func (p *Pipeline) ToRequest() *PipelineRequest {
	r := PipelineRequest{
		Id:                 p.Id,
		FlowId:             p.FlowId,
		Name:               p.Name,
		Description:        p.Description,
		WindowTime:         30, // default
		ConsumeAllMessages: p.ConsumeAllMessages,
		Metrics:            p.Metrics,
		Nodes:              []PipelineNode{},
	}
	if p.WindowTime != nil {
		r.WindowTime = *p.WindowTime
	}
	for _, operator := range p.Operators {
		node := PipelineNode{
			NodeId:          operator.Id,
			Inputs:          []NodeInput{},
			Config:          []NodeConfig{},
			InputSelections: operator.InputSelections,
		}
		for _, inputTopic := range operator.InputTopics {
			nodeInput := NodeInput{
				FilterType: inputTopic.FilterType,
				FilterIds:  inputTopic.FilterValue,
				TopicName:  inputTopic.Name,
				Values:     []NodeValue{},
			}
			if nodeInput.FilterType == "DeviceId" {
				nodeInput.FilterType = "deviceId"
			}
			if nodeInput.FilterType == "OperatorId" {
				nodeInput.FilterType = "operatorId"
			}
			if nodeInput.FilterType == "operatorId" && len(strings.Split(nodeInput.FilterIds, ":")) == 1 {
				log.Println("Pipeline uses legacy operatorId filter, fixing now...")
				nodeInput.FilterIds += ":" + r.Id
			}
			for _, mapping := range inputTopic.Mappings {
				nodeInput.Values = append(nodeInput.Values, NodeValue{
					Name: mapping.Dest,
					Path: mapping.Source,
				})
			}
			node.Inputs = append(node.Inputs, nodeInput)
		}
		for key, value := range operator.Config {
			node.Config = append(node.Config, NodeConfig{
				Name:  key,
				Value: value,
			})
		}
		r.Nodes = append(r.Nodes, node)
	}
	return &r
}

type PipelineRequest struct {
	Id                 string         `json:"id,omitempty"`
	FlowId             string         `json:"flowId,omitempty"`
	Name               string         `json:"name,omitempty"`
	Description        string         `json:"description,omitempty"`
	WindowTime         int            `json:"windowTime,omitempty"`
	ConsumeAllMessages bool           `json:"consumeAllMessages,omitempty"`
	Metrics            bool           `json:"metrics,omitempty"`
	Nodes              []PipelineNode `json:"nodes,omitempty"`
}

type PipelineNode struct {
	NodeId          string           `json:"nodeId,omitempty"`
	Inputs          []NodeInput      `json:"inputs,omitempty"`
	Config          []NodeConfig     `json:"config,omitempty"`
	InputSelections []InputSelection `json:"inputSelections,omitempty"`
}

type NodeConfig struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type InputSelection struct {
	InputName         string   `json:"inputName,omitempty"`
	AspectId          string   `json:"aspectId,omitempty"`
	FunctionId        string   `json:"functionId,omitempty"`
	CharacteristicIds []string `json:"characteristicIds,omitempty"`
	SelectableId      string   `json:"selectableId,omitempty"`
}

type NodeInput struct {
	FilterType string      `json:"filterType,omitempty"`
	FilterIds  string      `json:"filterIds,omitempty"`
	TopicName  string      `json:"topicName,omitempty"`
	Values     []NodeValue `json:"values,omitempty"`
}

type NodeValue struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type Operator struct {
	Id              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	ApplicationId   uuid.UUID         `json:"applicationId,omitempty"`
	ImageId         string            `json:"imageId,omitempty"`
	DeploymentType  string            `json:"deploymentType,omitempty"`
	OperatorId      string            `json:"operatorId,omitempty"`
	Config          map[string]string `json:"config,omitempty"`
	OutputTopic     string            `json:"outputTopic,omitempty"`
	InputTopics     []InputTopic      `json:"inputTopics,omitempty"`
	InputSelections []InputSelection  `json:"inputSelections,omitempty"`
}

type InputTopic struct {
	Name        string    `json:"name,omitempty"`
	FilterType  string    `json:"filterType,omitempty"`
	FilterValue string    `json:"filterValue,omitempty"`
	Mappings    []Mapping `json:"mappings,omitempty"`
}

type Mapping struct {
	Dest   string `json:"dest,omitempty"`
	Source string `json:"source,omitempty"`
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
	Id                string   `json:"id"`
	BaseType          string   `json:"baseType"`
	Name              string   `json:"name"`
	TargetWorkloadIds []string `json:"targetWorkloadIds,omitempty"`
}

type OpenidToken struct {
	AccessToken      string    `json:"access_token"`
	ExpiresIn        float64   `json:"expires_in"`
	RefreshExpiresIn float64   `json:"refresh_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	RequestTime      time.Time `json:"-"`
}
