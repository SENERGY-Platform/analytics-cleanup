import {AfterViewInit, Component, ViewChild} from '@angular/core';
import {MatPaginator} from "@angular/material/paginator";
import {MatSort} from "@angular/material/sort";
import {MatTable, MatTableDataSource} from "@angular/material/table";
import {CleanupService} from "../cleanup.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-kafka-topics',
  templateUrl: './kafka-topics.component.html',
  styleUrls: ['./kafka-topics.component.css']
})
export class KafkaTopicsComponent implements AfterViewInit {

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;
  @ViewChild(MatTable) table!: MatTable<string>;
  dataSource: MatTableDataSource<string>;

  /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
  displayedColumns = ['name', 'actions'];

  constructor(private cService: CleanupService,
              private snackBar: MatSnackBar) {
    this.dataSource = new MatTableDataSource();
  }

  ngAfterViewInit(): void {
    this.cService.getOrphanedKafkaTopics().
    subscribe((data: string[] | null) => {
      if (data != null) {
        this.dataSource.data = data;
      }
      this.dataSource.sort = this.sort;
      this.dataSource.paginator = this.paginator;
      this.table.dataSource = this.dataSource
    });
  }

  deleteTopic(item: string) {
    this.cService.deleteKafkaTopic(item).subscribe(() => {
      const index = this.dataSource.data.indexOf(item);
      if (index > -1) {
        this.dataSource.data.splice(index, 1);
        this.dataSource._updateChangeSubscription();
      }
      this.snackBar.open(item + ' deleted', undefined, {
        duration: 2000,
      });
    });
  }

  deleteTopics(){
    this.cService.deleteKafkaTopics().subscribe(() => {
      this.dataSource.data = [];
      this.snackBar.open('KafkaTopics deleted', undefined, {
        duration: 2000,
      });
    })
  }

}
