FROM golang:1.20-alpine AS builder

# 移动到工作目录：/build
WORKDIR /build

COPY go.* .
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件 app
RUN go build -o app

###################
# 接下来创建一个小镜像
###################
FROM scratch

ENV GIN_MODE=release \
    TZ=Asia/Shanghai

COPY --from=builder /build/app .

# 需要运行的命令
ENTRYPOINT ["/app"]