FROM golang:alpine
LABEL MAINTAINER toolmanp

WORKDIR /app
COPY . /app/

RUN mkdir -p build
RUN GOPROXY=goproxy.cn go mod tidy
RUN GOPROXY=goproxy.cn go build --ldflags "-s -w" -o build/middleware
RUN go clean -modcache -cache

CMD ["build/middleware"]
