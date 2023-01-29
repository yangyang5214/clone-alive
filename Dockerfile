FROM golang:1.19

RUN go env -w GOPROXY=https://goproxy.cn

RUN go install github.com/yangyang5214/clone-alive@latest

WORKDIR /usr/src/app

EXPOSE 8081

ENTRYPOINT clone-alive alive /usr/src/app





