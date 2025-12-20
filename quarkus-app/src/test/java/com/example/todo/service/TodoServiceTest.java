package com.example.todo.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import com.example.todo.dto.CreateTodoRequest;
import com.example.todo.dto.TodoResponse;
import com.example.todo.dto.UpdateTodoRequest;
import com.example.todo.model.TodoEntity;
import com.example.todo.repository.TodoRepository;
import io.quarkus.test.InjectMock;
import io.quarkus.test.junit.QuarkusTest;
import jakarta.inject.Inject;
import jakarta.ws.rs.NotFoundException;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import org.junit.jupiter.api.Test;

@QuarkusTest
class TodoServiceTest {

  @Inject TodoService service;

  @InjectMock TodoRepository repository;

  @Test
  void list_returnsResponsesInOrder() {
    TodoEntity first = newTodo("First");
    TodoEntity second = newTodo("Second");
    when(repository.listTodos(0, 2)).thenReturn(List.of(first, second));

    List<TodoResponse> result = service.list(0, 2);

    assertEquals(2, result.size());
    assertEquals(first.id, result.get(0).id());
    assertEquals(second.id, result.get(1).id());
  }

  @Test
  void get_throwsNotFoundWhenMissing() {
    UUID id = UUID.randomUUID();
    when(repository.findByIdOptional(id)).thenReturn(Optional.empty());

    assertThrows(NotFoundException.class, () -> service.get(id));
  }

  @Test
  void create_setsDefaultsAndPersists() {
    CreateTodoRequest request = new CreateTodoRequest("hello", null, null);

    TodoResponse created = service.create(request);

    verify(repository).persist(org.mockito.ArgumentMatchers.any(TodoEntity.class));
    assertEquals("hello", created.title());
    assertEquals(false, created.completed());
  }

  @Test
  void update_appliesProvidedFields() {
    UUID id = UUID.randomUUID();
    TodoEntity existing = newTodo("old");
    existing.id = id;
    when(repository.findByIdOptional(id)).thenReturn(Optional.of(existing));

    UpdateTodoRequest request = new UpdateTodoRequest("new title", "desc", true);

    TodoResponse updated = service.update(id, request);

    assertEquals("new title", updated.title());
    assertEquals("desc", updated.description());
    assertEquals(true, updated.completed());
  }

  @Test
  void delete_throwsNotFoundWhenMissing() {
    UUID id = UUID.randomUUID();
    when(repository.deleteById(id)).thenReturn(false);

    assertThrows(NotFoundException.class, () -> service.delete(id));
  }

  private TodoEntity newTodo(String title) {
    TodoEntity e = new TodoEntity();
    e.id = UUID.randomUUID();
    e.title = title;
    e.description = null;
    e.completed = false;
    e.createdAt = OffsetDateTime.now();
    e.updatedAt = OffsetDateTime.now();
    return e;
  }
}
