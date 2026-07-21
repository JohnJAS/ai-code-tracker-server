# AI Code Tracker Server

集中接收 `git-code-tracker` 在成功 Git push 后上传的 AI 代码统计记录。

## Run

1. 将 `.env.example` 复制为 `.env`，设置 MySQL 密码。
2. 运行 `docker compose up --build`。
3. 调用 `GET http://127.0.0.1:8080/healthz`，应返回 `{"status":"ok"}`。

服务启动时会创建 `repositories` 与 `commit_records` 表。

## Configure Clients

在每个已安装 tracker 的仓库中，将 `.ai-tracking/config.json` 的 `uploadUrl` 设置为服务地址：

```json
{
  "uploadUrl": "http://tracker.internal:8080/v1/records"
}
```

空字符串表示禁用上传。客户端在 `post-push` 上传本次成功推送对应的记录；服务不可用时，记录保存在本地 outbox，并会在下次 push 时重试。

## API

`POST /v1/records` 接收仓库 URL 与现有 CSV 行对应的 JSON 记录。服务使用规范化 `origin` URL 和 `commit_id` 去重，重复上传成功但不会重复计数。

当前版本不包含认证或 TLS 终止。仅应部署在受信任网络中，并由反向代理或网络策略限制访问范围。
