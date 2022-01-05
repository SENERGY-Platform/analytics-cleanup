import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTable } from '@angular/material/table';
import {CleanupService, KubeService} from "../cleanup.service";
import {KubeServicesDatasource} from "../kube-services-datasource";
import {finalize} from "rxjs/operators";

@Component({
  selector: 'app-servings-kube-services',
  templateUrl: './servings-kube-services.component.html',
  styleUrls: ['./servings-kube-services.component.css']
})
export class ServingsKubeServicesComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<KubeService>;
  dataSource: KubeServicesDatasource;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'name'];

  constructor(private cService: CleanupService) {
    this.dataSource = new KubeServicesDatasource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedServingKubeServices().pipe(
      finalize(() => this.dataSource.loadingSubject.next(false))
    ).subscribe((data: KubeService[] | null) => {
      this.dataSource.loadingSubject.next(true);
      if (data != null) {
        this.dataSource.data = data;
      }
      this.dataSource.sort = this.sort;
      this.dataSource.paginator = this.paginator;
      this.table.dataSource = this.dataSource
    });
  }
}
