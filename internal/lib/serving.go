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
	"errors"
	"github.com/parnurzeal/gorequest"
	"strconv"
)

type ServingService struct {
	url string
}

func NewServingService(url string) *ServingService {
	return &ServingService{url: url}
}

func (s *ServingService) GetServingServices(userId string, accessToken string) (servings []ServingInstance, errs []error) {
	request := gorequest.New()
	resp, body, errs := request.Get(s.url+"/admin/instance").Set("X-UserId", userId).
		Set("Authorization", "Bearer "+accessToken).End()
	if len(errs) < 1 {
		if resp.StatusCode != 200 {
			return servings, []error{errors.New("could not access serving service: " + strconv.Itoa(resp.StatusCode) + " " + body)}
		}
		errs[0] = json.Unmarshal([]byte(body), &servings)
	}
	return
}

func (s *ServingService) DeleteServingService(id string, userId string, accessToken string) (errs []error) {
	request := gorequest.New()
	resp, body, errs := request.Delete(s.url+"/admin/instance/"+id).Set("X-UserId", userId).Set("Authorization", "Bearer "+accessToken).End()
	if len(errs) < 1 {
		if resp.StatusCode != 204 {
			errs[0] = errors.New("could not delete serving: " + strconv.Itoa(resp.StatusCode) + " " + body)
		}
	}
	return
}

func servingGetDataAndTagFields(requestValues []ServingInstanceValue) (dataFields string, tagFields string) {
	dataFields = "{"
	tagFields = "{"
	for _, value := range requestValues {
		if value.Tag {
			if len(tagFields) > 1 {
				tagFields = tagFields + ","
			}
			tagFields = tagFields + "\"" + value.Name + ":" + value.Type + "\":\"" + value.Path + "\""
		} else {
			if len(dataFields) > 1 {
				dataFields = dataFields + ","
			}
			dataFields = dataFields + "\"" + value.Name + ":" + value.Type + "\":\"" + value.Path + "\""
		}
	}
	dataFields = dataFields + "}"
	tagFields = tagFields + "}"
	return
}
