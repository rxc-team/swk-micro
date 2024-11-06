FROM alpine:3.2
COPY ./config.env /app/config.env
COPY ./db-config.json /app/db-config.json
ADD ./docker/global-srv /app
WORKDIR /app
ENTRYPOINT [ "/app/global-srv" ]
