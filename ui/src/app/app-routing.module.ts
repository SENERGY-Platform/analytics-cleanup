import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import {PipesComponent} from "./modules/pipes/pipes.component";
import {HomeComponent} from "./modules/home/home.component";
import {ServingsComponent} from "./modules/servings/servings.component";
import {AnalyticsWorkloadsComponent} from "./modules/analytics-workloads/analytics-workloads.component";
import {ServingsWorkloadsComponent} from "./modules/servings-workloads/servings-workloads.component";

const routes: Routes = [
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: 'home', component: HomeComponent },
  { path: 'pipes', component: PipesComponent },
  { path: 'servings', component: ServingsComponent },
  { path: 'analytics-workloads', component: AnalyticsWorkloadsComponent },
  { path: 'serving-workloads', component: ServingsWorkloadsComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
