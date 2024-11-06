FROM alpine:3.9
# 设置时区
ENV TZ=Asia/Tokyo
# 设置程序源
RUN echo 'https://mirror.tuna.tsinghua.edu.cn/alpine/v3.9/main' >> /etc/apk/repositories
RUN echo 'https://mirror.tuna.tsinghua.edu.cn/alpine/v3.9/community' >> /etc/apk/repositories
# 安装程序
RUN apk add -U tzdata
RUN cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
RUN apk add --no-cache mongodb-tools
RUN apk add -U unzip
RUN apk add -U rclone
# 复制rcolne配置文件
COPY ./rclone.conf /root/.config/rclone/rclone.conf
# 复制资源文件
COPY ./assets /app/assets
# 复制i18n消息文件
COPY ./locales /app/locales
# 复制配置文件
COPY ./db-config.json /app/db-config.json
COPY ./config.env /app/config.env
# 复制消息文件，
# TODO 变更会取消该消息内容
COPY ./message-zh.json /app/message-zh.json
COPY ./message-en.json /app/message-en.json
COPY ./message-ja.json /app/message-ja.json

ADD ./docker/internal-api /app

WORKDIR /app

ENTRYPOINT [ "/app/internal-api" ]
