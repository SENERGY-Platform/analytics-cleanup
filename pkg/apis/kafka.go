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
	"errors"
	"time"

	"github.com/IBM/sarama"
	"github.com/SENERGY-Platform/analytics-cleanup/lib"
)

type KafkaAdmin struct {
	kafkaBootstrap string
	clusterAdmin   sarama.ClusterAdmin
}

func NewKafkaAdmin(kafkaBootstrap string) (*KafkaAdmin, error) {
	conf := sarama.NewConfig()
	conf.Admin.Timeout = 25 * time.Second
	admin, err := sarama.NewClusterAdmin([]string{kafkaBootstrap}, conf)
	if err != nil {
		return nil, err
	}
	return &KafkaAdmin{
		kafkaBootstrap: kafkaBootstrap,
		clusterAdmin:   admin,
	}, err
}

func (k *KafkaAdmin) Close() (err error) {
	if k.clusterAdmin != nil {
		defer func(admin sarama.ClusterAdmin) {
			err = admin.Close()
		}(k.clusterAdmin)
	}
	return
}

func (k *KafkaAdmin) DeleteTopic(name string) (err error) {
	err = k.clusterAdmin.DeleteTopic(name)
	if errors.Is(err, sarama.ErrUnknownTopicOrPartition) {
		err = lib.NewNotFoundError(err)
	}
	return
}

func (k *KafkaAdmin) GetTopics() (topics []string, err error) {
	topicInfos, err := k.clusterAdmin.ListTopics()
	if err != nil {
		return nil, err
	}
	for k := range topicInfos {
		topics = append(topics, k)
	}
	return topics, nil
}
