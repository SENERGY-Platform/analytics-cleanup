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
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type Server struct {
	cs CleanupService
}

func NewServer(cs *CleanupService) *Server {
	return &Server{cs: *cs}
}

func (s Server) CreateServer() {
	router := mux.NewRouter()
	s.cs.keycloak.Login()
	defer s.cs.keycloak.Logout()
	apiHandler := router.PathPrefix("/api").Subrouter()
	apiHandler.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)
	apiHandler.HandleFunc("/pipeservices", s.getOrphanedPipelineServices).Methods(http.MethodGet)
	apiHandler.HandleFunc("/pipeservices/{id}", s.deleteOrphanedPipelineService).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/pipeservices", s.deleteOrphanedPipelineServices).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/analyticsworkloads", s.getOrphanedAnalyticsWorkloads).Methods(http.MethodGet)
	apiHandler.HandleFunc("/analyticsworkloads/{name}", s.deleteOrphanedAnalyticsWorkload).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/analyticsworkloads", s.deleteOrphanedAnalyticsWorkloads).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/servingservices", s.getOrphanedServingServices).Methods(http.MethodGet)
	apiHandler.HandleFunc("/servingservices/{id}", s.deleteOrphanedServingService).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/servingservices", s.deleteOrphanedServingServices).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/servingworkloads", s.getOrphanedServingWorkloads).Methods(http.MethodGet)
	apiHandler.HandleFunc("/servingworkloads/{name}", s.deleteOrphanedServingWorkload).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/servingworkloads", s.deleteOrphanedServingWorkloads).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/servingkubeservices", s.getOrphanedServingKubeServices).Methods(http.MethodGet)
	apiHandler.HandleFunc("/servingkubeservices/{id}", s.deleteOrphanedServingKubeService).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/servingkubeservices", s.deleteOrphanedServingKubeServices).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/pipelinekubeservices", s.getOrphanedPipelineKubeServices).Methods(http.MethodGet)
	apiHandler.HandleFunc("/pipelinekubeservices/{id}", s.deleteOrphanedPipelineKubeService).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/pipelinekubeservices", s.deleteOrphanedPipelineKubeServices).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/influxmeasurements", s.getOrphanedInfluxMeasurements).Methods(http.MethodGet)
	apiHandler.HandleFunc("/influxmeasurements/{databaseId}/{measurementId}", s.deleteOrphanedInfluxMeasurement).Methods(http.MethodDelete)
	apiHandler.HandleFunc("/kafkatopics", s.getOrphanedKafkaTopics).Methods(http.MethodGet)
	apiHandler.HandleFunc("/kafkatopics/{name}", s.deleteOrphanedKafkaTopic).Methods(http.MethodDelete)
	apiHandler.Use(accessMiddleware)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/dist/ui")))
	logger := NewWebLogger(router, "CALL")
	c := cors.New(
		cors.Options{
			AllowedHeaders: []string{"Content-Type", "Authorization", "Accept", "Accept-Encoding", "X-CSRF-Token"},
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"},
		})
	handler := c.Handler(logger)
	log.Fatal(http.ListenAndServe(GetEnv("SERVERNAME", "")+":"+GetEnv("PORT", "8000"), handler))
}

func (s Server) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(Response{Message: "OK"})
}

func (s Server) getOrphanedPipelineServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	pipes, errs := s.cs.getOrphanedPipelineServices()
	if len(errs) > 0 {
		log.Printf("getOrphanedPipelineServices failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(pipes)
	}
}

