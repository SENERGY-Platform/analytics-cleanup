import {APP_INITIALIZER, NgModule} from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { LayoutModule } from '@angular/cdk/layout';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatTableModule } from '@angular/material/table';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSortModule } from '@angular/material/sort';
import {CoreModule} from "./core/core.module";
import {HomeComponent} from "./modules/home/home.component";
import { PipesComponent } from './modules/pipes/pipes.component';
import {HttpClientModule} from "@angular/common/http";
import {MatProgressSpinnerModule} from "@angular/material/progress-spinner";
import { ServingsComponent } from './modules/servings/servings.component';
import { AnalyticsWorkloadsComponent } from './modules/analytics-workloads/analytics-workloads.component';
import { ServingsWorkloadsComponent } from './modules/servings-workloads/servings-workloads.component';
import {KeycloakAngularModule, KeycloakService} from "keycloak-angular";
import {environment} from "../environments/environment";
import { ServingsKubeServicesComponent } from './modules/servings-kube-services/servings-kube-services.component';
import { AnalyticsKubeServicesComponent } from './modules/analytics-kube-services/analytics-kube-services.component';
import {InfluxMeasurementsComponent} from "./modules/influx-measurements/influx-measurements.component";

function initializeKeycloak(keycloak: KeycloakService) {
  return () =>
    keycloak.init({
      config: {
        url: environment.keycloak.url,
        realm: environment.keycloak.realm,
        clientId: environment.keycloak.clientId
      },
      initOptions: {
        onLoad: 'login-required',
        checkLoginIframe: false,
      }
    });
}

@NgModule({
  declarations: [
    AppComponent,
    HomeComponent,
    PipesComponent,
    ServingsComponent,
    AnalyticsWorkloadsComponent,
    ServingsWorkloadsComponent,
    ServingsKubeServicesComponent,
    AnalyticsKubeServicesComponent,
    InfluxMeasurementsComponent
  ],
  imports: [
    KeycloakAngularModule,
    BrowserModule,
    HttpClientModule,
    CoreModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    LayoutModule,
    MatToolbarModule,
    MatButtonModule,
    MatSidenavModule,
    MatIconModule,
    MatListModule,
    MatTableModule,
    MatPaginatorModule,
    MatSortModule,
    MatProgressSpinnerModule
  ],
  providers: [
    {
      provide: APP_INITIALIZER,
      useFactory: initializeKeycloak,
      multi: true,
      deps: [KeycloakService]
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
