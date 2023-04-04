FROM alpine:latest

RUN mkdir /app

COPY ListenerApp /app

CMD ["/app/ListenerApp"]
