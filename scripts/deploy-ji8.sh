#!/usr/bin/env bash
# ji8.ai 部署脚本：在本机（Git Bash）构建带 ji8 主题的 dujiao-next 全栈二进制并部署到 OVH。
# 本文件位于仓库 scripts/ 目录；D:\独角兽网页搭建\deploy-ji8.sh 是放在仓库外的同款副本。
# 用法：bash scripts/deploy-ji8.sh            # 构建 + 部署
#       bash scripts/deploy-ji8.sh build      # 只构建
#       bash scripts/deploy-ji8.sh deploy     # 只部署（使用上次构建产物）
set -Eeuo pipefail

REPO="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$REPO/build/dujiao-next-ji8"
SSH_TARGET="ovh"                      # ~/.ssh/config 里的 Host 别名（135.125.175.236:8422 ubuntu）
export COREPACK_ENABLE_DOWNLOAD_PROMPT=0
export COREPACK_HOME="$HOME/.cache/corepack"
PNPM="corepack pnpm@10.34.3"

build() {
  echo "==> [1/4] 构建 admin 前端（fullstack 模式）"
  (cd "$REPO/frontend/admin" && $PNPM install --frozen-lockfile && $PNPM exec vue-tsc -b && VITE_FULLSTACK=1 $PNPM exec vite build)
  echo "==> [2/4] 构建 user 前端（含 ji8 模板）"
  (cd "$REPO/frontend/user" && $PNPM install --frozen-lockfile && $PNPM run build)
  echo "==> [3/4] 嵌入前端产物"
  rm -rf "$REPO/internal/web/dist" && mkdir -p "$REPO/internal/web/dist"
  cp -r "$REPO/frontend/admin/dist" "$REPO/internal/web/dist/admin"
  cp -r "$REPO/frontend/user/dist" "$REPO/internal/web/dist/user"
  echo "==> [4/4] 交叉编译 linux/amd64"
  mkdir -p "$REPO/build"
  # 注意：不要加 BuildType=release，否则后台"一键升级"会用官方发行版覆盖掉带 ji8 主题的自编译二进制。
  VERSION_TAG="$(cd "$REPO" && git describe --tags --always 2>/dev/null || echo v1.4.8)-ji8"
  (cd "$REPO" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -tags release,fullstack \
    -ldflags="-s -w -X github.com/dujiao-next/internal/version.Version=${VERSION_TAG}" -o "$OUT" ./cmd/server)
  ls -la "$OUT"
}

deploy() {
  [[ -s "$OUT" ]] || { echo "找不到构建产物 $OUT，请先 build"; exit 1; }
  echo "==> 上传二进制"
  scp -q "$OUT" "$SSH_TARGET:/tmp/dujiao-next.new"
  echo "==> 服务器端替换并重启"
  ssh "$SSH_TARGET" 'bash -s' <<'REMOTE'
set -Eeuo pipefail
STAMP=$(date -u +%Y%m%d-%H%M%S)
sudo mkdir -p /var/backups/dujiao-next/binaries
sudo cp /opt/dujiao-next/dujiao-next "/var/backups/dujiao-next/binaries/dujiao-next.$STAMP"
sudo install -o dujiao -g dujiao -m 0750 /tmp/dujiao-next.new /opt/dujiao-next/dujiao-next.new
sudo mv -f /opt/dujiao-next/dujiao-next.new /opt/dujiao-next/dujiao-next
rm -f /tmp/dujiao-next.new
sudo systemctl restart dujiao-next
for i in $(seq 1 30); do
  if curl -fsS -m 3 http://127.0.0.1:8080/health >/dev/null 2>&1; then echo "本机健康检查通过"; break; fi
  sleep 1
  if [ "$i" = 30 ]; then echo "健康检查失败，回滚"; sudo cp "/var/backups/dujiao-next/binaries/dujiao-next.$STAMP" /opt/dujiao-next/dujiao-next; sudo systemctl restart dujiao-next; exit 1; fi
done
echo "storefront_template = $(curl -s http://127.0.0.1:8080/api/v1/public/config | grep -o '"storefront_template":"[a-z0-9]*"')"
echo "备份：/var/backups/dujiao-next/binaries/dujiao-next.$STAMP"
REMOTE
  echo "==> 公网检查"
  curl -s -m 15 -o /dev/null -w "https://ji8.ai/ -> %{http_code}\n" https://ji8.ai/
}

case "${1:-all}" in
  build) build ;;
  deploy) deploy ;;
  all) build; deploy ;;
  *) echo "用法: $0 [build|deploy]"; exit 1 ;;
esac
