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

package apis

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"strconv"

	"github.com/SENERGY-Platform/analytics-cleanup/lib"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

type PipelineService struct {
	pipelineUrl string
	engineUrl   string
}

func NewPipelineService(pipelineUrl string, engineUrl string) *PipelineService {
	return &PipelineService{pipelineUrl: pipelineUrl, engineUrl: engineUrl}
}

func (p PipelineService) GetPipelines(userId string, accessToken string) (pipes []lib.Pipeline, errs []error) {
	request := gorequest.New()
	resp, body, errs := request.Get(p.pipelineUrl+"/admin/pipeline").Set("X-UserId", userId).
		Set("Authorization", "Bearer "+accessToken).End()
	if len(errs) < 1 {
		if resp.StatusCode != 200 {
			return pipes, []error{errors.New("could not access pipeline registry: " + strconv.Itoa(resp.StatusCode) + " " + body)}
		}
		err := json.Unmarshal([]byte(body), &pipes)
		if err != nil {
			errs = append(errs, json.Unmarshal([]byte(body), &pipes))
		}
	}
	return
}

func (p PipelineService) DeletePipeline(id string, accessToken string) (errs []error) {
	request := gorequest.New()
	resp, body, errs := request.Delete(p.pipelineUrl+"/admin/pipeline/"+id).Set("Authorization", "Bearer "+accessToken).End()
	if len(errs) < 1 {
		if resp.StatusCode != 200 {
			errs = append(errs, errors.New("could not access pipeline registry: "+strconv.Itoa(resp.StatusCode)+" "+body))
		}
	}
	return
}

func (p PipelineService) CreatePipeline(instance *lib.PipelineRequest, userId string, userToken string) error {
	b, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	println(instance.Name)
	req, err := http.NewRequest(http.MethodPut, p.engineUrl+"/pipeline", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("X-UserId", userId)
	req.Header.Set("Content-Type", "application/json")

	http.DefaultClient.Timeout = 10 * time.Second

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected status code " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}
