package com.example.todo.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

public record TodoRequest(
        @NotBlank @Size(max = 200) String title,
        @Size(max = 2000) String description,
        Boolean completed
) {
}
