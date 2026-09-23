FROM golang:1.27.0-alpine AS development

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY src/ingest-service/go.mod ./

EXPOSE 8080

CMD ["air"]
