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
// @Summary Get all orphaned pipe service
// @Description	Gets all orphaned pipe services
// @Produce json
// @Success	200 {array} lib.Pipeline
// @Failure	500 {string} str
// @Router /pipeservices [get]
func getOrphanedPipelineServices(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/pipeservices", func(c *gin.Context) {
		pipes, errs := service.GetOrphanedPipelineServices()
		if len(errs) > 0 {
			util.Logger.Error("could not get OrphanedPipelineServices", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, pipes)
	}
}

func deleteOrphanedPipelineService(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipeservices/:id", func(c *gin.Context) {
		errs := service.DeleteOrphanedPipelineService(c.Param("id"), c.GetHeader(HeaderAuth)[7:])
		if len(errs) > 0 {
			util.Logger.Error("could not delete OrphanedPipelineService", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func deleteOrphanedPipelineServices(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipeservices", func(c *gin.Context) {
		pipes, errs := service.DeleteOrphanedPipelineServices()
		if len(errs) > 0 {
			util.Logger.Error("could not delete OrphanedPipelineServices", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, pipes)
	}
}

func getOrphanedAnalyticsWorkloads(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/analyticsworkloads", func(c *gin.Context) {
		wls, errs := service.GetOrphanedAnalyticsWorkloads()
		if len(errs) > 0 {
			util.Logger.Error("could not get OrphanedAnalyticsWorkloads", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, wls)
	}
}

func deleteOrphanedAnalyticsWorkload(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/analyticsworkloads/:name", func(c *gin.Context) {
		err := service.DeleteOrphanedAnalyticsWorkload(c.Param("name"))
		if err != nil {
			util.Logger.Error("could not delete OrphanedAnalyticsWorkload", "error", err)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func deleteOrphanedAnalyticsWorkloads(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/analyticsworkloads", func(c *gin.Context) {
		wls, errs := service.DeleteOrphanedAnalyticsWorkloads()
		if len(errs) > 0 {
			util.Logger.Error("could not delete OrphanedAnalyticsWorkloads", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, wls)
	}
}

func getOrphanedKubeServices(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/pipelinekubeservices", func(c *gin.Context) {
		wls, errs := service.GetOrphanedKubeServices(lib.PIPELINE)
		if len(errs) > 0 {
			util.Logger.Error("could not get OrphanedKubeServices", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, wls)
	}
}

func deleteOrphanedKubeService(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipelinekubeservices/:id", func(c *gin.Context) {
		err := service.DeleteOrphanedKubeService(lib.PIPELINE, c.Param("id"))
		if err != nil {
			util.Logger.Error("could not delete OrphanedKubeService", "error", err)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func deleteOrphanedKubeServices(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/pipelinekubeservices", func(c *gin.Context) {
		services, errs := service.DeleteOrphanedKubeServices(lib.PIPELINE)
		if len(errs) > 0 {
			util.Logger.Error("could not delete OrphanedKubeServices", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, services)
	}
}

func getOrphanedKafkaTopics(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/kafkatopics", func(c *gin.Context) {
		topics, errs := service.GetOrphanedKafkaTopics()
		if len(errs) > 0 {
			util.Logger.Error("could not get OrphanedKafkaTopics", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.JSON(http.StatusOK, topics)
	}
}

func deleteOrphanedKafkaTopic(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/kafkatopics/:name", func(c *gin.Context) {
		err := service.DeleteOrphanedKafkaTopic(c.Param("name"))
		if err != nil {
			util.Logger.Error("could not delete OrphanedKafkaTopic", "error", err)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func deleteOrphanedKafkaTopics(service service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodDelete, "/kafkatopics", func(c *gin.Context) {
		errs := service.DeleteOrphanedKafkaTopics()
		if len(errs) > 0 {
			util.Logger.Error("could not delete OrphanedKafkaTopics", "error", errs)
			_ = c.Error(errors.New(MessageSomethingWrong))
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func getHealthCheckH(_ service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, HealthCheckPath, func(c *gin.Context) {
		c.Status(http.StatusOK)
	}
}

func getSwaggerDocH(_ service.CleanupService) (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/doc", func(gc *gin.Context) {
		if _, err := os.Stat("docs/swagger.json"); err != nil {
			_ = gc.Error(err)
			return
		}
		gc.Header("Content-Type", gin.MIMEJSON)
		gc.File("docs/swagger.json")
	}
}
