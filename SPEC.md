# AI HTML Feedback Bridge

一个极简后端服务，用来保存 AI 生成的 HTML 页面中，用户补充给 AI 的信息。

项目第一版只验证一个闭环：

```text
AI 生成 HTML 页面
  -> 用户在页面中补充信息
  -> HTML 页面把补充内容提交到后端
  -> 后端按 interaction_id 保存
  -> 下一轮 AI 根据 interaction_id 读取
  -> AI 继续处理任务
```

这个项目不追求复杂功能。第一版的核心目标是让“AI 输出内容”从静态回答，向“可以收集用户补充信息的协作界面”迈出最小一步。

## 项目定位

当前 AI 的默认输出多为 Markdown。Markdown 稳定、易复制，但在复杂信息表达、任务确认、需求补充、审稿反馈等场景中，交互能力有限。

HTML 可以承载更高信息密度，也可以包含表单、按钮、输入框等交互入口。但真正值得探索的问题不是 HTML 是否更好看，而是：

> 用户在 AI 生成的 HTML 页面里补充的信息，能否被保存下来，并让下一轮 AI 自动读取？

本项目第一版只解决这个问题。

## 第一版范围

后端只做四件事：

1. 接收用户补充内容；
2. 按 `interaction_id` 保存内容；
3. 根据 `interaction_id` 返回之前保存的内容；
4. 对单次提交内容做大小限制。

后端明确不做：

1. 不生成 HTML；
2. 不保存 AI 生成的 HTML 页面；
3. 不理解 HTML 页面结构；
4. 不规定表单字段；
5. 不限制用户补充内容格式；
6. 不解析业务字段；
7. 不做 diff；
8. 不做版本管理；
9. 不做用户系统；
10. 不做后台管理页面；
11. 不规定 AI 应该如何生成 HTML。

后端只保存“用户补充给 AI 的原始内容”。

## 技术栈

第一版建议技术栈：

- 语言：Go
- 数据库：SQLite
- 部署：Docker + Docker Compose
- 配置：环境变量，后续可增加配置文件

选择 SQLite 的原因：

1. 第一版数据模型非常简单；
2. 不需要额外数据库服务；
3. 部署和开源体验更轻；
4. 更适合最小可用版本；
5. 后续如果出现多人协作、权限、历史记录或大规模数据需求，再考虑迁移到 PostgreSQL。

## 后端架构建议

为了方便后续 AI 或开发者分工协作，建议按清晰边界组织 Go 后端：

```text
backend/
  cmd/
    server/
      main.go
  internal/
    config/
      config.go
    httpapi/
      router.go
      handlers.go
      errors.go
    storage/
      sqlite.go
      interactions.go
      migrations.go
  migrations/
    001_create_interactions.sql
  go.mod
  go.sum
skills/
  html-feedback-bridge/
    SKILL.md
examples/
  requests/
  html/
docker-compose.yml
Dockerfile
README.md
LICENSE
```

模块职责建议：

- `cmd/server`：程序入口，只负责加载配置、初始化依赖、启动 HTTP 服务。
- `internal/config`：读取环境变量，提供默认值，集中管理配置。
- `internal/httpapi`：HTTP 路由、请求读取、`interaction_id` 校验、响应格式、错误映射。
- `internal/storage`：SQLite 连接、迁移、保存和读取交互内容的具体实现。
- `migrations`：数据库表结构变更，保持可追踪。
- `examples`：后续放示例 HTML 和 curl 请求。
- `skills`：Skill 初始目录，第一版只保留方向说明，不写死具体 HTML 生成规则。

这个拆法的目的，是让后续任务可以自然拆分：

- 一位 AI/开发者实现 HTTP 接口；
- 一位 AI/开发者实现 SQLite 存储；
- 一位 AI/开发者补测试；
- 一位 AI/开发者完善 Docker 和文档；
- 一位 AI/开发者后续设计 Skill 和示例页面。

架构收敛说明：

- v0.1 不单独拆 `internal/submission` 包；
- 当前业务语义很薄，HTTP 层和 SQLite 存储层已经足够清晰；
- `internal/httpapi` 负责协议细节、输入校验和错误响应；
- `internal/storage` 负责数据结构、建表、upsert 和查询；
- 后续如果出现历史记录、权限、任务归组或复杂业务规则，再拆出独立领域包。

## 数据模型

第一版只需要一张表：`interactions`。

字段建议：

```text
interaction_id TEXT PRIMARY KEY
content_type TEXT NOT NULL
content      TEXT NOT NULL
created_at   DATETIME NOT NULL
updated_at   DATETIME NOT NULL
```

