import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTable } from '@angular/material/table';
import { InfluxMeasurementsDataSource } from './influx-measurements-datasource';
import {CleanupService, InfluxDatabase, KubeService} from "../cleanup.service";
import {finalize} from "rxjs/operators";

@Component({
  selector: 'app-influx-measurements',
  templateUrl: './influx-measurements.component.html',
  styleUrls: ['./influx-measurements.component.css']
})
export class InfluxMeasurementsComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<InfluxDatabase>;
  dataSource: InfluxMeasurementsDataSource;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'databaseId'];

  constructor(private cService: CleanupService) {
    this.dataSource = new InfluxMeasurementsDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedInfluxMeasurements().pipe(
      finalize(() => this.dataSource.loadingSubject.next(false))
    ).subscribe((data: InfluxDatabase[] | null) => {
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
