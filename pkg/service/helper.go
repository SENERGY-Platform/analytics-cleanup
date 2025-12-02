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
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/SENERGY-Platform/analytics-cleanup/lib"
	pipeModels "github.com/SENERGY-Platform/analytics-pipeline/lib"
)

func pipeInWorkloads(pipe pipeModels.Pipeline, workloads []lib.Workload) bool {
	for _, workload := range workloads {
		if strings.Contains(workload.Name, pipe.Id) {
			return true
		}
	}
	return false
}

func workloadInPipes(workload lib.Workload, pipes []pipeModels.Pipeline) bool {
	for _, pipe := range pipes {
		if strings.Contains(workload.Name, pipe.Id) {
			return true
		}
	}
	return false
}

func serviceInWorkloads(service lib.KubeService, workloads []lib.Workload) bool {
	ws := strings.Split(service.TargetWorkloadIds[0], ":")
	if len(ws) == 3 {
		for _, workload := range workloads {
			if workload.Name == ws[2] {
				return true
			}
		}
	}
	return false
}

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func ToJson(resp string) map[string]interface{} {
	data := map[string]interface{}{}
	json.Unmarshal([]byte(resp), &data)
	return data
}

func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
}

var kafkaInternalAnalyticsRx = regexp.MustCompile("(analytics-[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}.*-(repartition|changelog))")
var kafkaInternalAnalyticsPipelineIdRx = regexp.MustCompile("analytics-([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})")

func isInternalAnalyticsTopic(topic string) bool {
	return kafkaInternalAnalyticsRx.MatchString(topic)
}

func pipelineExists(topic string, envs []map[string]string) bool {
	id := kafkaInternalAnalyticsPipelineIdRx.FindString(topic)
	for _, env := range envs {
		appId, ok := env["CONFIG_APPLICATION_ID"]
		if !ok {
			continue
		}
		if strings.HasPrefix(appId, id) {
			return true
		}
	}
	return false
}
