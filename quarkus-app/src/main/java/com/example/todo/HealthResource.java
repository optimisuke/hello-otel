package com.example.todo;

import io.smallrye.common.annotation.NonBlocking;
import jakarta.inject.Inject;
import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;
import java.util.Map;
import org.eclipse.microprofile.config.inject.ConfigProperty;

@Path("/health")
@Produces(MediaType.APPLICATION_JSON)
public class HealthResource {

  @Inject
  @ConfigProperty(name = "quarkus.application.name", defaultValue = "todo-api-quarkus")
  String serviceName;

  @GET
  @NonBlocking
  public Map<String, Object> health() {
    return Map.of("status", "healthy", "service", serviceName);
  }
}
