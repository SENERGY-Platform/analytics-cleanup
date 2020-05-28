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
	"fmt"

	"strconv"

	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
)

type PipelineService struct {
	url string
}

func NewPipelineService(url string) *PipelineService {
	return &PipelineService{url: url}
}

func (p PipelineService) GetPipelines(userId string, accessToken string) (pipes []Pipeline, err error) {
	request := gorequest.New()
	resp, body, _ := request.Get(p.url+"/admin/pipeline").Set("X-UserId", userId).Set("Authorization", "Bearer "+accessToken).End()
	if resp.StatusCode != 200 {
		fmt.Println("could not access pipeline registry: "+strconv.Itoa(resp.StatusCode), resp.Body)
		return pipes, errors.New("could not access pipeline registry")
	}
	err = json.Unmarshal([]byte(body), &pipes)
	return
}

func (p PipelineService) DeletePipeline(id string, userId string, accessToken string) (err error) {
	request := gorequest.New()
	resp, _, e := request.Delete(p.url+"/admin/pipeline/"+id).Set("X-UserId", userId).Set("Authorization", "Bearer "+accessToken).End()
	if resp.StatusCode != 200 {
		fmt.Println("could not access pipeline registry: "+strconv.Itoa(resp.StatusCode), resp.Body)
		err = errors.New("could not access pipeline registry")
	}
	if len(e) > 0 {
		fmt.Println("something went wrong", e)
		err = errors.New("could not get pipeline from service")
	}
	return
}
