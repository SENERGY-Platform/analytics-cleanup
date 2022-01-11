import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import {PipesComponent} from "./modules/pipes/pipes.component";
import {HomeComponent} from "./modules/home/home.component";
import {ServingsComponent} from "./modules/servings/servings.component";
import {AnalyticsWorkloadsComponent} from "./modules/analytics-workloads/analytics-workloads.component";
import {ServingsWorkloadsComponent} from "./modules/servings-workloads/servings-workloads.component";
import {AnalyticsKubeServicesComponent} from "./modules/analytics-kube-services/analytics-kube-services.component";
import {ServingsKubeServicesComponent} from "./modules/servings-kube-services/servings-kube-services.component";
import {InfluxMeasurementsComponent} from "./modules/influx-measurements/influx-measurements.component";

const routes: Routes = [
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: 'home', component: HomeComponent },
  { path: 'pipes', component: PipesComponent },
  { path: 'servings', component: ServingsComponent },
  { path: 'analytics-workloads', component: AnalyticsWorkloadsComponent },
  { path: 'serving-workloads', component: ServingsWorkloadsComponent },
  { path: 'analytics-services', component: AnalyticsKubeServicesComponent },
  { path: 'serving-services', component: ServingsKubeServicesComponent },
  { path: 'influx-measurements', component: InfluxMeasurementsComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
