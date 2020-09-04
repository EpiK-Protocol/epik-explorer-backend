FROM alpine:3.12.0

WORKDIR /app/
COPY ./app /app/app
COPY ./conf/config_dist.yml /app/conf/config.yml

EXPOSE 3002

ENTRYPOINT [ "/app/app" ]