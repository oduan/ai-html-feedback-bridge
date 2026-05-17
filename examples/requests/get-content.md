# 获取已保存的用户补充内容

```bash
curl -i http://localhost:8080/interactions/demo-001
```

预期响应（假设之前保存了 JSON 内容）：

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"feedback":"需要补充项目排期","approved":true}
```
