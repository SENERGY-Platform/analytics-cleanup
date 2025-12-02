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

package api

import (
	"errors"
	"net/http"
	"os"

	"github.com/SENERGY-Platform/analytics-cleanup/lib"
	"github.com/SENERGY-Platform/analytics-cleanup/pkg/service"
	"github.com/SENERGY-Platform/analytics-cleanup/pkg/util"
	"github.com/gin-gonic/gin"
)

// getOrphanedPipelineServices godoc
// @Summary Get all orphaned pipe services
// @Description	Gets all orphaned pipe services
// @Tags pipeline-services
// @Produce json
// @Success	200 {array} lib.Pipeline
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /pipeservices [get]
func getOrphanedPipelineServices(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/pipeservices", func(c *gin.Context) {
		pipes, err := service.GetOrphanedPipelineServices(c.GetString(UserIdKey), c.GetHeader(HeaderAuth)[7:])
		if err != nil {
			util.Logger.Error("could not get OrphanedPipelineServices", "error", err)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, pipes)
	}
}

// deleteOrphanedPipelineService godoc
// @Summary Delete orphaned pipeline service
// @Description Deletes an orphaned pipeline service by ID
// @Tags pipeline-services
// @Param   id path string true "Pipeline Service ID"
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "something went wrong"
// @Router /pipeservices/{id} [delete]
func deleteOrphanedPipelineService(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipeservices/:id", func(c *gin.Context) {
		err := service.DeleteOrphanedPipelineService(c.Param("id"), c.GetHeader(HeaderAuth)[7:])
		if err != nil {
			util.Logger.Error("could not delete OrphanedPipelineService", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// deleteOrphanedPipelineServices godoc
// @Summary Delete orphaned pipeline services
// @Description Deletes all orphaned pipeline services
// @Tags pipeline-services
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /pipeservices [delete]
func deleteOrphanedPipelineServices(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipeservices", func(c *gin.Context) {
		_, err := service.DeleteOrphanedPipelineServices(c.Param("id"), c.GetHeader(HeaderAuth)[7:])
		if err != nil {
			util.Logger.Error("could not delete OrphanedPipelineServices", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.Status(http.StatusNotFound)
	}
}

// getOrphanedAnalyticsWorkloads godoc
// @Summary Get all orphaned workloads
// @Description	Gets all orphaned workloads
// @Tags workloads
// @Produce json
// @Success	200 {array} lib.Workload
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /analyticsworkloads [get]
func getOrphanedAnalyticsWorkloads(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/analyticsworkloads", func(c *gin.Context) {
		wls, err := service.GetOrphanedAnalyticsWorkloads(c.Param("id"), c.GetHeader(HeaderAuth)[7:])
		if err != nil {
			util.Logger.Error("could not get OrphanedAnalyticsWorkloads", "error", err)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, wls)
	}
}

// deleteOrphanedAnalyticsWorkload godoc
// @Summary Delete orphaned workload
// @Description Deletes an orphaned workload by name
// @Tags workloads
// @Param   name path string true "Workload name"
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "something went wrong"
// @Router /analyticsworkloads/{name} [delete]
func deleteOrphanedAnalyticsWorkload(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/analyticsworkloads/:name", func(c *gin.Context) {
		err := service.DeleteOrphanedAnalyticsWorkload(c.Param("name"))
		if err != nil {
			util.Logger.Error("could not delete OrphanedAnalyticsWorkload", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// deleteOrphanedAnalyticsWorkloads godoc
// @Summary Delete orphaned workloads
// @Description Deletes all orphaned workloads
// @Tags workloads
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /analyticsworkloads [delete]
func deleteOrphanedAnalyticsWorkloads(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/analyticsworkloads", func(c *gin.Context) {
		wls, err := service.DeleteOrphanedAnalyticsWorkloads(c.Param("id"), c.GetHeader(HeaderAuth)[7:])
		if err != nil {
			util.Logger.Error("could not delete OrphanedAnalyticsWorkloads", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.JSON(http.StatusOK, wls)
	}
}

// getOrphanedKubeServices godoc
// @Summary Get all orphaned kube services
// @Description	Gets all orphaned kube services
// @Tags kube-services
// @Produce json
// @Success	200 {array} lib.KubeService
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /pipelinekubeservices [get]
func getOrphanedKubeServices(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/pipelinekubeservices", func(c *gin.Context) {
		wls, err := service.GetOrphanedKubeServices(lib.PIPELINE)
		if err != nil {
			util.Logger.Error("could not get OrphanedKubeServices", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.JSON(http.StatusOK, wls)
	}
}

// deleteOrphanedKubeService godoc
// @Summary Delete orphaned kube service
// @Description Deletes an orphaned kube service by name
// @Tags kube-services
// @Param name path string true "Kube Service name"
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "something went wrong"
// @Router /pipelinekubeservices/{name} [delete]
func deleteOrphanedKubeService(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipelinekubeservices/:id", func(c *gin.Context) {
		err := service.DeleteOrphanedKubeService(lib.PIPELINE, c.Param("id"))
		if err != nil {
			util.Logger.Error("could not delete OrphanedKubeService", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// deleteOrphanedKubeServices godoc
// @Summary Delete orphaned kube services
// @Description Deletes all orphaned kube services
// @Tags kube-services
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /pipelinekubeservices [delete]
func deleteOrphanedKubeServices(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipelinekubeservices", func(c *gin.Context) {
		services, err := service.DeleteOrphanedKubeServices(lib.PIPELINE)
		if err != nil {
			util.Logger.Error("could not delete OrphanedKubeServices", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.JSON(http.StatusOK, services)
	}
}

// getOrphanedKafkaTopics godoc
// @Summary Get all orphaned kafka topics
// @Description	Gets all orphaned kafka topics
// @Tags kafka-topics
// @Produce json
// @Success	200 {array} string
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /kafkatopics [get]
func getOrphanedKafkaTopics(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/kafkatopics", func(c *gin.Context) {
		topics, err := service.GetOrphanedKafkaTopics()
		if err != nil {
			util.Logger.Error("could not get OrphanedKafkaTopics", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.JSON(http.StatusOK, topics)
	}
}

// deleteOrphanedKafkaTopic godoc
// @Summary Delete orphaned kafka topic
// @Description Deletes an orphaned kafka topic by name
// @Tags kafka-topics
// @Param name path string true "Kafka Topic name"
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "something went wrong"
// @Router /kafkatopics/{name} [delete]
func deleteOrphanedKafkaTopic(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/kafkatopics/:name", func(c *gin.Context) {
		err := service.DeleteOrphanedKafkaTopic(c.Param("name"))
		if err != nil {
			util.Logger.Error("could not delete OrphanedKafkaTopic", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// deleteOrphanedKafkaTopics godoc
// @Summary Delete orphaned kafka topics
// @Description Deletes all orphaned kafka topics
// @Tags kafka-topics
// @Success 204
// @Failure 403 {string} string "forbidden"
// @Failure 409 {string} string "already running"
// @Failure 500 {string} string "something went wrong"
// @Router /kafkatopics [delete]
func deleteOrphanedKafkaTopics(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/kafkatopics", func(c *gin.Context) {
		err := service.DeleteOrphanedKafkaTopics()
		if err != nil {
			util.Logger.Error("could not delete OrphanedKafkaTopics", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// getDeleteOrphanedKafkaTopicsStatus godoc
// @Summary Get kafka topic deletion status
// @Description Get the status of kafka topic deletion
// @Tags kafka-topics
// @Success 200 {object} lib.DeleteStatus
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /kafkatopics/status [get]
func getDeleteOrphanedKafkaTopicsStatus(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/kafkatopics/status", func(c *gin.Context) {
		data := service.GetDeleteOrphanedKafkaTopicsStatus()
		c.JSON(http.StatusOK, data)
	}
}

// stopDeleteOrphanedKafkaTopics godoc
// @Summary Stop kafka topic deletion
// @Description Stops the deletion process of kafka topics
// @Tags kafka-topics
// @Success 200
// @Failure 403 {string} string "forbidden"
// @Failure 500 {string} string "something went wrong"
// @Router /kafkatopics/stop [post]
func stopDeleteOrphanedKafkaTopics(service *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/kafkatopics/stop", func(c *gin.Context) {
		err := service.StopDeleteOrphanedKafkaTopics()
		if err != nil {
			util.Logger.Error("could not stop the deletion of OrphanedKafkaTopics", "error", err)
			_ = c.Error(handleError(err))
			return
		}
		c.Status(http.StatusOK)
	}
}

func getHealthCheckH(_ *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, HealthCheckPath, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}

func getSwaggerDocH(_ *service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/doc", func(gc *gin.Context) {
		if _, err := os.Stat("docs/swagger.json"); err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Header("Content-Type", gin.MIMEJSON)
		gc.File("docs/swagger.json")
	}
}
