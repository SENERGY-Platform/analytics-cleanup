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

import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {ErrorHandlerService} from "../core/services/error-handler.service";
import {environment} from "../../environments/environment";
import {catchError, map} from "rxjs/operators";
import {Observable} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class CleanupService {

  constructor(private httpClient: HttpClient, private errorHandlerService: ErrorHandlerService) {}

  getOrphanedPipelineServices(): Observable<AnalyticsPipeline[] | null>  {
    return this.httpClient.get<AnalyticsPipeline[]>(environment.gateway + '/pipeservices').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedPipelineServices: Error', null)),
    );
  }

  getOrphanedServingServices(): Observable<Export[] | null>  {
    return this.httpClient.get<Export[]>(environment.gateway + '/servingservices').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedServingServices: Error', null)),
    );
  }

  getOrphanedAnalyticsWorkloads(): Observable<KubeWorkload[] | null>  {
    return this.httpClient.get<KubeWorkload[]>(environment.gateway + '/analyticsworkloads').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedAnalyticsWorkloads: Error', null)),
    );
  }

  getOrphanedServingWorkloads(): Observable<KubeWorkload[] | null>  {
    return this.httpClient.get<KubeWorkload[]>(environment.gateway + '/servingworkloads').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedServingWorkloads: Error', null)),
    );
  }

  getOrphanedServingKubeServices(): Observable<KubeService[] | null>  {
    return this.httpClient.get<KubeService[]>(environment.gateway + '/servingkubeservices').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedServingKubeServices: Error', null)),
    );
  }

  getOrphanedPipelineKubeServices(): Observable<KubeService[] | null>  {
    return this.httpClient.get<KubeService[]>(environment.gateway + '/pipelinekubeservices').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedServingKubeServices: Error', null)),
    );
  }

  getOrphanedInfluxMeasurements(): Observable<InfluxDatabase[] | null>  {
    return this.httpClient.get<InfluxDatabase[]>(environment.gateway + '/influxmeasurements').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedInfluxMeasurements: Error', null)),
    );
  }

  getOrphanedKafkaTopics(): Observable<string[] | null>  {
    return this.httpClient.get<string[]>(environment.gateway + '/kafkatopics').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedKafkaTopics: Error', null)),
    );
  }

  deletePipeline(id: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/pipeservices/' + id)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deletePipeline: Error', null)));
  }

  deleteServing(id: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/servingservices/' + id)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteServing: Error', null)));
  }

  deleteAnalyticsWorkload(name: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/analyticsworkloads/' + name)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteAnalyticsWorkload: Error', null)));
  }

  deleteServingWorkload(name: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/servingworkloads/' + name)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteServingWorkload: Error', null)));
  }

  deleteAnalyticsKubeService(id: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/pipelinekubeservices/' + id)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteAnalyticsKubeService: Error', null)));
  }

  deleteServingKubeService(id: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/servingkubeservices/' + id)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteServingKubeService: Error', null)));
  }

  deleteInfluxMeasurement(databaseId: string, measurementId: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/influxmeasurements/' + databaseId+'/'+measurementId)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteInfluxMeasurement: Error', null)));
  }

  deleteKafkaTopic(topic: string): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/kafkatopics/' + topic)
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteKafkaTopic: Error', null)));
  }

  deletePipelines(): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/pipeservices')
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deletePipelines: Error', null)));
  }

  deleteServings(): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/servingservices')
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteServings: Error', null)));
  }

  deleteAnalyticsWorkloads(): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/analyticsworkloads')
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteAnalyticsWorkloads: Error', null)));
  }

  deleteServingWorkloads(): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/servingworkloads')
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteServingWorkloads: Error', null)));
  }

  deleteServingKubeServices(): Observable<KubeService[] | null> {
    return this.httpClient.delete<KubeService[]>(environment.gateway + '/servingkubeservices').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'deleteServingKubeServices: Error', null)),
    );
  }

  deleteAnalyticsKubeServices(): Observable<KubeService[] | null> {
    return this.httpClient.delete<KubeService[]>(environment.gateway + '/pipelinekubeservices').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'deleteAnalyticsKubeServices: Error', null)),
    );
  }

  deleteInfluxMeasurements(): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/influxmeasurements')
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteInfluxMeasurements: Error', null)));
  }

  deleteKafkaTopics(): Observable<unknown> {
    return this.httpClient
      .delete(environment.gateway + '/kafkatopics')
      .pipe(catchError(this.errorHandlerService.handleError(CleanupService.name, 'deleteKafkaTopics: Error', null)));
  }

}

export interface AnalyticsPipeline {
  id: string;
  name: string;
  UserId: string
  createdAt: string
  updatedAt: string
}

export interface Export {
  ID: string;
  Name: string;
  UserId: string
  CreatedAt: string
  UpdatedAt: string
}

export interface KubeWorkload {
  id: string;
  name: string;
  imageUuid: string
}

export interface KubeService {
  id: string;
  baseType: string;
  name: string
}

export interface InfluxDatabase {
  id: string;
  databaseId: string;
}

