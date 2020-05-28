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
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
	file   *os.File
}

func NewLogger(fileName string, prefix string) *Logger {
	file, err := os.OpenFile(fileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	logger := log.New(file, prefix, log.LstdFlags)
	return &Logger{logger: logger, file: file}
}

func (l Logger) Close() {
	err := l.file.Close()
	if err != nil {
		log.Println(err)
	}
}

func (l Logger) Print(v ...interface{}) {
	log.Println(v)
	l.logger.Println(v)
}
