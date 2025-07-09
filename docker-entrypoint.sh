#!/bin/sh
set -e
# 输出当前环境变量，保证docker 将环境变量全部传递给容器
# 输出当前环境变量，保证docker将环境变量全部传递给容器
echo "------------------- ENV BEGIN -------------------"
env
echo "------------------- ENV END ---------------------"

# 1. 等待数据库服务启动 (用 root 用户)
echo "等待数据库连接..."
until mariadb -h"${DB_HOST}" -u"${DB_USER}" --password="${DB_PASSWORD}" --skip-ssl -e "SELECT 1;" > /dev/null 2>&1; do
    echo "等待 MariaDB 启动..."
    sleep 2
done
echo "数据库连接成功!"

# 2. 运行数据库迁移 (用业务用户)

echo "当前环境: $ENV"

echo "检查数据库初始化状态..."
MIGRATION_TABLE_COUNT=$(mariadb -h"${DB_HOST}" -u"${DB_USER}" --password="${DB_PASSWORD}" --skip-ssl "${DB_DBNAME}" -Ns -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '${DB_DBNAME}' AND table_name = 'migrations';")

if [ "$MIGRATION_TABLE_COUNT" -eq 0 ]; then
    echo "数据库未初始化或迁移表不存在，开始执行迁移..."
    go run cmd/tools/db/migrate.go -env="$ENV" -direction=up
    echo "数据库迁移完成!"
else
    echo "数据库已初始化，检查是否有新的迁移..."
    go run cmd/tools/db/migrate.go -env="$ENV" -direction=up
fi

# 3. 运行 API 资源发现工具，确保 Casbin 权限资源是最新的
echo "正在发现并注册 API 资源..."
# 注意: 此工具会自动加载 .env 文件和对应环境的配置，无需额外传递参数
go run cmd/tools/permission/discover.go
echo "API 资源注册完成!"

# 4. 启动主程序 (air or other commands)
# 使用 trap 捕获退出信号，以便优雅地关闭子进程
trap 'kill -TERM $child_pid 2>/dev/null' EXIT

echo "启动 CRM 应用 (air)..."
# 在后台运行 air，并将其 PID 存储在文件中
"$@" &
child_pid=$!
echo "$child_pid" > /tmp/air.pid

# 用一个无限循环代替 `wait` 命令，以确保容器本身不会因为 air 进程被终止而退出。
# 这使得 debug-toggle.sh 可以自由地管理 air 和 dlv 子进程。
echo "Entrypoint is now in a holding loop, allowing for process toggling."
while true; do
  sleep 86400
done 