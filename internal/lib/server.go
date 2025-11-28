/*
 * Copyright 2021 InfAI (CC SES)
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
	"errors"
	"net/http"

	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cs CleanupService
}

func NewServer(cs *CleanupService) *Server {
	return &Server{cs: *cs}
}

func (s Server) CreateServer() (err error) {
	s.cs.keycloak.Login()
	defer s.cs.keycloak.Logout()

	if !DebugMode() {
		gin.SetMode(gin.ReleaseMode)
	}
	port := GetEnv("SERVER_PORT", "8000")

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	prefix := r.Group("/api")

	prefix.Use(accessMiddleware())

	prefix.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	prefix.GET("/pipeservices", func(c *gin.Context) {
		pipes, errs := s.cs.getOrphanedPipelineServices()
		if len(errs) > 0 {
			Logger.Error("getOrphanedPipelineServices failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, pipes)
	})

	prefix.DELETE("/pipeservices/:id", func(c *gin.Context) {
		errs := s.cs.deleteOrphanedPipelineService(c.Param("id"), c.GetHeader("Authorization")[7:])
		if len(errs) > 0 {
			Logger.Error("deleteOrphanedPipelineService failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	prefix.DELETE("/pipeservices", func(c *gin.Context) {
		pipes, errs := s.cs.deleteOrphanedPipelineServices()
		if len(errs) > 0 {
			Logger.Error("deleteOrphanedPipelineServices failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, pipes)
	})

	prefix.GET("/analyticsworkloads", func(c *gin.Context) {
		wls, errs := s.cs.getOrphanedAnalyticsWorkloads()
		if len(errs) > 0 {
			Logger.Error("getOrphanedAnalyticsWorkloads failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, wls)
	})

	prefix.DELETE("/analyticsworkloads/:name", func(c *gin.Context) {
		err = s.cs.deleteOrphanedAnalyticsWorkload(c.Param("name"))
		if err != nil {
			Logger.Error("deleteOrphanedAnalyticsWorkload failed", "error", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	prefix.DELETE("/analyticsworkloads", func(c *gin.Context) {
		wls, errs := s.cs.deleteOrphanedAnalyticsWorkloads()
		if len(errs) > 0 {
			Logger.Error("deleteOrphanedAnalyticsWorkloads failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, wls)
	})

	prefix.GET("/pipelinekubeservices", func(c *gin.Context) {
		services, errs := s.cs.getOrphanedKubeServices(PIPELINE)
		if len(errs) > 0 {
			Logger.Error("getOrphanedKubeServices failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, services)
	})

	prefix.DELETE("/pipelinekubeservices/:id", func(c *gin.Context) {
		err = s.cs.deleteOrphanedKubeService(PIPELINE, c.Param("id"))
		if err != nil {
			Logger.Error("deleteOrphanedKubeService failed", "error", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	prefix.DELETE("/pipelinekubeservices", func(c *gin.Context) {
		services, errs := s.cs.deleteOrphanedKubeServices(PIPELINE)
		if len(errs) > 0 {
			Logger.Error("deleteOrphanedKubeServices failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, services)
	})

	prefix.GET("/kafkatopics", func(c *gin.Context) {
		topics, errs := s.cs.getOrphanedKafkaTopics()
		if len(errs) > 0 {
			Logger.Error("getOrphanedKafkaTopics failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, topics)
	})

	prefix.DELETE("/kafkatopics/:name", func(c *gin.Context) {
		err = s.cs.deleteOrphanedKafkaTopic(c.Param("name"))
		if err != nil {
			Logger.Error("deleteOrphanedKafkaTopic failed", "error", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	prefix.DELETE("/kafkatopics", func(c *gin.Context) {
		errs := s.cs.deleteOrphanedKafkaTopics()
		if len(errs) > 0 {
			Logger.Error("deleteOrphanedKafkaTopics failed", "error", errs)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})

	if !DebugMode() {
		err = r.Run(":" + port)
	} else {
		err = r.Run("127.0.0.1:" + port)
	}
	if err != nil {
		Logger.Error("could not start api server", "error", err)
	}
	return
}

func accessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if checkUserAdmin(c.GetHeader("Authorization")) {
			c.Next()
		}
		c.Status(http.StatusForbidden)
	}
}

func checkUserAdmin(tokenString string) (access bool) {
	access = false
	if tokenString != "" {
		claims, err := jwt.Parse(tokenString)
		if err != nil {
			err = errors.New("Error parsing token: " + err.Error())
			return
		}
		if StringInSlice("admin", claims.RealmAccess["roles"]) {
			Logger.Debug("Authenticated user " + claims.Sub)
			access = true
		}
	}
	return
}
