# Build stage
FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn -B -q package -DskipTests

# Runtime stage
FROM eclipse-temurin:21-jre
WORKDIR /app

# OpenTelemetry Java Agent
ENV OTEL_AGENT_VERSION=2.6.0
RUN curl -sSL -o opentelemetry-javaagent.jar \
    https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v${OTEL_AGENT_VERSION}/opentelemetry-javaagent.jar

COPY --from=build /app/target/todo-api-spring-0.1.0.jar app.jar
EXPOSE 8080
CMD ["java", "-javaagent:/app/opentelemetry-javaagent.jar", "-jar", "/app/app.jar"]
