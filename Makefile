# Переменные для проекта
APP_NAME = my-go-app
DOCKER_IMAGE = my-go-app
MAIN_FILE = ./cmd/main.go

# Создание и запуск приложения через Go
run:
	go run $(MAIN_FILE)

# Сборка бинарника
build:
	go build -o $(APP_NAME) $(MAIN_FILE)

# Сборка Docker-образа
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Запуск Docker-контейнера
docker-run:
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# Полный цикл: сборка и запуск Docker
docker: docker-build docker-run

# Удаление бинарника
clean:
	rm -f $(APP_NAME)
