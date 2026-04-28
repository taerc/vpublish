# 多阶段构建 - 构建阶段
FROM docker.m.daocloud.io/library/golang:1.24-alpine AS builder

# 切换阿里云镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN GOPROXY=https://goproxy.cn,direct go mod download

# 复制源代码
COPY . .

# 设置构建参数（在 docker build 时通过 ARG 传入）
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

# 构建应用
RUN GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
    -X 'github.com/taerc/vpublish/internal/version.Version=${VERSION}' \
    -X 'github.com/taerc/vpublish/internal/version.GitCommit=${GIT_COMMIT}' \
    -X 'github.com/taerc/vpublish/internal/version.BuildTime=${BUILD_TIME}'" \
    -o vpublish-server ./cmd/server

# 运行阶段
FROM docker.m.daocloud.io/library/alpine:3.19

# 切换阿里云镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 创建非 root 用户
RUN addgroup -g 1000 vpublish && \
    adduser -D -u 1000 -G vpublish vpublish

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/vpublish-server .

# 复制配置文件模板
COPY --from=builder /build/configs/config.yaml.example ./configs/

# 复制数据库迁移文件
COPY --from=builder /build/migrations ./migrations/

# 创建 uploads 目录
RUN mkdir -p uploads && \
    chown -R vpublish:vpublish /app

# 切换到非 root 用户
USER vpublish

# 暴露端口
# 8080 - 主服务端口
# 8081 - MCP 服务端口
EXPOSE 8080 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./vpublish-server"]