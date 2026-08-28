# Protocol Parser Server

Go 后端负责协议报文解析，并将每次成功解析的完整结果保存到 MySQL。

## 支持协议

- `2929`：通用头、位置上报及扩展字段解析。
- `JT-808`：帧转义与还原、XOR校验、通用消息头、分包头，以及通用应答、心跳、注册/鉴权报文内容、参数设置、终端控制、位置/批量位置、位置附加项和文本消息解析。
- `VDF`：普通HQ20终端上传帧、登录/上线/定位消息及GPS、状态、基站、WiFi、电量等附加信息解析。

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
ADMIN_USERNAME=admin
ADMIN_PASSWORD=请设置初始管理员密码
AUTH_SECRET=请设置至少32位随机签名密钥
AUTH_EXPIRE_HOURS=12
HTTP_ADDR=:8080
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174
```

然后运行 `go run .`。程序启动时会自动读取当前目录中的 `.env`。
真实 `.env` 已被 Git 忽略，不会提交数据库密码；系统环境变量优先级高于 `.env`。

首次启动时程序会自动：

1. 创建 `protocol_parser` 数据库（如不存在）。
2. 创建 `parse_history` 历史记录表和查询索引。
3. 创建 `users` 用户表；表为空时使用 `ADMIN_USERNAME`、`ADMIN_PASSWORD` 创建首个管理员。
4. 启动 HTTP 服务并监听 `8080` 端口。

密码使用带随机盐的 bcrypt 哈希保存，不使用明文或MD5。`ADMIN_PASSWORD` 只在首次创建用户时生效。

生产环境必须显式配置 `CORS_ALLOWED_ORIGINS`，多个可信前端地址使用英文逗号分隔；不要填写 `*`。服务会为请求返回请求标识和基础安全响应头，并在收到 Ctrl+C 或系统停止信号后等待在途请求完成再关闭数据库。

## 用户表设计

`users` 包含用户名、bcrypt密码哈希、显示名称、角色、邮箱、手机号、状态、最后登录时间及创建/更新时间。用户名具有唯一索引，密码哈希字段为 `VARCHAR(255)`，便于后续调整安全算法。

## 登录接口

```text
POST /api/auth/login       用户名密码登录
GET  /api/auth/me          获取当前用户
PUT  /api/auth/profile     修改显示名称、邮箱和手机号
POST /api/auth/logout      退出当前会话
```

除登录接口外，解析与历史记录接口均要求携带：

```text
Authorization: Bearer <token>
```

MySQL 用户需要具有创建数据库、创建表以及增删改查权限。

## 历史记录接口

```text
GET    /api/parser/history             分页与关键词搜索
GET    /api/parser/history/:id         查看完整记录
DELETE /api/parser/history/:id         删除单条记录
DELETE /api/parser/history             清空全部记录
POST   /api/parser/analyze             解析成功后自动保存
```
