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

# 3. 启动主程序
echo "启动 CRM 应用..."
exec "$@" 