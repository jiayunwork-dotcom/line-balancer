# line-balancer — Go 语言产线节拍平衡与瓶颈分析 HTTP 后端服务，支持多种启发式工位分配算法和效率指标计算

给定工序列表和产量需求，计算节拍时间 (takt)，用多种启发式算法将工序分配到工位，
报告瓶颈工位和产线效率。

## 构建 / 运行 / 测试

```text
go build ./...
go run . -addr :8080                     # 启动 HTTP 服务（/api/balance, /api/takt, /api/metrics）
go test ./...
```

## 评测镜像

本目录评测专用文件（勿覆盖项目自带 Dockerfile/README）：

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

两种架构都要构建并进容器验证：

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh line-balancer linux/arm64
./build_benzhi_docker.sh line-balancer linux/amd64
docker run -it line-balancer:latest
```
