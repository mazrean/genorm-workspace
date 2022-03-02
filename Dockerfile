FROM golang:1.18rc1-alpine3.15

RUN apk add --update --no-cache git

ENV DOCKERIZE_VERSION v0.6.1
RUN go install github.com/jwilder/dockerize@$DOCKERIZE_VERSION

ENV AIR_VERSION v1.29.0
RUN go install github.com/cosmtrek/air@$AIR_VERSION

WORKDIR /app

COPY ./go.work go.work
COPY ./workspace/go.* ./workspace/
COPY ./genorm/go.* ./genorm/

WORKDIR /app/genorm

RUN go mod download
COPY ./genorm/ ./

RUN go build -buildvcs=false -o genorm ./cmd/

WORKDIR /app/workspace

RUN go mod download

COPY ./workspace/.air.toml ./.air.toml

ENTRYPOINT ["dockerize", "-wait", "tcp://mariadb:3306", "-timeout", "5m", "air"]
CMD ["-c", ".air.toml"]
