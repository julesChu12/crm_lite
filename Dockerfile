# Dockerfile for Production
FROM golang:1.24-alpine

# 安装 mysql-client 用于数据库连接检测
RUN apk add --no-cache mysql-client

WORKDIR /app

# 优化 Docker 缓存：先复制依赖文件并下载
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# 复制所有源码
COPY . .

# 编译应用
RUN go build -o /usr/local/bin/crm-lite ./cmd/server

# 复制 entrypoint 脚本并授权
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh && \
    sed -i 's/\r$//' /usr/local/bin/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["crm-lite", "start"]