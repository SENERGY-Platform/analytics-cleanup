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
	"encoding/json"
	"os"
	"strings"
)

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

func IntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
}
