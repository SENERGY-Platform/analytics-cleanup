import { Injectable } from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {ErrorHandlerService} from "../../core/services/error-handler.service";
import {environment} from "../../../environments/environment";
import {catchError, map} from "rxjs/operators";
import {Observable} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class PipesService {

  constructor(private httpClient: HttpClient, private errorHandlerService: ErrorHandlerService) {}

  getOrphanedPipelineServices(): Observable<AnalyticsPipeline[] | null>  {
    return this.httpClient.get<AnalyticsPipeline[]>(environment.gateway + '/pipeservices').pipe(
      map((resp) => resp || null),
      catchError(this.errorHandlerService.handleError("", 'getOrphanedPipelineServices: Error', null)),
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


