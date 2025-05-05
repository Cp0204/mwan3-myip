# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置构建参数
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY . .

# 初始化 Go 模块
RUN go mod init mwan3-myip && go mod tidy

# 编译Go程序
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-amd64} \
    go build -o mwan3-myip

# 最终阶段
FROM scratch

# 复制编译好的二进制文件
COPY --from=builder /app/mwan3-myip /mwan3-myip
COPY --from=builder /app/public /public

# 设置工作目录
WORKDIR /

# 暴露端口
EXPOSE 800

# 运行程序
CMD ["/mwan3-myip"]