FROM golang:1.18.1-bullseye

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /run-orch

CMD [ "/run-orch" ]