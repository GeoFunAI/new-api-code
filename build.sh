#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取版本号
VERSION=$(cat VERSION 2>/dev/null || echo "v0.0.0")

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Building New API${NC}"
echo -e "${GREEN}Version: ${VERSION}${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查必要的工具
echo -e "${YELLOW}Checking dependencies...${NC}"
command -v bun >/dev/null 2>&1 || { echo -e "${RED}Error: bun is required but not installed.${NC}" >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo -e "${RED}Error: go is required but not installed.${NC}" >&2; exit 1; }

# 设置 Go 环境变量
export GO111MODULE=on
export CGO_ENABLED=0
export GOPROXY=https://goproxy.cn,https://goproxy.io,direct

# 检测目标平台
TARGETOS=${TARGETOS:-$(go env GOOS)}
TARGETARCH=${TARGETARCH:-$(go env GOARCH)}
export GOOS=${TARGETOS}
export GOARCH=${TARGETARCH}

echo -e "${YELLOW}Target Platform: ${GOOS}/${GOARCH}${NC}"

# 步骤 1: 构建前端
echo -e "${GREEN}Step 1: Building frontend...${NC}"
cd web

# 检查是否存在 node_modules，如果不存在则安装依赖
if [ ! -d "node_modules" ] && [ ! -d ".bun" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    bun install
fi

# 构建前端
echo -e "${YELLOW}Building frontend with bun...${NC}"
DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=${VERSION} bun run build

if [ ! -d "dist" ]; then
    echo -e "${RED}Error: Frontend build failed, dist directory not found.${NC}"
    exit 1
fi

echo -e "${GREEN}Frontend build completed!${NC}"
cd ..

# 步骤 2: 下载 Go 依赖
echo -e "${GREEN}Step 2: Downloading Go dependencies...${NC}"
go mod download

# 步骤 3: 构建后端
echo -e "${GREEN}Step 3: Building backend...${NC}"
echo -e "${YELLOW}Building for ${GOOS}/${GOARCH}...${NC}"

# 构建标志
LDFLAGS="-s -w -X 'github.com/QuantumNous/new-api/common.Version=${VERSION}'"

# 构建可执行文件
OUTPUT_NAME="new-api"
if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME="new-api.exe"
fi

go build -ldflags "${LDFLAGS}" -o ${OUTPUT_NAME}

if [ ! -f "${OUTPUT_NAME}" ]; then
    echo -e "${RED}Error: Backend build failed, ${OUTPUT_NAME} not found.${NC}"
    exit 1
fi

echo -e "${GREEN}Backend build completed!${NC}"

# 显示构建结果
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}Output: ${OUTPUT_NAME}${NC}"
echo -e "${GREEN}Platform: ${GOOS}/${GOARCH}${NC}"
echo -e "${GREEN}Version: ${VERSION}${NC}"
echo -e "${GREEN}========================================${NC}"

# 显示文件信息
if command -v ls >/dev/null 2>&1; then
    echo -e "${YELLOW}File size:${NC}"
    ls -lh ${OUTPUT_NAME} | awk '{print $5}'
fi