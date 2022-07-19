FROM golang:1.18 AS build_stage

ENV PACKAGE_PATH=moviedb-carga
RUN mkdir -p /go/src/
WORKDIR /go/src/$PACKAGE_PATH
COPY . /go/src/$PACKAGE_PATH/
RUN go mod download
RUN go build -o moviedb-carga

ARG GO_ENV
ENV GO_ENV $GO_ENV

ENV TZ=America/Fortaleza
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/$PACKAGE_PATH/
ENTRYPOINT ./moviedb-carga
