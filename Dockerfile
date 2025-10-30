FROM golang:1.22-alpine AS build-env
ARG app
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on
WORKDIR /go/src
ADD ./ /go/src/$app
WORKDIR /go/src/$app
RUN /bin/sh build.sh $app

FROM golang:1.22-alpine
ARG app
ENV APP $app
WORKDIR /app
COPY --from=build-env  /go/src/$app/output /app
EXPOSE 8080
ENTRYPOINT /bin/sh run.sh