/*
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lib

import (
	"github.com/Shopify/sarama"
	"time"
)

type KafkaAdmin struct {
	kafkaBootstrap string
}

func NewKafkaAdmin(kafkaBootstrap string) *KafkaAdmin {
	return &KafkaAdmin{
		kafkaBootstrap: kafkaBootstrap,
	}
}

func (this *KafkaAdmin) DeleteTopic(name string) (err error) {
	admin, err := this.getAdmin()
	if err != nil {
		return err
	}
	defer admin.Close()
	return admin.DeleteTopic(name)
}

func (this *KafkaAdmin) GetTopics() (topics []string, err error) {
	admin, err := this.getAdmin()
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	topicInfos, err := admin.ListTopics()
	if err != nil {
		return nil, err
	}
	for k := range topicInfos {
		topics = append(topics, k)
	}
	return topics, nil
}

func (this *KafkaAdmin) getAdmin() (admin sarama.ClusterAdmin, err error) {
	sconfig := sarama.NewConfig()
	sconfig.Admin.Timeout = 25 * time.Second
	return sarama.NewClusterAdmin([]string{this.kafkaBootstrap}, sconfig)
}
