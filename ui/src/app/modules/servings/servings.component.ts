import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTable } from '@angular/material/table';
import { ServingsDataSource} from './servings-datasource';
import {AnalyticsPipeline, CleanupService, Export} from "../cleanup.service";
import {finalize} from "rxjs/operators";

@Component({
  selector: 'app-servings',
  templateUrl: './servings.component.html',
  styleUrls: ['./servings.component.css']
})
export class ServingsComponent implements AfterViewInit {
  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<Export>;
  dataSource: ServingsDataSource;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['ID', 'Name'];

  constructor(private cService: CleanupService) {
    this.dataSource = new ServingsDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedServingServices().pipe(
      finalize(() => this.dataSource.loadingSubject.next(false))
    ).subscribe((data: Export[] | null) => {
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
