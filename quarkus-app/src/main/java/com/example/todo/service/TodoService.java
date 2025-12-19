package com.example.todo.service;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import jakarta.transaction.Transactional;
import jakarta.ws.rs.NotFoundException;
import com.example.todo.dto.CreateTodoRequest;
import com.example.todo.dto.TodoResponse;
import com.example.todo.dto.UpdateTodoRequest;
import com.example.todo.model.TodoEntity;
import com.example.todo.repository.TodoRepository;
import java.util.List;
import java.util.UUID;

@ApplicationScoped
public class TodoService {

  @Inject TodoRepository repository;

  public List<TodoResponse> list(int skip, int limit) {
    return repository.listTodos(skip, limit).stream().map(this::toResponse).toList();
  }

  public TodoResponse get(UUID id) {
    TodoEntity entity =
        repository
            .findByIdOptional(id)
            .orElseThrow(() -> new NotFoundException("Todo with id %s not found".formatted(id)));
    return toResponse(entity);
  }

  @Transactional
  public TodoResponse create(CreateTodoRequest request) {
    TodoEntity entity = new TodoEntity();
    entity.id = UUID.randomUUID();
    entity.title = request.title();
    entity.description = request.description();
    entity.completed = request.completed() != null ? request.completed() : false;
    repository.persist(entity);
    return toResponse(entity);
  }

  @Transactional
  public TodoResponse update(UUID id, UpdateTodoRequest request) {
    TodoEntity entity =
        repository
            .findByIdOptional(id)
            .orElseThrow(() -> new NotFoundException("Todo with id %s not found".formatted(id)));

    if (request.title() != null) {
      entity.title = request.title();
    }
    if (request.description() != null) {
      entity.description = request.description();
    }
    if (request.completed() != null) {
      entity.completed = request.completed();
    }
    return toResponse(entity);
  }

  @Transactional
  public void delete(UUID id) {
    boolean deleted = repository.deleteById(id);
    if (!deleted) {
      throw new NotFoundException("Todo with id %s not found".formatted(id));
    }
  }

  private TodoResponse toResponse(TodoEntity entity) {
    return new TodoResponse(
        entity.id,
        entity.title,
        entity.description,
        entity.completed,
        entity.createdAt,
        entity.updatedAt);
  }
}
