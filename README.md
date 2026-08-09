# my_storage · 统一存储（PaaS）

GoFrame 对象存储编排：预签名上传/下载、元数据、配额、按站点隔离前缀。  
总后台 UI：`my_manage_backend` → `/paas/storage`；子站经 `app_key` 调用开放 API。

**不做**：自带运营后台（用总后台）；图片处理/CDN；ffmpeg。

## 端口

容器/进程 `:8005`；Docker 宿主机映射 **`8015:8005`**（避免与 `jh_game_service` 的 8005 冲突）。与 my_media `:8004`、my_manage `:8003` 区分。

## 快速开始

```bash
docker exec postgres psql -U postgres -c "CREATE DATABASE my_storage;" 2>/dev/null || true
make migrate
cp manifest/config/config.example.yaml manifest/config/config.yaml
go mod tidy && make build
./bin/storageapi
# curl http://127.0.0.1:8005/healthz   # 本地直跑
# Docker: curl http://127.0.0.1:8015/healthz
```

## Docker

```bash
make docker-up
make docker-rebuild
```

## API 摘要

| 门面 | Header | 能力 |
|------|--------|------|
| Admin | `X-Admin-Token` | 对象列表/上传凭证/确认/下载凭证/删除；clients 同步 |
| Open | `X-App-Key` + `X-App-Secret` | 本站对象上传/列表/下载/删除/用量 |

对象 key：`{site_code}/{biz}/{code}/{filename}`
