FROM golang:nanoserver

RUN go get github.com/gorilla/websocket

ENV APP_HOME /src
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"
EXPOSE 8010
CMD ["go", "run", "main.go"]
