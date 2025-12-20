package com.example.todo.api;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.equalTo;
import static org.hamcrest.Matchers.hasSize;

import com.example.todo.dto.CreateTodoRequest;
import com.example.todo.dto.UpdateTodoRequest;
import com.example.todo.model.TodoEntity;
import com.example.todo.repository.TodoRepository;
import io.quarkus.test.InjectMock;
import io.quarkus.test.junit.QuarkusTest;
import io.restassured.http.ContentType;
import jakarta.ws.rs.NotFoundException;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

@QuarkusTest
class TodoResourceTest {

  @InjectMock TodoRepository repository;

  @Test
  void list_returnsOk() {
    TodoEntity todo = newTodo("list item");
    Mockito.when(repository.listTodos(0, 1)).thenReturn(List.of(todo));

    given()
        .when()
        .get("/api/v1/todos?skip=0&limit=1")
        .then()
        .statusCode(200)
        .body("$", hasSize(1))
        .body("[0].title", equalTo("list item"));
  }

  @Test
  void get_returnsNotFound() {
    UUID id = UUID.randomUUID();
    Mockito.when(repository.findByIdOptional(id)).thenReturn(Optional.empty());

    given()
        .when()
        .get("/api/v1/todos/" + id)
        .then()
        .statusCode(404);
  }

  @Test
  void create_validatesAndPersists() {
    Mockito.doAnswer(invocation -> null)
        .when(repository)
        .persist(Mockito.any(TodoEntity.class));

    CreateTodoRequest req = new CreateTodoRequest("title", "desc", null);

    given()
        .contentType(ContentType.JSON)
        .body(req)
        .when()
        .post("/api/v1/todos")
        .then()
        .statusCode(201)
        .body("title", equalTo("title"))
        .body("description", equalTo("desc"))
        .body("completed", equalTo(false));
  }

  @Test
  void update_throwsNotFound() {
    UUID id = UUID.randomUUID();
    Mockito.when(repository.findByIdOptional(id)).thenReturn(Optional.empty());

    UpdateTodoRequest req = new UpdateTodoRequest("new", null, null);

    given()
        .contentType(ContentType.JSON)
        .body(req)
        .when()
        .put("/api/v1/todos/" + id)
        .then()
        .statusCode(404);
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
