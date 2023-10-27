FROM golang:1.20-alpine

ENV PACKAGE_PATH=moviedb-charge
RUN mkdir -p /go/src/
WORKDIR /go/src/$PACKAGE_PATH
COPY . /go/src/$PACKAGE_PATH/
RUN go mod download
RUN go build -o moviedb-charge

ARG GO_ENV
ENV GO_ENV $GO_ENV

ENV TZ=America/Fortaleza
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/$PACKAGE_PATH/
ENTRYPOINT ./moviedb-charge
