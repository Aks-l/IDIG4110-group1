# Build from the repository root:
#   docker build -f infra/prod/rules-engine.prod.Dockerfile .
FROM golang:1.27.1-alpine AS build

WORKDIR /src
# rules-engine's go.mod points at ../shared, so both keep their layout.
COPY src/shared/go.mod src/shared/go.sum ./shared/
COPY src/rules-engine/go.mod src/rules-engine/go.sum ./rules-engine/
WORKDIR /src/rules-engine
RUN go mod download

WORKDIR /src
COPY src/shared ./shared
COPY src/rules-engine ./rules-engine
WORKDIR /src/rules-engine
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/rules-engine ./cmd

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/rules-engine ./rules-engine
COPY src/rules-engine/migrations ./migrations

ENV MIGRATIONS_DIR=/app/migrations
EXPOSE 8080
ENTRYPOINT ["/app/rules-engine"]
