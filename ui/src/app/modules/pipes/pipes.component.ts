import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import {MatTable, MatTableDataSource} from '@angular/material/table';
import { PipesDataSource } from './pipes-datasource';
import {finalize} from "rxjs/operators";
import {AnalyticsPipeline, CleanupService} from "../cleanup.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-pipes',
  templateUrl: './pipes.component.html',
  styleUrls: ['./pipes.component.css']
})
export class PipesComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<AnalyticsPipeline>;
  dataSource: MatTableDataSource<AnalyticsPipeline>;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'name','UserId', 'createdAt', 'updatedAt', 'actions'];

  constructor(private cService: CleanupService,
              private snackBar: MatSnackBar
  ) {
    this.dataSource = new MatTableDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedPipelineServices().
    subscribe((data: AnalyticsPipeline[] | null) => {
      if (data != null) {
        this.dataSource.data = data;
      }
      this.dataSource.sort = this.sort;
      this.dataSource.paginator = this.paginator;
      this.table.dataSource = this.dataSource
    });
  }

  deletePipe(item: AnalyticsPipeline) {
     this.cService.deletePipeline(item.id).subscribe(() => {
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

}
