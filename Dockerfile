# FROM gradle:4.7.0-jdk8-alpine AS build
# COPY --chown=gradle:gradle . /home/gradle/src
# WORKDIR /home/gradle/src
# RUN gradle build --no-daemon 

# FROM openjdk:8-jre-slim

# EXPOSE 8080

# RUN mkdir /app

# COPY --from=build /home/gradle/src/build/libs/*.jar /app/spring-boot-application.jar

# ENTRYPOINT ["java", "-XX:+UnlockExperimentalVMOptions", "-XX:+UseCGroupMemoryLimitForHeap", "-Djava.security.egd=file:/dev/./urandom","-jar","/app/spring-boot-application.jar"]

# Используем официальный образ Golang для сборки
FROM golang:1.23-alpine AS build

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы Go в контейнер
COPY . .

# Скачиваем зависимости и собираем бинарный файл
RUN go mod download
RUN go build -o main ./cmd/main.go

# Используем минимальный образ для запуска бинарника
FROM alpine:latest

# Открываем порт 8080 для прослушивания HTTP-запросов
EXPOSE 8080

# Копируем скомпилированный бинарный файл из предыдущего шага
COPY --from=build /app/main /app/main

# Указываем команду для запуска Go-приложения
ENTRYPOINT ["/app/main"]
