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

