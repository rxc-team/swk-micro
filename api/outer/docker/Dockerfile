FROM alpine:3.2
COPY ./db-config.json /app/db-config.json
COPY ./config.env /app/config.env
ADD ./docker/outer-api /app
WORKDIR /app
ENTRYPOINT [ "/app/outer-api" ]
