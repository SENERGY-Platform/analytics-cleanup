import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import {PipesComponent} from "./modules/pipes/pipes.component";
import {HomeComponent} from "./modules/home/home.component";

const routes: Routes = [
  { path: '', redirectTo: 'home', pathMatch: 'full' },
  { path: 'home', component: HomeComponent },
  { path: 'pipes', component: PipesComponent },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
