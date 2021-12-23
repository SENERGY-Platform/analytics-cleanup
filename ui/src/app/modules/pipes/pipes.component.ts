import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTable } from '@angular/material/table';
import { PipesDataSource } from './pipes-datasource';
import {finalize} from "rxjs/operators";
import {AnalyticsPipeline, CleanupService} from "../cleanup.service";

@Component({
  selector: 'app-pipes',
  templateUrl: './pipes.component.html',
  styleUrls: ['./pipes.component.css']
})
export class PipesComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<AnalyticsPipeline>;
  dataSource: PipesDataSource;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['id', 'name','UserId', 'createdAt', 'updatedAt'];

  constructor(private cService: CleanupService) {
    this.dataSource = new PipesDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedPipelineServices().pipe(
      finalize(() => this.dataSource.loadingSubject.next(false))
    ).subscribe((data: AnalyticsPipeline[] | null) => {
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
