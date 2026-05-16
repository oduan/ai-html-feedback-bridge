# 保存纯文本内容

```bash
curl -i \
  -X POST \
  -H "Content-Type: text/plain; charset=utf-8" \
  --data '这是一个用户补充的纯文本反馈内容。' \
  http://localhost:8080/interactions/demo-002
```

预期响应：

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"ok":true,"interaction_id":"demo-002"}
```
