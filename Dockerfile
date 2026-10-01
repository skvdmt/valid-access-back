# Подготовка.
FROM golang:alpine AS preper
ARG NAME
WORKDIR /usr/src/${NAME}
# Создание директорий.
RUN mkdir /etc/${NAME}
RUN mkdir /var/log/${NAME}
# Копирование файлов.
COPY . .
COPY ./config/prod.yaml /etc/${NAME}/prod.yaml
RUN go mod download

# Тестирование.
FROM preper AS testing
RUN --mount=type=secret,id=access_key \
  export ACCESS_KEY=$(cat /run/secrets/access_key) && \
  go test --tags=unit -v ./... && \
  go test --tags=integration -v ./... && \
  go test --tags=e2e -v ./...

# Сборка.
FROM preper AS building
RUN go build -v -o /usr/local/bin/${NAME} ./cmd/main.go

# Релиз.
FROM alpine AS release
EXPOSE 8000
ARG NAME
# Настройки.
RUN apk add tzdata
RUN ln -s /usr/share/zoneinfo/Europe/Moscow /etc/localtime
# Создание директорий.
RUN mkdir /etc/${NAME}
RUN mkdir /var/log/${NAME}
# Копирование файлов.
COPY ./config/prod.yaml /etc/${NAME}/prod.yaml
COPY --from=building /usr/local/bin/${NAME} /usr/local/bin/${NAME}
# Создание точки входа.
COPY ./docker-entrypoint.sh /usr/local/bin
RUN echo "exec ${NAME}" >> /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT [ "docker-entrypoint.sh" ]
