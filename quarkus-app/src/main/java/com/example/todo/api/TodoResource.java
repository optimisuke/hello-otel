package com.example.todo.api;

import jakarta.inject.Inject;
import jakarta.validation.Valid;
import jakarta.ws.rs.Consumes;
import jakarta.ws.rs.DefaultValue;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.POST;
import jakarta.ws.rs.PUT;
import jakarta.ws.rs.DELETE;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.PathParam;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.QueryParam;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;
import com.example.todo.dto.CreateTodoRequest;
import com.example.todo.dto.TodoResponse;
import com.example.todo.dto.UpdateTodoRequest;
import com.example.todo.service.TodoService;
import java.util.Map;
import java.util.UUID;

@Path("/api/v1/todos")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class TodoResource {

  private static final int DEFAULT_LIMIT = 100;
  private static final int MAX_LIMIT = 500;

  @Inject TodoService service;

  @GET
  public Response list(
      @QueryParam("skip") @DefaultValue("0") int skip,
      @QueryParam("limit") @DefaultValue("" + DEFAULT_LIMIT) int limit) {
    if (skip < 0 || limit < 1 || limit > MAX_LIMIT) {
      return Response.status(Response.Status.BAD_REQUEST)
          .entity(Map.of("detail", "skip must be >= 0 and limit between 1 and " + MAX_LIMIT))
          .build();
    }
    return Response.ok(service.list(skip, limit)).build();
  }

  @GET
  @Path("/{id}")
  public TodoResponse get(@PathParam("id") UUID id) {
    return service.get(id);
  }

  @POST
  public Response create(@Valid CreateTodoRequest request) {
    TodoResponse created = service.create(request);
    return Response.status(Response.Status.CREATED).entity(created).build();
  }

  @PUT
  @Path("/{id}")
  public TodoResponse update(@PathParam("id") UUID id, @Valid UpdateTodoRequest request) {
    if (request.title() == null && request.description() == null && request.completed() == null) {
      throw new IllegalArgumentException("At least one field must be provided");
    }
    return service.update(id, request);
  }

  @DELETE
  @Path("/{id}")
  public Response delete(@PathParam("id") UUID id) {
    service.delete(id);
    return Response.noContent().build();
  }
}