语义说明：

- `interaction_id`：一次 AI HTML 页面收集用户补充信息的交互实例标识，由 AI 或 HTML 页面生成。
- `content_type`：保存提交时的 `Content-Type`，用于读取时原样返回。
- `content`：用户补充给 AI 的原始文本内容。
- `created_at`：首次保存时间，使用 UTC 时间格式。
- `updated_at`：最近一次覆盖保存时间，使用 UTC 时间格式。

同一个 `interaction_id` 再次提交时，覆盖之前内容，并更新 `updated_at`。

时间字段要求：

- 后端统一使用 UTC 时间；
- 使用 RFC3339 字符串保存和返回，例如 `2026-05-16T09:30:00Z`；
- 不依赖服务器本地时区。

命名说明：

- `interaction_id` 不表示用户登录会话，也不表示整段 AI 对话；
- 它表示一次具体的“AI HTML 页面 -> 用户补充信息 -> 后端保存”的交互；
- 一个任务、筛选流程或多轮对话中，可以产生多个 `interaction_id`；
- 后续如果需要把多个交互归属于同一个任务或对话，可以再增加 `task_id`、`conversation_id` 或 `turn_id` 等字段。

## 配置项

第一版至少需要以下配置：

```text
PORT=8080
SQLITE_PATH=./data/app.db
MAX_CONTENT_SIZE=1048576
```

说明：

- `PORT`：HTTP 服务端口，默认 `8080`。
- `SQLITE_PATH`：SQLite 数据库文件路径。
- `MAX_CONTENT_SIZE`：单次提交最大字节数，默认 `1048576`，即 1MB。

`MAX_CONTENT_SIZE` 不应写死在代码中。超过限制时，应返回明确错误。

## HTTP 接口

第一版只提供两个核心接口。

### 提交用户补充内容

```http
POST /interactions/{interaction_id}
```

作用：提交并保存某个 `interaction_id` 下，用户补充给 AI 的内容。

请求体：任意文本内容，可以是 JSON、HTML、XML、Markdown、纯文本或其他文本格式。

请求头示例：

```http
Content-Type: application/json
```

或：

```http
Content-Type: text/html
```

后端行为：

1. 从路径读取 `interaction_id`；
2. 以流式限制方式读取请求体；
3. 按 `MAX_CONTENT_SIZE` 在读取过程中检查大小；
4. 保存请求体原文和 `Content-Type`；
5. 如果同一个 `interaction_id` 已存在，覆盖之前内容；
6. 返回保存结果。

成功响应：

```json
{
  "ok": true,
  "interaction_id": "demo-001"
}
```

建议错误响应：

```json
{
  "ok": false,
  "error": "content too large"
}
```

接口命名说明：

- URL 不再追加 `/submission`；
- `interaction_id` 已经唯一指向一次 HTML 交互实例；
- 该资源当前保存的主体就是用户补充内容；
- 如果后续一个 interaction 下需要挂载多个子资源，再考虑增加子路径；
- 主接口使用 `POST`，和 HTML 页面提交表单或脚本提交内容的语义保持一致；
- 同一个 `interaction_id` 重复 `POST` 时，后端执行覆盖保存。

### 获取用户补充内容

```http
GET /interactions/{interaction_id}
```

作用：根据 `interaction_id` 获取用户之前补充给 AI 的原始内容。

后端行为：

1. 从路径读取 `interaction_id`；
2. 查询保存的内容；
3. 如果存在，响应体返回原始内容；
4. 响应头使用保存时的 `Content-Type`；
5. 如果不存在，返回 `404`。

