# AI HTML Feedback Bridge

一个极简后端服务，用来保存 AI 生成的 HTML 页面中，用户补充给 AI 的信息。

项目核心闭环：

```
AI 生成 HTML 页面 → 用户补充信息 → 后端保存 → 下一轮 AI 读取 → AI 继续处理
```

> **当前版本：v0.1** — 跑通最小闭环。

## 快速启动

```bash
docker compose up --build
```

服务将在 `http://localhost:8080` 启动。

## 接口

### 保存用户补充内容

```http
POST /interactions/{interaction_id}
```

请求体：任意文本内容（JSON、HTML、纯文本等）。

```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  --data '{"feedback":"需要补充项目排期"}' \
  http://localhost:8080/interactions/demo-001
```

### 获取用户补充内容

```http
GET /interactions/{interaction_id}
```

```bash
curl -i http://localhost:8080/interactions/demo-001
```

## 配置项

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | HTTP 服务监听端口 |
| `SQLITE_PATH` | `./data/app.db` | SQLite 数据库文件路径 |
| `MAX_CONTENT_SIZE` | `1048576` | 单次提交最大字节数（1MB） |

## 目录结构

```
backend/
  cmd/server/main.go          # 程序入口
  internal/config/config.go    # 配置读取
  internal/httpapi/            # HTTP 路由、处理器、错误响应、CORS
  internal/storage/            # SQLite 连接、迁移、数据操作
  migrations/                  # 数据库迁移 SQL
skills/
  html-feedback-bridge/        # Skill 初始目录（方向说明）
examples/requests/             # 示例 curl 请求
```

## 技术栈

- **语言**: Go
- **数据库**: SQLite
- **部署**: Docker + Docker Compose

## 文档

- [SPEC.md](./SPEC.md) — 功能规格
- [PLAN.md](./PLAN.md) — 开发计划

## 许可

MIT
