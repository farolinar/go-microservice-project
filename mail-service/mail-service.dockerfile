FROM alpine:latest

RUN mkdir /app

COPY MailApp /app
COPY templates /templates

CMD ["/app/MailApp"]
