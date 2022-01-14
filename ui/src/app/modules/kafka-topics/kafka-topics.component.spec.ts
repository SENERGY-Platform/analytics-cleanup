import { ComponentFixture, TestBed } from '@angular/core/testing';

import { KafkaTopicsComponent } from './kafka-topics.component';

describe('KafkaTopicsComponent', () => {
  let component: KafkaTopicsComponent;
  let fixture: ComponentFixture<KafkaTopicsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ KafkaTopicsComponent ]
    })
    .compileComponents();
  });

  beforeEach(() => {
    fixture = TestBed.createComponent(KafkaTopicsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
