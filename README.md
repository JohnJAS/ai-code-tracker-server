# AI Code Tracker Server

集中接收 `git-code-tracker` 在成功 Git push 后上传的 AI 代码统计记录。

## Run

以下命令均适用于 Git Bash。服务默认监听 `:8080`，并会在启动时创建 `repositories` 与 `commit_records` 表。

### Docker 启动本地开发版本

```bash
cp .env.example .env
# 编辑 .env，设置 MYSQL_PASSWORD 和 MYSQL_ROOT_PASSWORD
./scripts/build.sh
docker compose up -d
```

不带参数的 `build.sh` 使用当前本地工作区构建 `ai-code-tracker-server:latest`，包含未提交的改动。`docker compose up` 只运行镜像，不会重新构建服务镜像。

### Docker 启动指定版本

正式版本以 Git tag 为唯一来源。下面的命令会从 `v1.2.0` 对应的提交构建 `ai-code-tracker-server:v1.2.0`；即使当前本地代码不同，也不会影响该镜像：

```bash
./scripts/build.sh v1.2.0
APP_VERSION=v1.2.0 docker compose up -d
```

传入的版本必须是已有 Git tag；不存在时构建会失败。默认镜像名是 `ai-code-tracker-server`，可通过 `IMAGE_NAME` 覆盖，并在构建和启动时使用相同的值：

```bash
IMAGE_NAME=registry.example.com/team/ai-code-tracker-server ./scripts/build.sh v1.2.0
IMAGE_NAME=registry.example.com/team/ai-code-tracker-server APP_VERSION=v1.2.0 docker compose up -d
```

### MySQL 数据目录

MySQL 默认将数据持久化到仓库内的 `./data/mysql`，该目录已被 Git 忽略。`docker compose down` 不会删除数据；只有手动删除该目录才会清空数据库。可通过 `MYSQL_DATA_DIR` 挂载任意宿主机目录：

```bash
MYSQL_DATA_DIR=/d/docker-data/ai-code-tracker-mysql docker compose up -d
```

### 本机直接启动服务

本机需要安装 Go（本项目要求 Go 1.25.12）和 MySQL。请先创建 `tracker` 数据库及对应的 MySQL 用户，然后设置连接串并启动：

```bash
export MYSQL_DSN="tracker:your-password@tcp(127.0.0.1:3306)/tracker?parseTime=true"
go run ./cmd/server
```

也可以自行指定监听地址：

```bash
export LISTEN_ADDR=127.0.0.1:8080
export MYSQL_DSN="tracker:your-password@tcp(127.0.0.1:3306)/tracker?parseTime=true"
go run ./cmd/server
```

### Docker 启动 MySQL，本机启动服务

```bash
cp .env.example .env
# 编辑 .env，设置 MYSQL_PASSWORD 和 MYSQL_ROOT_PASSWORD
docker compose up -d mysql
export MYSQL_DSN="tracker:your-password@tcp(127.0.0.1:3306)/tracker?parseTime=true"
go run ./cmd/server
```

启动后可验证服务状态：

```bash
curl http://127.0.0.1:8080/healthz
```

## Configure Clients

在每个已安装 tracker 的仓库中，将 `.ai-tracking/config.json` 的 `uploadUrl` 设置为服务地址：

```json
{
  "uploadUrl": "http://tracker.internal:8080/v1/records"
}
```

空字符串表示禁用上传。客户端在 `post-push` 上传本次成功推送对应的记录；服务不可用时，记录保存在本地 outbox，并会在下次 push 时重试。

## Smoke Test

After the service is running, send a real test record and confirm it appears in the dashboard:

```bash
go run ./cmd/smoke-test
```

The default endpoint is `http://127.0.0.1:8080/v1/records`. Override it with `-url` or `AI_TRACKER_UPLOAD_URL`:

```bash
go run ./cmd/smoke-test -url http://tracker.internal:8080/v1/records
```

Each run generates a unique repository URL and commit SHA, so it persists one test record. The dashboard is available at `GET /` and its data API is `GET /v1/dashboard`.

## API

`POST /v1/records` 接收仓库 URL 与现有 CSV 行对应的 JSON 记录。服务使用规范化 `origin` URL 和 `commit_id` 去重，重复上传成功但不会重复计数。

当前版本不包含认证或 TLS 终止。仅应部署在受信任网络中，并由反向代理或网络策略限制访问范围。
