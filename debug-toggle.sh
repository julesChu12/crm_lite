#!/usr/bin/env sh
set -ex # Exit on error, print commands

# debug-toggle.sh
# 用于在容器内在 air 热加载 与 delve 远程调试之间切换。

APP_DIR=/app
AIR_PID_FILE=/tmp/air.pid
DLV_PID_FILE=/tmp/dlv.pid

start_debug() {
  echo "[toggle] switching to delve debug mode"

  # 终止 air - 使用 pkill 更健壮
  pkill -f "[a]ir" || true

  # 启动 delve (headless exec mode)
  cd "$APP_DIR" || exit 1

  echo "[toggle] Compiling debug binary..."
  go build -o /app/main.bin .
  echo "[toggle] Build complete."

  # 通过 dlv exec 启动调试，并传递 start 参数
  dlv exec --headless --listen=:2345 --api-version=2 --accept-multiclient /app/main.bin -- start &
  echo $! > "$DLV_PID_FILE"
  echo "[toggle] dlv process started with PID $(cat $DLV_PID_FILE) with 'start' command"
}

stop_debug() {
  echo "[toggle] switching back to air hot reload"
  # 终止 dlv
  pkill -f "[d]lv exec" || true

  # 重启 air - 使用 nohup 在后台稳定运行
  cd "$APP_DIR" || exit 1
  echo "[toggle] starting air in background mode..."
  nohup air > /tmp/air.log 2>&1 &
  echo $! > "$AIR_PID_FILE"
  echo "[toggle] air process restarted with PID $(cat $AIR_PID_FILE)"
  echo "[toggle] air logs are being written to /tmp/air.log"
  echo "[toggle] stop_debug completed successfully"
}

case "$1" in
  start-debug)
    start_debug
    ;;
  stop-debug)
    stop_debug
    ;;
  *)
    echo "Usage: $0 {start-debug|stop-debug}"
    exit 1
    ;;

esac 