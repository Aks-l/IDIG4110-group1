FROM golang:1.27.1-alpine AS development

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY src/auth-service/go.mod ./

EXPOSE 8080

CMD ["air"]
