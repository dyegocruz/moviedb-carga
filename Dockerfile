FROM golang:1.15 AS build_stage
ENV PACKAGE_PATH=moviedb-carga
RUN mkdir -p /go/src/
WORKDIR /go/src/$PACKAGE_PATH
COPY . /go/src/$PACKAGE_PATH/
RUN go mod download
RUN go build -o moviedb-carga
#=============================================================
#--------------------- final stage ---------------------------
#=============================================================
#FROM registry.unimedfortaleza.com.br/unimed-oracle-node/unimed-oracle-node:1.1 AS final_stage

ARG NODE_ENV
ENV GO_ENV $NODE_ENV

RUN echo "Oh dang look at that: ${GO_ENV}"
RUN echo "Oh dang look at that: ${NODE_ENV}"

ENV TZ=America/Fortaleza
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

#ENV PACKAGE_PATH=moviedb-carga
#COPY --from=build_stage /go/src/$PACKAGE_PATH/moviedb-carga /go/src/$PACKAGE_PATH/
#COPY --from=build_stage /go/src/$PACKAGE_PATH/$GO_ENV.env /go/src/$PACKAGE_PATH/
WORKDIR /go/src/$PACKAGE_PATH/
ENTRYPOINT ./moviedb-carga
EXPOSE 1323
