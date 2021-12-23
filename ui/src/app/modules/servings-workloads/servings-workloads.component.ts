import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTable } from '@angular/material/table';
import { ServingsWorkloadsDataSource } from './servings-workloads-datasource';
import {CleanupService, KubeWorkload} from "../cleanup.service";
import {finalize} from "rxjs/operators";

@Component({
  selector: 'app-servings-workloads',
  templateUrl: './servings-workloads.component.html',
  styleUrls: ['./servings-workloads.component.css']
})
export class ServingsWorkloadsComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<KubeWorkload>;
  dataSource: ServingsWorkloadsDataSource;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'name'];

  constructor(private cService: CleanupService) {
    this.dataSource = new ServingsWorkloadsDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedServingWorkloads().pipe(
      finalize(() => this.dataSource.loadingSubject.next(false))
    ).subscribe((data: KubeWorkload[] | null) => {
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
