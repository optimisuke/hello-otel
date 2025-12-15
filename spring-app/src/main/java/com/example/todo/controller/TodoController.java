package com.example.todo.controller;

import com.example.todo.dto.TodoRequest;
import com.example.todo.dto.TodoResponse;
import com.example.todo.model.Todo;
import com.example.todo.repository.TodoRepository;
import jakarta.validation.Valid;
import java.util.List;
import java.util.UUID;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.server.ResponseStatusException;

@RestController
@RequestMapping("/api/v1/todos")
@Slf4j
public class TodoController {

    private final TodoRepository repository;

    public TodoController(TodoRepository repository) {
        this.repository = repository;
    }

    @GetMapping
    public List<TodoResponse> list() {
        List<TodoResponse> todos = repository.findAll().stream().map(this::toResponse).toList();
        log.info("todos_list count={}", todos.size());
        return todos;
    }

    @GetMapping("/{id}")
    public TodoResponse get(@PathVariable UUID id) {
        TodoResponse response = repository.findById(id)
                .map(this::toResponse)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "Todo not found"));
        log.info("todo_get id={}", id);
        return response;
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public TodoResponse create(@Valid @RequestBody TodoRequest request) {
        Todo todo = new Todo();
        todo.setTitle(request.title());
        todo.setDescription(request.description());
        todo.setCompleted(request.completed() != null && request.completed());
        Todo saved = repository.save(todo);
        log.info("todo_created id={} title={}", saved.getId(), saved.getTitle());
        return toResponse(saved);
    }

    @PutMapping("/{id}")
    public TodoResponse update(@PathVariable UUID id, @Valid @RequestBody TodoRequest request) {
        Todo todo = repository.findById(id)
                .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "Todo not found"));
        todo.setTitle(request.title());
        todo.setDescription(request.description());
        if (request.completed() != null) {
            todo.setCompleted(request.completed());
        }
        Todo saved = repository.save(todo);
        log.info("todo_updated id={} completed={}", saved.getId(), saved.isCompleted());
        return toResponse(saved);
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void delete(@PathVariable UUID id) {
        if (!repository.existsById(id)) {
            throw new ResponseStatusException(HttpStatus.NOT_FOUND, "Todo not found");
        }
        repository.deleteById(id);
        log.info("todo_deleted id={}", id);
    }

    private TodoResponse toResponse(Todo todo) {
        return new TodoResponse(
                todo.getId(),
                todo.getTitle(),
                todo.getDescription(),
                todo.isCompleted(),
                todo.getCreatedAt(),
                todo.getUpdatedAt()
        );
    }
}
