# 保存 JSON 内容

```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  --data '{"feedback":"需要补充项目排期","approved":true}' \
  http://localhost:8080/interactions/demo-001
```

预期响应：

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"ok":true,"interaction_id":"demo-001"}
```
