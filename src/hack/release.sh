#!/bin/sh

set -o errexit
set -o nounset

readonly BIN_PATH=bin

# 串行构建，支持在 CI 通过“分片”并行（多个 Runner/Job 共同完成）
# 可选环境变量：
#   SHARD_TOTAL 分片总数（默认 1）
#   SHARD_INDEX 当前分片索引（默认 0，范围 [0, SHARD_TOTAL)）
SHARD_TOTAL=${SHARD_TOTAL:-1}
SHARD_INDEX=${SHARD_INDEX:-0}
# 基本健壮性：非数字/越界回退到安全值
case "$SHARD_TOTAL" in ''|*[!0-9]*) SHARD_TOTAL=1;; esac
case "$SHARD_INDEX" in ''|*[!0-9]*) SHARD_INDEX=0;; esac
if [ "$SHARD_TOTAL" -lt 1 ]; then SHARD_TOTAL=1; fi
if [ "$SHARD_INDEX" -ge "$SHARD_TOTAL" ]; then SHARD_INDEX=$((SHARD_INDEX % SHARD_TOTAL)); fi

# 预热依赖缓存，减少后续下载
echo "[deps] Warming Go module cache..."
go mod download >/dev/null 2>&1 || true

# 生成目标列表并过滤不支持的目标
TARGETS=$(go tool dist list | awk '!/^(linux\/loong64|android\/|ios\/|js\/wasm)/')
echo "[shard] TOTAL=${SHARD_TOTAL} INDEX=${SHARD_INDEX}"

i=0
printf "%s\n" "$TARGETS" | while IFS= read -r dist; do
  mod=$(( i % SHARD_TOTAL ))
  i=$((i+1))
  # 非本分片的目标直接跳过
  if [ "$mod" -ne "$SHARD_INDEX" ]; then
    continue
  fi
  platform="${dist%/*}"
  arch="${dist#*/}"
  echo "[build] PLATFORM=$platform ARCH=$arch"
  make PLATFORM="$platform" ARCH="$arch" bililive
done

# 移除了打包循环，因为我们不再需要打包二进制文件
# 直接保留未打包的二进制文件在bin目录中
