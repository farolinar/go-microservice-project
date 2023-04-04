FROM alpine:latest

RUN mkdir /app

COPY LoggerApp /app

CMD ["/app/LoggerApp"]
