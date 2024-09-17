FROM golang:1.23-alpine AS build

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main ./cmd/main.go

FROM alpine:latest

EXPOSE 8080

COPY --from=build /app/main /app/main

ENTRYPOINT ["/app/main"]
