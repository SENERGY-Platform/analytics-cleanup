import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import {MatTable, MatTableDataSource} from '@angular/material/table';
import {AnalyticsPipeline, CleanupService, KubeService} from "../cleanup.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-analytics-kube-services',
  templateUrl: './analytics-kube-services.component.html',
  styleUrls: ['./analytics-kube-services.component.css']
})
export class AnalyticsKubeServicesComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<KubeService>;
  dataSource: MatTableDataSource<KubeService>;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'name', 'actions'];

  constructor(private cService: CleanupService,
              private snackBar: MatSnackBar) {
    this.dataSource = new MatTableDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedPipelineKubeServices().
    subscribe((data: KubeService[] | null) => {
      if (data != null) {
        this.dataSource.data = data;
      }
      this.dataSource.sort = this.sort;
      this.dataSource.paginator = this.paginator;
      this.table.dataSource = this.dataSource
    });
  }

  deleteService(item: KubeService) {
    this.cService.deleteAnalyticsKubeService(item.id).subscribe(() => {
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

  deleteServices(){
    this.cService.deleteAnalyticsKubeServices().subscribe(() => {
      this.dataSource.data = [];
      this.snackBar.open('Analytics Kube Services deleted', undefined, {
        duration: 2000,
      });
    })
  }
}
