# AI HTML Feedback Bridge 开发计划

本文档用于指导后续 AI 或开发者按步骤实现 v0.1。执行时必须逐步推进，每完成一个阶段，都要对照该阶段验收标准检查，不要跳步扩展。

规格来源：[SPEC.md](./SPEC.md)。

## 全局工程规则

1. 以 `SPEC.md` 为唯一功能规格来源。
2. v0.1 只实现用户补充内容的保存和读取，不增加用户系统、历史记录、后台页面、token、HTML 生成能力。
3. 优先测试先行。能先写测试的模块，先写失败测试，再实现功能。
4. 每完成一个阶段，必须运行对应测试或检查命令。
5. 所有 Go 代码必须通过 `gofmt`。
6. 所有数据库写入必须使用参数化 SQL。
7. 请求体大小限制必须在流式读取时进行。
8. 保存接口统一返回 `200 OK`。
9. 时间字段统一使用 UTC / RFC3339。
10. 不引入不必要的框架。优先使用 Go 标准库和一个 SQLite driver。

## v0.1 完成标准

v0.1 完成时必须满足：

1. 可以通过 `POST /interactions/{interaction_id}` 保存用户补充内容。
2. 可以通过 `GET /interactions/{interaction_id}` 原样读取用户补充内容。
3. 重复 `POST` 同一个 `interaction_id` 会覆盖旧内容。
4. 保存和读取时 `Content-Type` 保持一致。
5. 超过 `MAX_CONTENT_SIZE` 的请求返回 `413`。
6. 非法 `interaction_id` 返回 `400`。
7. 不存在的 `interaction_id` 返回 `404`。
8. CORS 允许任意来源。
9. 数据保存到 SQLite。
10. Docker Compose 可以启动服务。

## 阶段 0：项目骨架

目标：建立 Go 后端目录结构和最小可编译项目。

计划文件：

```text
backend/
  cmd/server/main.go
  internal/config/config.go
  internal/httpapi/router.go
  internal/httpapi/handlers.go
  internal/httpapi/errors.go
  internal/storage/sqlite.go
  internal/storage/interactions.go
  internal/storage/migrations.go
  migrations/001_create_interactions.sql
  go.mod
  go.sum
```

执行任务：

1. 创建 `backend/` 目录。
2. 初始化 Go module。
3. 建立空的 package 结构。
4. `main.go` 先能启动一个最小 HTTP server。
5. 不实现业务逻辑。

验收标准：

1. `cd backend && go test ./...` 可以运行。
2. `cd backend && go run ./cmd/server` 可以启动服务。
3. 目录结构和 `SPEC.md` 一致。
4. 不出现 `internal/submission` 包。

## 阶段 1：配置读取

目标：集中读取服务配置，并提供默认值。

配置项：

```text
PORT=8080
SQLITE_PATH=./data/app.db
MAX_CONTENT_SIZE=1048576
```

测试先行：

1. 先写 `internal/config` 的单元测试。
2. 测试默认值。
3. 测试环境变量覆盖。
4. 测试非法 `MAX_CONTENT_SIZE` 返回错误。

执行任务：

1. 实现 `config.Load()`。
2. 将 `PORT` 解析为服务监听端口。
3. 将 `MAX_CONTENT_SIZE` 解析为整数。
4. 明确错误信息，不 panic。

验收标准：

1. `go test ./internal/config` 通过。
2. 未设置环境变量时使用默认值。
3. 设置环境变量时正确覆盖。
4. 非法配置能返回可读错误。

## 阶段 2：SQLite 初始化与迁移

目标：服务启动时能创建 SQLite 数据库和 `interactions` 表。

数据表：

```text
interaction_id TEXT PRIMARY KEY
content_type   TEXT NOT NULL
content        TEXT NOT NULL
created_at     TEXT NOT NULL
updated_at     TEXT NOT NULL
```

测试先行：

1. 使用临时 SQLite 文件写 storage 测试。
2. 测试初始化后表存在。
3. 测试迁移可以重复执行。

执行任务：

1. 添加 SQLite driver。
2. 实现数据库打开逻辑。
3. 实现迁移执行逻辑。
4. 创建 `migrations/001_create_interactions.sql`。
5. 时间字段使用 TEXT 保存 RFC3339 UTC 字符串。

验收标准：

