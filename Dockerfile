# Подготовка.
FROM golang:alpine AS preper
ARG NAME
WORKDIR /usr/src/${NAME}
# Создание директории конфигурации.
RUN mkdir /etc/${NAME}
# Копирование файлов.
COPY . .
COPY ./config/prod.yaml /etc/${NAME}/prod.yaml
RUN go mod download

# Тестирование.
FROM preper AS testing
ARG ACCESS_KEY
RUN go test --tags=unit -v ./...
RUN go test --tags=integration -v ./...
RUN go test --tags=e2e -v ./...

# Сборка.
FROM preper AS building
RUN go build -v -o /usr/local/bin/${NAME} ./cmd/main.go

# Релиз.
FROM alpine AS release
ARG NAME
# Настройки.
RUN apk add tzdata
RUN ln -s /usr/share/zoneinfo/Europe/Moscow /etc/localtime
# Создание директории журналов.
RUN mkdir /var/log/${NAME}
# Создание директории конфигурации.
RUN mkdir /etc/${NAME}
# Копирование файлов.
COPY ./config/prod.yaml /etc/${NAME}/prod.yaml
COPY --from=building /usr/local/bin/${NAME} /usr/local/bin/${NAME}
# Создание сертификатов.
RUN apk add openssl
RUN mkdir /etc/ssl/certs/${NAME}
RUN openssl req -x509 -out /etc/ssl/certs/${NAME}/cert.crt -keyout /etc/ssl/certs/${NAME}/private.key \
  	-newkey rsa:2048 -nodes -sha256 \
  	-subj '/CN=localhost' -extensions EXT -config <( \
   	printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
# Создание точки входа.
COPY ./docker-entrypoint.sh /usr/local/bin
RUN echo "exec ${NAME}" >> /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
EXPOSE 8000
ENTRYPOINT [ "docker-entrypoint.sh" ]
