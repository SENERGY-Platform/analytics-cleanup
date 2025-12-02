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

package config

import (
	sb_config_hdl "github.com/SENERGY-Platform/go-service-base/config-hdl"
)

type LoggerConfig struct {
	Level string `json:"level" env_var:"LOGGER_LEVEL"`
}

type KeycloakConfig struct {
	Url          string `json:"url" env_var:"KEYCLOAK_URL"`
	Realm        string `json:"realm" env_var:"KEYCLOAK_REALM"`
	ClientId     string `json:"client_id" env_var:"KEYCLOAK_CLIENT_ID"`
	ClientSecret string `json:"client_secret" env_var:"KEYCLOAK_CLIENT_SECRET"`
	User         string `json:"user" env_var:"KEYCLOAK_USER"`
	Password     string `json:"password" env_var:"KEYCLOAK_PASSWORD"`
}

type Rancher2Config struct {
	Endpoint            string `json:"endpoint" env_var:"RANCHER2_ENDPOINT"`
	AccessKey           string `json:"access_key" env_var:"RANCHER2_ACCESS_KEY"`
	SecretKey           string `json:"secret_key" env_var:"RANCHER2_SECRET_KEY"`
	PipelineProjectId   string `json:"pipe_project_id" env_var:"RANCHER2_PIPELINE_PROJECT_ID"`
	PipelineNamespaceId string `json:"pipe_namespace_id" env_var:"RANCHER2_PIPELINE_NAMESPACE_ID"`
	ServingProjectId    string `json:"serv_project_id" env_var:"RANCHER2_SERVING_PROJECT_ID"`
}

type Config struct {
	Logger                LoggerConfig   `json:"logger" env_var:"LOGGER_CONFIG"`
	URLPrefix             string         `json:"url_prefix" env_var:"URL_PREFIX"`
	ServerPort            int            `json:"server_port" env_var:"SERVER_PORT"`
	Debug                 bool           `json:"debug" env_var:"DEBUG"`
	Keycloak              KeycloakConfig `json:"keycloak"`
	PipelineApiEndpoint   string         `json:"pipeline_api_endpoint" env_var:"PIPELINE_API_ENDPOINT"`
	FlowEngineApiEndpoint string         `json:"flow_engine_api_endpoint" env_var:"FLOW_ENGINE_API_ENDPOINT"`
	KafkaBootstrap        string         `json:"kafka_bootstrap" env_var:"KAFKA_BOOTSTRAP"`
	Mode                  string         `json:"mode" env_var:"MODE"`
	CronSchedule          string         `json:"cron_schedule" env_var:"CRON_SCHEDULE"`
	Driver                string         `json:"driver" env_var:"DRIVER"`
	Rancher2Config        Rancher2Config `json:"rancher2" env_var:"RANCHER2_CONFIG"`
}

func New(path string) (*Config, error) {
	cfg := Config{
		ServerPort:     8080,
		Debug:          false,
		Mode:           "web",
		KafkaBootstrap: "localhost:9092",
		Keycloak: KeycloakConfig{
			Url:          "http://localhost",
			ClientId:     "local",
			ClientSecret: "local",
		},
		Driver: "rancher2",
		Rancher2Config: Rancher2Config{
			PipelineNamespaceId: "analytics-pipelines",
		},
	}
	err := sb_config_hdl.Load(&cfg, nil, envTypeParser, nil, path)
	return &cfg, err
}
