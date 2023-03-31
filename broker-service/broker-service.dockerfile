# build a tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY BrokerApp /app

CMD ["/app/BrokerApp"]
