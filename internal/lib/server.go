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
	router.HandleFunc("/api/health", s.healthCheck).Methods("GET")
	router.HandleFunc("/api/pipeservices", s.getOrphanedPipelineServices).Methods("GET")
	router.HandleFunc("/api/analyticsworkloads", s.getOrphanedAnalyticsWorkloads).Methods("GET")
	router.HandleFunc("/api/servingservices", s.getOrphanedServingServices).Methods("GET")
	router.HandleFunc("/api/servingworkloads", s.getOrphanedServingWorkloads).Methods("GET")
	router.HandleFunc("/api/kubeservices", s.getOrphanedKubeServices).Methods("GET")
	router.HandleFunc("/api/influxmeasurements", s.getOrphanedInfluxMeasurements).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/dist/ui")))
	logger := NewWebLogger(router, "CALL")
	access := accessMiddleware(logger)
	c := cors.New(
		cors.Options{
			AllowedHeaders: []string{"Content-Type", "Authorization", "Accept", "Accept-Encoding", "X-CSRF-Token"},
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS", "PUT"},
		})
	handler := c.Handler(access)
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
	_ = json.NewEncoder(w).Encode(s.cs.getOrphanedPipelineServices())
}

func (s Server) getOrphanedAnalyticsWorkloads(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(s.cs.getOrphanedAnalyticsWorkloads())
}

func (s Server) getOrphanedServingServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(s.cs.getOrphanedServingServices())
}

func (s Server) getOrphanedServingWorkloads(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(s.cs.getOrphanedServingWorkloads())
}

func (s Server) getOrphanedKubeServices(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(s.cs.getOrphanedServingKubeServices())
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
