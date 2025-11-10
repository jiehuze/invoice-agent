#!/bin/bash
set -e

# 设置运行模式，默认为 dev
RUN_MODE=${RUN_MODE:-dev}

# 日志目录和文件（建议放在 /app/logs）
LOG_DIR="${PWD}/logs"
LOG_FILE="${LOG_DIR}/invoice-agent.${RUN_MODE}.log"

# 创建日志目录
mkdir -p "$LOG_DIR"

# 校验配置目录是否存在
#if [ ! -d "/app/conf/${RUN_MODE}" ]; then
#  echo "[$(date)] ERROR: config directory '/app/conf/${RUN_MODE}' not found!" >&2
#  exit 1
#fi

echo "[$(date)] Starting invoice-agent in mode: $RUN_MODE, logging to $LOG_FILE"

# 启动程序，并将 stdout + stderr 合并写入日志文件
# 使用 exec 保证 PID 1 是 invoice-agent（便于信号传递）
exec ./invoice-agent dev >> "$LOG_FILE" 2>&1