成功响应示例：

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"feedback":"需要补充项目排期","approved":true}
```

不存在时的响应示例：

```json
{
  "ok": false,
  "error": "interaction not found"
}
```

## HTTP 状态码

```text
200 OK                  获取成功；提交保存成功
400 Bad Request         interaction_id 为空或格式非法
404 Not Found           interaction 不存在
413 Payload Too Large   请求体超过 MAX_CONTENT_SIZE
500 Internal Server Error 服务端内部错误
```

第一版不区分首次保存和覆盖保存，保存接口统一返回 `200 OK`。

## interaction_id 建议

第一版不需要用户系统，`interaction_id` 是唯一定位一次用户补充内容提交目标的 key。

它的语义不是“会话”，而是“一次 HTML 交互实例”。例如同一个多轮对话里，第一轮生成了一个需求确认 HTML，第三轮又生成了一个排期确认 HTML，这两个 HTML 应该使用两个不同的 `interaction_id`。

建议约束：

- 长度限制，例如 `1-128` 个字符；
- 允许字母、数字、下划线、短横线和点；
- 不把 `interaction_id` 当文件路径使用；
- 所有数据库操作使用参数化 SQL。

示例：

```text
demo-001
task-2026-001.r1.html1
task-2026-001.r3.timeline
review.ab12cd
```

## 内容格式策略

后端不规定 `content` 必须是什么格式。

允许内容包括：

- JSON；
- HTML 片段；
- XML；
- Markdown；
- 纯文本；
- 其他文本格式。

保存时记录提交的 `Content-Type`。读取时用相同 `Content-Type` 返回，便于下一轮 AI 或外部工具按原格式处理。

如果请求没有提供 `Content-Type`，建议默认保存为：

```text
text/plain; charset=utf-8
```

## 安全与边界

第一版不做登录鉴权，因此默认适合本地、内网、Demo 或受控环境。

需要注意：

1. 必须以流式读取方式限制请求体大小；
2. 不执行用户提交内容；
3. 不把用户提交内容拼接进 SQL；
4. 不根据 `interaction_id` 读写任意文件；
5. 错误响应不要泄露数据库路径或内部堆栈；
6. 如果公开部署，后续应增加 token 或其他访问控制。

已确认的 v0.1 实现规则：

1. CORS 允许任意来源；
2. 返回 `Access-Control-Allow-Origin: *`；
3. 支持 `GET`、`POST` 和 `OPTIONS`；
4. 允许 `Content-Type` 请求头；
5. 不启用 cookie 或 credential 模式；
6. 请求体大小限制必须在流式读取时进行，不允许完整读入内存后再判断；
7. 保存状态统一返回 `200 OK`；
8. 时间字段统一使用 UTC 时间格式。

## Docker 部署建议

第一版建议提供：

- `Dockerfile`：构建 Go 服务；
- `docker-compose.yml`：挂载 SQLite 数据目录并启动服务；
- `.env.example`：展示可配置项。

Compose 形态建议：

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

## 开发和测试建议

后续实现时建议优先覆盖以下测试：

1. 保存 JSON 内容后能原样读取；
2. 保存纯文本内容后能原样读取；
3. 同一个 `interaction_id` 重复提交会覆盖旧内容；
4. 读取不存在的 `interaction_id` 返回 `404`；
5. 超过 `MAX_CONTENT_SIZE` 返回 `413`；
6. 保存和读取时 `Content-Type` 保持一致；
7. 非法 `interaction_id` 返回 `400`；
8. SQLite 初始化时自动创建表。
9. CORS 返回任意来源允许头；
10. 保存成功统一返回 `200 OK`；
11. `created_at` 和 `updated_at` 使用 UTC 时间。

## Skill 目录定位

本项目会包含一个 Skill，但第一版不设计完整规则。

Skill 的方向是：

> 让 AI 在适合的场景下尽量以 HTML 形式输出内容，并在需要用户补充信息时，能配合后端服务完成信息回传。

第一版 `skills/html-feedback-bridge/SKILL.md` 只放方向说明，不放具体实现规则。

第一版 Skill 不规定：

1. HTML 必须长什么样；
2. 页面必须包含哪些表单；
3. 用户补充信息必须是什么格式；
4. AI 必须如何生成提交逻辑；
5. AI 必须如何组织页面结构。

这些内容留到后续真实使用后再设计。

## Roadmap

### v0.1

目标：跑通最小闭环。

- Go 后端；
- SQLite；
- Docker / Docker Compose；
- 保存用户补充内容接口；
- 获取用户补充内容接口；
- 内容大小限制；
- README；
- Skill 初始目录。

### v0.2

可选增强：

- 更完整的 Skill；
- 示例 HTML；
- 示例请求；
- 简单 token；
- 历史记录。

### v0.3

可选增强：

- Web UI 查看提交内容；
- 更多示例场景；
- 更稳定的前端提交脚本；
- 根据实际使用决定是否迁移到 PostgreSQL。

## 开发优先级

建议后续按以下顺序推进：

1. 建立 Go 项目骨架和配置读取；
2. 建立 SQLite 连接和表迁移；
3. 实现 SQLite storage 的保存和读取方法；
4. 实现两个 HTTP handlers；
5. 增加大小限制和错误响应；
6. 补充单元测试和接口测试；
7. 增加 Dockerfile 和 docker-compose；
8. 添加 Skill 初始目录；
9. 增加示例请求和示例 HTML。

第一版完成标准：

```text
可以通过 POST 保存某个 interaction_id 的用户补充内容，
可以通过 GET 原样读取该内容，
并且超过大小限制时会被拒绝。
```
