package com.example.todo;

import io.quarkus.hibernate.orm.panache.PanacheEntityBase;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.OffsetDateTime;
import java.util.UUID;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

@Entity
@Table(name = "todos")
public class TodoEntity extends PanacheEntityBase {

  @Id
  @Column(name = "id", nullable = false, updatable = false)
  public UUID id;

  @Column(name = "title", length = 200, nullable = false)
  public String title;

  @Column(name = "description")
  public String description;

  @Column(name = "completed", nullable = false)
  public boolean completed = false;

  @CreationTimestamp
  @Column(
      name = "created_at",
      nullable = false,
      updatable = false,
      columnDefinition = "TIMESTAMP WITH TIME ZONE")
  public OffsetDateTime createdAt;

  @UpdateTimestamp
  @Column(
      name = "updated_at",
      nullable = false,
      columnDefinition = "TIMESTAMP WITH TIME ZONE")
  public OffsetDateTime updatedAt;
}
