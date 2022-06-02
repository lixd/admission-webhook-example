# 编译环境
FROM golang:1.17 as build
ENV GOPROXY=https://goproxy.cn GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /webhook
COPY . /webhook
# -ldflags="-s -w" 减小二进制文件体积 https://golang.org/cmd/link/#hdr-Command_Line
RUN go build -ldflags="-s -w" -o main

# 运行环境
FROM alpine:latest
WORKDIR /root
# 时区信息
COPY --from=build /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
# 二进制文件
COPY --from=build /webhook/main .
# 配置文件
COPY  /cert /cert/
ENTRYPOINT  ["./main"]
