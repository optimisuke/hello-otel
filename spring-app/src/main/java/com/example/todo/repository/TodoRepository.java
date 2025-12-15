package com.example.todo.repository;

import com.example.todo.model.Todo;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface TodoRepository extends JpaRepository<Todo, UUID> {
}