func (s Server) deleteOrphanedPipelineService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	errs := s.cs.deleteOrphanedPipelineService(vars["id"], req.Header.Get("Authorization")[7:])
	if len(errs) > 0 {
		log.Printf("deleteOrphanedPipelineService failed: %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedPipelineServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	pipes, errs := s.cs.deleteOrphanedPipelineServices()
	if len(errs) > 0 {
		log.Printf("deleteOrphanedPipelineServices failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(pipes)
	}
}

func (s Server) getOrphanedAnalyticsWorkloads(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	workloads, errs := s.cs.getOrphanedAnalyticsWorkloads()
	if len(errs) > 0 {
		log.Printf("getOrphanedAnalyticsWorkloads failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(workloads)
	}
}

func (s Server) deleteOrphanedAnalyticsWorkload(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	err := s.cs.deleteOrphanedAnalyticsWorkload(vars["name"])
	if err != nil {
		log.Printf("deleteOrphanedAnalyticsWorkloads failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedAnalyticsWorkloads(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	workloads, errs := s.cs.deleteOrphanedAnalyticsWorkloads()
	if len(errs) > 0 {
		log.Printf("deleteOrphanedAnalyticsWorkloads failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(workloads)
	}
}

func (s Server) getOrphanedServingServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(s.cs.getOrphanedServingServices())
}

func (s Server) deleteOrphanedServingService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	errs := s.cs.deleteOrphanedServingService(vars["id"], req.Header.Get("Authorization")[7:])
	if len(errs) > 0 {
		log.Printf("deleteOrphanedServingService failed: %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedServingServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	errs := s.cs.deleteOrphanedServingServices()
	if len(errs) > 0 {
		log.Printf("deleteOrphanedServingServices failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) getOrphanedServingWorkloads(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	workloads, errs := s.cs.getOrphanedServingWorkloads()
	if len(errs) > 0 {
		log.Printf("getOrphanedServingWorkloads failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(workloads)
	}
}

func (s Server) deleteOrphanedServingWorkload(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	err := s.cs.deleteOrphanedServingWorkload(vars["name"])
	if err != nil {
		log.Printf("deleteOrphanedServingWorkload failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedServingWorkloads(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	workloads, errs := s.cs.deleteOrphanedServingWorkloads()
	if len(errs) > 0 {
		log.Printf("deleteOrphanedServingWorkloads failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(workloads)
	}
}

func (s Server) getOrphanedServingKubeServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	services, errs := s.cs.getOrphanedKubeServices(SERVING)
	if len(errs) > 0 {
		log.Printf("getOrphanedKafkaTopics failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(services)
}

func (s Server) deleteOrphanedServingKubeService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	err := s.cs.deleteOrphanedKubeService(SERVING, vars["id"])
	if err != nil {
		log.Printf("deleteOrphanedServingKubeService failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedServingKubeServices(w http.ResponseWriter, req *http.Request) {
	services, errs := s.cs.deleteOrphanedKubeServices(SERVING)
	if len(errs) > 0 {
		log.Printf("deleteOrphanedKubeServices failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(services)
	}
}

func (s Server) getOrphanedPipelineKubeServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	services, errs := s.cs.getOrphanedKubeServices(PIPELINE)
	if len(errs) > 0 {
		log.Printf("getOrphanedKafkaTopics failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(services)
}

func (s Server) deleteOrphanedPipelineKubeService(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	err := s.cs.deleteOrphanedKubeService(PIPELINE, vars["id"])
	if err != nil {
		log.Printf("deleteOrphanedPipelineKubeService failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedPipelineKubeServices(w http.ResponseWriter, req *http.Request) {
	services, errs := s.cs.deleteOrphanedKubeServices(PIPELINE)
	if len(errs) > 0 {
		log.Printf("deleteOrphanedPipelineKubeServices failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(services)
	}
}

func (s Server) getOrphanedInfluxMeasurements(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	measurements, err := s.cs.getOrphanedInfluxMeasurements()
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(measurements)
	}
}

func (s Server) deleteOrphanedInfluxMeasurement(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	err := s.cs.deleteOrphanedInfluxMeasurement(vars["measurementId"], vars["databaseId"])
	if err != nil {
		log.Printf("deleteOrphanedInfluxMeasurement failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedInfluxMeasurements(w http.ResponseWriter, req *http.Request) {
	err := s.cs.deleteOrphanedInfluxMeasurements()
	if err != nil {
		log.Printf("deleteOrphanedInfluxMeasurements failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) getOrphanedKafkaTopics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	topics, errs := s.cs.getOrphanedKafkaTopics()
	if len(errs) > 0 {
		log.Printf("getOrphanedKafkaTopics failed %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(topics)
}

func (s Server) deleteOrphanedKafkaTopic(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	err := s.cs.deleteOrphanedKafkaTopic(vars["name"])
	if err != nil {
		log.Printf("deleteOrphanedKafkaTopic failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s Server) deleteOrphanedKafkaTopics(w http.ResponseWriter, req *http.Request) {
	errs := s.cs.deleteOrphanedKafkaTopics()
	if len(errs) > 0 {
		log.Printf("deleteOrphanedKafkaTopics failed: %s", errs)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func accessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkUserAdmin(r) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

func checkUserAdmin(req *http.Request) (access bool) {
	access = false
	if req.Header.Get("Authorization") != "" {
		token, claims := parseJWTToken(req.Header.Get("Authorization")[7:])
		if token.Valid {
			if StringInSlice("admin", claims.RealmAccess["roles"]) {
				log.Printf("Authenticated user %s\n", claims.Sub)
				access = true
			}
		} else {
			log.Printf("Invalid token for user %s\n", claims.Sub)
		}
	}
	return
}

func parseJWTToken(encodedToken string) (token *jwt.Token, claims Claims) {
	const PEM_BEGIN = "-----BEGIN PUBLIC KEY-----"
	const PEM_END = "-----END PUBLIC KEY-----"

	token, _ = jwt.ParseWithClaims(encodedToken, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		switch GetEnv("JWT_SIGNING_METHOD", "rsa") {
		case "rsa":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			key := GetEnv("JWT_SIGNING_KEY", "")
			if !strings.HasPrefix(key, PEM_BEGIN) {
				key = PEM_BEGIN + "\n" + key + "\n" + PEM_END
			}
			return jwt.ParseRSAPublicKeyFromPEM([]byte(key))
		case "hmac":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(GetEnv("JWT_SIGNING_KEY", "")), nil
		}
		return "", nil
	})
	return
}
