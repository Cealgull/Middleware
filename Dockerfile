FROM golang:alpine as build
LABEL MAINTAINER toolmanp

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN GOPROXY=goproxy.cn go mod download 

COPY . /app/
RUN mkdir -p build
RUN GOPROXY=goproxy.cn go mod tidy
RUN GOPROXY=goproxy.cn go build --ldflags "-s -w" -o build/middleware
RUN go clean -modcache -cache


FROM alpine:latest

COPY --from=build /app/build/middleware /app/middleware
CMD ["/app/middleware"]
