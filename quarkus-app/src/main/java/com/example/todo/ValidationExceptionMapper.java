package com.example.todo;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.ConstraintViolationException;
import jakarta.ws.rs.core.Response;
import jakarta.ws.rs.ext.ExceptionMapper;
import jakarta.ws.rs.ext.Provider;
import java.util.List;
import java.util.Map;

@Provider
public class ValidationExceptionMapper
    implements ExceptionMapper<ConstraintViolationException> {

  @Override
  public Response toResponse(ConstraintViolationException exception) {
    List<Map<String, String>> details =
        exception.getConstraintViolations().stream()
            .map(this::toDetail)
            .toList();
    return Response.status(Response.Status.BAD_REQUEST)
        .entity(Map.of("detail", details))
        .build();
  }

  private Map<String, String> toDetail(ConstraintViolation<?> violation) {
    return Map.of(
        "field", violation.getPropertyPath().toString(),
        "message", violation.getMessage());
  }
}
