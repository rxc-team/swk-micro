FROM alpine:3.9
# 设置时区
ENV TZ=Asia/Tokyo
# 设置程序源
RUN echo 'https://mirror.tuna.tsinghua.edu.cn/alpine/v3.9/main' >> /etc/apk/repositories
RUN echo 'https://mirror.tuna.tsinghua.edu.cn/alpine/v3.9/community' >> /etc/apk/repositories
# 安装程序
RUN apk add -U tzdata
RUN cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
# 复制配置文件
COPY ./db-config.json /app/db-config.json
COPY ./config.env /app/config.env

ADD ./docker/system-api /app

WORKDIR /app

ENTRYPOINT [ "/app/system-api" ]
