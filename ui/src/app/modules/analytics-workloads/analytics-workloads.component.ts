import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTable } from '@angular/material/table';
import { AnalyticsWorkloadsDataSource} from './analytics-workloads-datasource';
import {finalize} from "rxjs/operators";
import {CleanupService, KubeWorkload} from "../cleanup.service";

@Component({
  selector: 'app-analytics-workloads',
  templateUrl: './analytics-workloads.component.html',
  styleUrls: ['./analytics-workloads.component.css']
})
export class AnalyticsWorkloadsComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<KubeWorkload>;
  dataSource: AnalyticsWorkloadsDataSource;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'name'];

  constructor(private cService: CleanupService) {
    this.dataSource = new AnalyticsWorkloadsDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedAnalyticsWorkloads().pipe(
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
