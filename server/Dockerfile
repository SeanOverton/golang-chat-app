FROM golang:1.23.3

ENV APP_HOME /src
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"

ADD src "$APP_HOME"

RUN go get github.com/gorilla/websocket

EXPOSE 8080
CMD ["go", "run", "main.go"]
