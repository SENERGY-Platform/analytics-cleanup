/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
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
	"time"

	engineModels "github.com/SENERGY-Platform/analytics-flow-engine/lib"
	pipeModels "github.com/SENERGY-Platform/analytics-pipeline/lib"
)

const PIPELINE = "pipeline"

type Pipeline struct {
	pipeModels.Pipeline
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
		Nodes:              []engineModels.PipelineNode{},
	}
	r.WindowTime = p.WindowTime
	for _, operator := range p.Operators {
		node := engineModels.PipelineNode{
			NodeId:          operator.Id,
			Inputs:          []engineModels.NodeInput{},
			Config:          []engineModels.NodeConfig{},
			InputSelections: operator.InputSelections,
			PersistData:     operator.PersistData,
		}
		for _, inputTopic := range operator.InputTopics {
			nodeInput := engineModels.NodeInput{
				FilterType: inputTopic.FilterType,
				FilterIds:  inputTopic.FilterValue,
				TopicName:  inputTopic.Name,
				Values:     []engineModels.NodeValue{},
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
				nodeInput.Values = append(nodeInput.Values, engineModels.NodeValue{
					Name: mapping.Dest,
					Path: mapping.Source,
				})
			}
			node.Inputs = append(node.Inputs, nodeInput)
		}
		for key, value := range operator.Config {
			node.Config = append(node.Config, engineModels.NodeConfig{
				Name:  key,
				Value: value,
			})
		}
		r.Nodes = append(r.Nodes, node)
	}
	return &r
}

type PipelineRequest engineModels.PipelineRequest

type Workload struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	ImageUuid   string            `json:"imageUuid,omitempty"`
	Environment map[string]string `json:"environment"`
	Labels      map[string]string `json:"labels"`
}

type KubeService struct {
	Id                string   `json:"id"`
	BaseType          string   `json:"baseType"`
	Name              string   `json:"name"`
	TargetWorkloadIds []string `json:"targetWorkloadIds,omitempty"`
}
type OpenIdToken struct {
	AccessToken      string    `json:"access_token"`
	ExpiresIn        float64   `json:"expires_in"`
	RefreshExpiresIn float64   `json:"refresh_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	RequestTime      time.Time `json:"-"`
}

type DeleteStatus struct {
	Total     int
	Remaining int
	Running   bool
	Errors    []error
}