1. `go test ./internal/storage` 通过。
2. 临时数据库中存在 `interactions` 表。
3. 重复初始化不会失败。
4. 不把 `interaction_id` 当文件路径使用。

## 阶段 3：Storage 保存和读取

目标：实现交互内容的 upsert 和 get。

建议接口：

```go
type Interaction struct {
    InteractionID string
    ContentType   string
    Content       string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

测试先行：

1. 保存 JSON 内容后能读取原文。
2. 保存纯文本内容后能读取原文。
3. 重复保存同一个 `interaction_id` 会覆盖 `content` 和 `content_type`。
4. 重复保存时 `created_at` 不变，`updated_at` 更新。
5. 读取不存在的数据返回明确的 not found 错误。

执行任务：

1. 实现 `SaveInteraction`。
2. 实现 `GetInteraction`。
3. 使用参数化 SQL。
4. 所有时间使用 UTC。
5. 时间序列化为 RFC3339。

验收标准：

1. `go test ./internal/storage` 通过。
2. 内容可以原样保存和读取。
3. 覆盖保存行为符合规格。
4. not found 错误可被 HTTP 层映射为 `404`。

## 阶段 4：HTTP 路由和错误响应

目标：实现两个核心接口和统一错误响应。

接口：

```http
POST /interactions/{interaction_id}
GET  /interactions/{interaction_id}
```

测试先行：

1. 使用 `httptest` 写 handler 测试。
2. 测试合法 `POST` 返回 `200` 和 JSON。
3. 测试合法 `GET` 返回原始内容和保存时的 `Content-Type`。
4. 测试不存在返回 `404`。
5. 测试非法 `interaction_id` 返回 `400`。

执行任务：

1. 实现路由。
2. 实现 `interaction_id` 校验。
3. 实现 JSON 错误响应。
4. 默认 `Content-Type` 为 `text/plain; charset=utf-8`。
5. 服务端内部错误返回 `500`，不泄露内部路径或堆栈。

验收标准：

1. `go test ./internal/httpapi` 通过。
2. `POST /interactions/demo-001` 返回：

```json
{
  "ok": true,
  "interaction_id": "demo-001"
}
```

3. `GET /interactions/demo-001` 返回保存的原始内容。
4. 错误响应结构统一：

```json
{
  "ok": false,
  "error": "..."
}
```

## 阶段 5：请求体大小限制

目标：实现 `MAX_CONTENT_SIZE`，且必须在流式读取时限制。

测试先行：

1. 设置较小的 `MAX_CONTENT_SIZE`。
2. 发送刚好等于限制的内容，应保存成功。
3. 发送超过限制的内容，应返回 `413`。
4. 超限内容不能写入数据库。

执行任务：

1. 在读取 request body 时使用限制 reader。
2. 不允许完整读入内存后再判断大小。
3. 超限返回 `413 Payload Too Large`。
4. 错误信息建议为 `content too large`。

验收标准：

1. 相关 handler 测试通过。
2. 超限请求返回 `413`。
3. 超限请求不会污染已有数据。
4. `MAX_CONTENT_SIZE` 来自配置，不写死。

## 阶段 6：CORS 与 OPTIONS

目标：允许 AI 生成的 HTML 页面从任意来源提交内容。

确认规则：

1. `Access-Control-Allow-Origin: *`
2. 支持 `GET`、`POST`、`OPTIONS`
3. 允许 `Content-Type`
4. 不启用 cookie 或 credential 模式

测试先行：

1. `POST` 响应包含 CORS 头。
2. `GET` 响应包含 CORS 头。
3. `OPTIONS /interactions/{interaction_id}` 返回成功。
4. `OPTIONS` 响应包含允许方法和允许请求头。

执行任务：

1. 实现 CORS middleware。
2. 实现 OPTIONS 预检响应。
3. 不加入鉴权或 cookie 相关逻辑。

验收标准：

1. CORS handler 测试通过。
2. 浏览器页面可以跨源提交 `Content-Type: application/json` 请求。
3. 响应头不包含 `Access-Control-Allow-Credentials: true`。

## 阶段 7：端到端接口测试

目标：验证 HTTP 层、storage 和 SQLite 能完整协同。

测试先行：

1. 使用临时 SQLite 文件启动完整 router。
2. 通过 HTTP 保存内容。
3. 通过 HTTP 读取内容。
4. 验证覆盖保存。
5. 验证 404、400、413、CORS。

执行任务：

1. 添加集成测试。
2. 避免依赖固定本地端口。
3. 每个测试使用独立临时数据库。

验收标准：

1. `cd backend && go test ./...` 通过。
2. 测试覆盖核心闭环。
3. 测试不依赖外部服务。

## 阶段 8：Docker 与 Compose

目标：提供可部署的最小服务。

执行任务：

1. 编写根目录 `Dockerfile`。
2. 编写根目录 `docker-compose.yml`。
3. SQLite 数据目录挂载到容器外。
4. 环境变量与 `SPEC.md` 保持一致。
5. 如果需要，添加 `.dockerignore`。

建议 Compose：

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      PORT: "8080"
      SQLITE_PATH: "/data/app.db"
      MAX_CONTENT_SIZE: "1048576"
    volumes:
      - ./data:/data
```

