# Protocol Parser Server

Go 后端负责协议报文解析，并将每次成功解析的完整结果保存到 MySQL。

## MySQL 配置

程序默认连接 `127.0.0.1:3306`，使用 `root` 用户和空密码。数据库名默认为 `protocol_parser`。
建议在 PowerShell 中通过环境变量设置实际账号信息：

```powershell
$env:MYSQL_HOST = "127.0.0.1"
$env:MYSQL_PORT = "3306"
$env:MYSQL_USER = "root"
$env:MYSQL_PASSWORD = "你的MySQL密码"
$env:MYSQL_DATABASE = "protocol_parser"
go run .
```

首次启动时程序会自动：

1. 创建 `protocol_parser` 数据库（如不存在）。
2. 创建 `parse_history` 历史记录表和查询索引。
3. 启动 HTTP 服务并监听 `8080` 端口。

MySQL 用户需要具有创建数据库、创建表以及增删改查权限。

## 历史记录接口

```text
GET    /api/parser/history             分页与关键词搜索
GET    /api/parser/history/:id         查看完整记录
DELETE /api/parser/history/:id         删除单条记录
DELETE /api/parser/history             清空全部记录
POST   /api/parser/analyze             解析成功后自动保存
```

