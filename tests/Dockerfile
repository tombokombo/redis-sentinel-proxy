FROM alpine:3.10.1

RUN apk --update --no-cache add redis

ADD ./configs /configs

RUN chmod -R 777 /configs

CMD ["/usr/bin/redis-server"]
