package com.example.todo;

import java.time.OffsetDateTime;
import java.util.UUID;

public record TodoResponse(
    UUID id,
    String title,
    String description,
    boolean completed,
    OffsetDateTime createdAt,
    OffsetDateTime updatedAt) {
}