验收标准：

1. `docker compose up --build` 可以启动服务。
2. 容器内服务监听 `8080`。
3. 数据库文件写入挂载目录。
4. 重启容器后数据仍可读取。

## 阶段 9：示例请求

目标：提供最小可复制的接口使用示例。

计划文件：

```text
examples/
  requests/
    save-json.md
    save-text.md
    get-content.md
```

执行任务：

1. 增加保存 JSON 的 curl 示例。
2. 增加保存纯文本的 curl 示例。
3. 增加读取内容的 curl 示例。
4. 示例使用 `POST /interactions/demo-001` 和 `GET /interactions/demo-001`。

验收标准：

1. 示例命令可以对本地 `localhost:8080` 服务执行。
2. 示例与真实接口一致。
3. 示例不引入额外业务字段要求。

## 阶段 10：Skill 初始目录

目标：创建 Skill 初始目录，只放方向说明。

计划文件：

```text
skills/
  html-feedback-bridge/
    SKILL.md
```

执行任务：

1. 创建 Skill 目录。
2. `SKILL.md` 只说明方向。
3. 不写具体 HTML 模板。
4. 不规定表单字段。
5. 不规定提交脚本实现。

验收标准：

1. `skills/html-feedback-bridge/SKILL.md` 存在。
2. 内容只描述方向：适合时输出 HTML，并可配合后端回传用户补充信息。
3. 没有写死 HTML 结构、表单字段、JS 代码或后端地址。

## 阶段 11：README

目标：提供轻量 README，让用户快速理解和启动项目。

执行任务：

1. 创建根目录 `README.md`。
2. 简要说明项目是什么。
3. 链接到 `SPEC.md` 和 `PLAN.md`。
4. 写清楚 Docker Compose 启动方式。
5. 写清楚两个接口。
6. 写清楚配置项。
7. 声明 Skill 当前只是初始目录。

验收标准：

1. README 不重复承载完整规格。
2. README 可以让新用户在几分钟内启动服务。
3. README 中的接口路径、配置项和 `SPEC.md` 一致。

## 阶段 12：最终验收

目标：确认 v0.1 符合规格，可以作为最小可用版本。

必须执行：

```text
cd backend
go test ./...
gofmt -w .
```

如果 Docker 可用，还必须执行：

```text
docker compose up --build
```

手动接口验收：

```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  --data '{"feedback":"需要补充项目排期","approved":true}' \
  http://localhost:8080/interactions/demo-001

curl -i http://localhost:8080/interactions/demo-001
```

最终验收标准：

1. 所有 Go 测试通过。
2. 所有 Go 文件已格式化。
3. Docker Compose 可以启动。
4. POST 保存成功返回 `200 OK`。
5. GET 能读取原始内容。
6. Content-Type 能保持一致。
7. 超限请求返回 `413`。
8. CORS 头正确。
9. SQLite 数据可持久化。
10. README、SPEC、PLAN 三份文档一致。

## 执行记录模板

后续 AI 每完成一个阶段，应在回复中按此格式汇报：

```text
完成阶段：阶段 N - 标题
改动文件：
- path/to/file

已运行检查：
- command

验收结果：
- [x] 标准 1
- [x] 标准 2

未完成或阻塞：
- 无
```

如果某一项验收无法完成，必须明确说明原因，不要继续推进下一阶段。
