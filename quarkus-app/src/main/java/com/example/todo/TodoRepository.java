package com.example.todo;

import io.quarkus.hibernate.orm.panache.PanacheRepositoryBase;
import io.quarkus.panache.common.Sort;
import jakarta.enterprise.context.ApplicationScoped;
import java.util.List;
import java.util.UUID;

@ApplicationScoped
public class TodoRepository implements PanacheRepositoryBase<TodoEntity, UUID> {

  public List<TodoEntity> listTodos(int skip, int limit) {
    return findAll(Sort.descending("createdAt"))
        .range(skip, skip + limit - 1)
        .list();
  }
}
