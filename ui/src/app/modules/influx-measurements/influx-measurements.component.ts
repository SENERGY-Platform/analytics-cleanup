import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import {MatTable, MatTableDataSource} from '@angular/material/table';
import {CleanupService, InfluxDatabase} from "../cleanup.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-influx-measurements',
  templateUrl: './influx-measurements.component.html',
  styleUrls: ['./influx-measurements.component.css']
})
export class InfluxMeasurementsComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<InfluxDatabase>;
  dataSource: MatTableDataSource<InfluxDatabase>;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'databaseId', 'actions'];

  constructor(private cService: CleanupService,
              private snackBar: MatSnackBar) {
    this.dataSource = new MatTableDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedInfluxMeasurements()
      .subscribe((data: InfluxDatabase[] | null) => {
      if (data != null) {
        this.dataSource.data = data;
      }
      this.dataSource.sort = this.sort;
      this.dataSource.paginator = this.paginator;
      this.table.dataSource = this.dataSource
    });
  }

  deleteMeasurement(item: InfluxDatabase) {
    this.cService.deleteInfluxMeasurement(item.databaseId, item.id).subscribe(() => {
      const index = this.dataSource.data.indexOf(item);
      if (index > -1) {
        this.dataSource.data.splice(index, 1);
        this.dataSource._updateChangeSubscription();
      }
      this.snackBar.open(item.id + ' deleted', undefined, {
        duration: 2000,
      });
    });
  }

  deleteMeasurements(){
    this.cService.deleteInfluxMeasurements().subscribe(() => {
      this.dataSource.data = [];
      this.snackBar.open('Measurements deleted', undefined, {
        duration: 2000,
      });
    })
  }
}
