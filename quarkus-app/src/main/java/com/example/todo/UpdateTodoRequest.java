package com.example.todo;

import jakarta.validation.constraints.Size;

public record UpdateTodoRequest(
    @Size(max = 200) String title,
    String description,
    Boolean completed) {
}
