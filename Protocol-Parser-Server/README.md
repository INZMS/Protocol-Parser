# Protocol Parser Server

Go 后端负责协议报文解析，并将每次成功解析的完整结果保存到 MySQL。

## 支持协议

- `2929`：通用头、位置上报及扩展字段解析。
- `JT-808`：帧转义与还原、XOR校验、通用消息头、分包头，以及通用应答、心跳、注册/鉴权报文内容、参数设置、终端控制、位置/批量位置、位置附加项和文本消息解析。

解析接口由前端明确传入协议名称，不实现设备注册管理、在线会话、TCP长连接或指令下发业务。

## MySQL 配置

复制示例配置并填写本机 MySQL 密码：

```cmd
copy .env.example .env
```

编辑 `.env`：

```dotenv
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=你的MySQL密码
MYSQL_DATABASE=protocol_parser
```

然后运行 `go run .`。程序启动时会自动读取当前目录中的 `.env`。
真实 `.env` 已被 Git 忽略，不会提交数据库密码；系统环境变量优先级高于 `.env`。

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
