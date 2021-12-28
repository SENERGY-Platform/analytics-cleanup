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

package main

import (
	"fmt"
	"github.com/SENERGY-Platform/analytics-cleanup/internal/lib"
	rancher_api "github.com/SENERGY-Platform/analytics-cleanup/internal/rancher-api"
	rancher2_api "github.com/SENERGY-Platform/analytics-cleanup/internal/rancher2-api"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	pipeline := lib.NewPipelineService(
		lib.GetEnv("PIPELINE_API_ENDPOINT", ""),
		lib.GetEnv("FLOW_ENGINE_API_ENDPOINT", ""),
	)
	serving := lib.NewServingService(
		lib.GetEnv("SERVING_API_ENDPOINT", ""),
	)
	var driver lib.Driver
	switch selectedDriver := lib.GetEnv("DRIVER", "rancher"); selectedDriver {
	case "rancher":
		driver = rancher_api.NewRancher(
			lib.GetEnv("RANCHER_ENDPOINT", ""),
			lib.GetEnv("RANCHER_ACCESS_KEY", ""),
			lib.GetEnv("RANCHER_SECRET_KEY", ""),
			lib.GetEnv("RANCHER_PIPELINES_STACK_ID", ""),
			lib.GetEnv("RANCHER_SERVING_STACK_ID", ""),
		)
	case "rancher2":
		driver = rancher2_api.NewRancher2(
			lib.GetEnv("RANCHER2_ENDPOINT", ""),
			lib.GetEnv("RANCHER2_ACCESS_KEY", ""),
			lib.GetEnv("RANCHER2_SECRET_KEY", ""),
			lib.GetEnv("RANCHER2_SERVING_NAMESPACE_ID", ""),
			lib.GetEnv("RANCHER2_SERVING_PROJECT_ID", ""),
			lib.GetEnv("RANCHER2_PIPELINE_NAMESPACE_ID", ""),
			lib.GetEnv("RANCHER2_PIPELINE_PROJECT_ID", ""),
		)
	default:
		panic("No driver selected")
	}
	keycloak := lib.NewKeycloakService(
		lib.GetEnv("KEYCLOAK_ADDRESS", "http://test"),
		lib.GetEnv("KEYCLOAK_CLIENT_ID", "test"),
		lib.GetEnv("KEYCLOAK_CLIENT_SECRET", "test"),
		lib.GetEnv("KEYCLOAK_REALM", "test"),
		lib.GetEnv("KEYCLOAK_USER", "test"),
		lib.GetEnv("KEYCLOAK_PW", "test"),
	)

	kafkaAdmin := lib.NewKafkaAdmin(lib.GetEnv("KAFKA_BOOTSTRAP", "127.0.0.1:9092"))

	if lib.GetEnv("MODE", "web") == "web" {
		fmt.Println("starting webserver")
		logger := lib.NewFileLogger("logs/cleanup.log", "")
		defer logger.Close()
		service := lib.NewCleanupService(*keycloak, driver, *pipeline, *serving, *logger, kafkaAdmin)
		server := lib.NewServer(service)
		server.CreateServer()
	} else {
		if lib.GetEnv("CRON_SCHEDULE", "* * * * *") == "false" {
			logger := lib.NewFileLogger("logs/cleanup.log", "")
			defer logger.Close()
			service := lib.NewCleanupService(*keycloak, driver, *pipeline, *serving, *logger, kafkaAdmin)
			service.StartCleanupService()
			os.Exit(0)
		} else {
			c := cron.New()
			_, err = c.AddFunc(lib.GetEnv("CRON_SCHEDULE", "* * * * *"), func() {
				log.Println("Start cleanup")
				currentTime := time.Now()
				if _, err := os.Stat("logs"); os.IsNotExist(err) {
					os.Mkdir("logs", 0644)
				}
				logger := lib.NewFileLogger("logs/cleanup-"+currentTime.Format("02-01-2006-15:04:05")+".log", "")
				defer logger.Close()
				service := lib.NewCleanupService(*keycloak, driver, *pipeline, *serving, *logger, kafkaAdmin)
				service.StartCleanupService()
			})
			if err != nil {
				log.Fatal("Error starting job: " + err.Error())
			}
			c.Start()
		}
	}
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-shutdown
	log.Println("received shutdown signal", sig)
}
