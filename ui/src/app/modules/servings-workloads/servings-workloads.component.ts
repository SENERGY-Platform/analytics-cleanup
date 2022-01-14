import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import {MatTable, MatTableDataSource} from '@angular/material/table';
import {AnalyticsPipeline, CleanupService, KubeWorkload} from "../cleanup.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-servings-workloads',
  templateUrl: './servings-workloads.component.html',
  styleUrls: ['./servings-workloads.component.css']
})
export class ServingsWorkloadsComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<KubeWorkload>;
  dataSource: MatTableDataSource<KubeWorkload>;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'name', 'actions'];

  constructor(private cService: CleanupService,
              private snackBar: MatSnackBar) {
    this.dataSource = new MatTableDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedServingWorkloads().
    subscribe((data: KubeWorkload[] | null) => {
      if (data != null) {
        this.dataSource.data = data;
      }
      this.dataSource.sort = this.sort;
      this.dataSource.paginator = this.paginator;
      this.table.dataSource = this.dataSource
    });
  }

  deleteWorkload(item: KubeWorkload) {
    this.cService.deleteServingWorkload(item.name).subscribe(() => {
      const index = this.dataSource.data.indexOf(item);
      if (index > -1) {
        this.dataSource.data.splice(index, 1);
        this.dataSource._updateChangeSubscription();
      }
      this.snackBar.open(item.name + ' deleted', undefined, {
        duration: 2000,
      });
    });
  }

  deleteWorkloads(){
    this.cService.deleteServingWorkloads().subscribe(() => {
      this.dataSource.data = [];
      this.snackBar.open('Serving Workloads deleted', undefined, {
        duration: 2000,
      });
    })
  }
}
