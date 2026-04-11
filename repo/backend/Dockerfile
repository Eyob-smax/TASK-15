# Stage 1: Build
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum* ./
# Prime the module cache when a go.sum is already present, but do not rely on
# this step for final dependency resolution because the source tree is not yet copied.
RUN go mod download || true
COPY . .
# Build only the application binary in Docker. Avoid `go mod tidy` here because
# it can pull test-only modules and introduce flaky network failures in CI/builds.
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o /fitcommerce-api ./cmd/api

# Stage 2: Runtime
FROM alpine:3.20
# postgresql-client provides pg_dump, required for the backup service.
# This is the primary supported deployment path; pg_dump must be available
# in the runtime image. Local/host PATH usage is optional for development only.
RUN apk add --no-cache ca-certificates tzdata postgresql-client
WORKDIR /app
COPY --from=builder /fitcommerce-api .
COPY database/ ./database/

EXPOSE 8080
ENTRYPOINT ["./fitcommerce-api"]
