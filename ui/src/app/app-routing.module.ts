import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import {PipesComponent} from "./modules/pipes/pipes.component";
import {HomeComponent} from "./modules/home/home.component";
import {AnalyticsWorkloadsComponent} from "./modules/analytics-workloads/analytics-workloads.component";
import {AnalyticsKubeServicesComponent} from "./modules/analytics-kube-services/analytics-kube-services.component";
import {KafkaTopicsComponent} from "./modules/kafka-topics/kafka-topics.component";

const routes: Routes = [
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: 'home', component: HomeComponent },
  { path: 'pipes', component: PipesComponent },
  { path: 'analytics-workloads', component: AnalyticsWorkloadsComponent },
  { path: 'analytics-services', component: AnalyticsKubeServicesComponent },
  { path: 'kafka-topics', component: KafkaTopicsComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
