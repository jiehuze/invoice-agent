
```shell
# Dockerfile
FROM golang:1.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o ocr-service .

# 运行时镜像（基于 Ubuntu + OpenCV + Tesseract）
FROM ubuntu:22.04

# 安装系统依赖
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    tesseract-ocr \
    libtesseract-dev \
    libleptonica-dev \
    tesseract-ocr-chi-sim \
    tesseract-ocr-eng \
    libgtk-3-0 \
    libcanberra-gtk3-module \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# 安装 OpenCV（gocv 依赖）
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    libopencv-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制二进制文件
COPY --from=builder /app/ocr-service /ocr-service

EXPOSE 8080

CMD ["/ocr-service"]
```

## 本地运行
```shell
# 安装 OpenCV（Ubuntu 示例）
sudo apt install libopencv-dev

# 安装 Tesseract + 中文包
sudo apt install tesseract-ocr tesseract-ocr-chi-sim

# 运行服务
go run main.go
```

特性	说明
多语言支持	chi_sim+eng 同时识别中英文
货币符号识别	白名单包含 ¥$€£元人民币
图像预处理	灰度 + Otsu 二值化，大幅提升清晰度
错误回退	预处理失败时自动使用原图
安全上传	限制文件大小、验证图像格式
生产就绪	包含健康检查 /health 和 Docker 支持