# GPT Team

`gpt-team` 是一个前后端分离的内部管理后台，主要用于维护账号、邮箱后缀、卡密和系统用户。仓库当前分为两个独立子项目：

- `frontend/`：管理后台前端，基于 Bun + Turbo + React + Vite。
- `go/`：后端 API，基于 Go + Gin + GORM + PostgreSQL。

## 功能概览

- 账号管理：维护 Plus / Business 账号，支持状态筛选、质保账号、查看邮箱内容等。
- 邮箱管理：维护 Cloudmail / DuckMail 邮箱后缀和访问凭据。
- 卡密管理：支持批量导入卡密、激活、查询、账单查询、3DS 查询、身份信息刷新。
- 用户管理：支持管理员 / 普通用户两种角色。
- 登录鉴权：基于 Cookie Session，首次启动会自动创建默认管理员。

## 目录结构

```text
.
├── README.md
├── frontend
│   ├── apps/web          # React 管理后台
│   ├── packages/ui       # 共享 UI 组件
│   ├── package.json
│   └── turbo.json
└── go
    ├── cmd/api           # API 启动入口
    ├── internal          # 业务代码
    ├── go.mod
    ├── .envrc            # 本地实际环境变量文件（自行创建，不提交）
    └── .envrc.example    # 后端环境变量示例
```

## 技术栈

### 前端

- Bun workspace
- Turbo
- React 19
- Vite 7
- TanStack Router
- TanStack Query
- Tailwind CSS 4

### 后端

- Go
- Gin
- GORM
- PostgreSQL

## 环境要求

- Bun `1.3+`
- Node.js `20+`
- Go `1.26+`
- 一个可访问的 PostgreSQL 实例

说明：

- 后端启动时会自动检测并创建 `POSTGRES_DB_NAME` 指定的数据库，并执行自动迁移。
- 因此 `POSTGRES_DSN` 对应的数据库账号需要具备连接 PostgreSQL 和创建数据库的权限。
- 前端开发模式默认通过 Vite 代理访问后端，不建议浏览器直接跨域请求后端接口。

## 源码启动

推荐启动顺序：先启动 PostgreSQL，再启动后端，最后启动前端。

### 1. 启动后端 API

进入后端目录：

```bash
cd go
```

先复制后端环境变量模板：

```bash
cp .envrc.example .envrc
```

如果你使用 `direnv`，再执行：

```bash
direnv allow
```

然后按需修改 `.envrc`。模板内容如下：

```bash
export PORT=8088
export POSTGRES_DSN='postgres://<user>:<password>@127.0.0.1:5432/postgres?sslmode=disable'
export POSTGRES_DB_NAME=gpt_team
export ACCOUNT_ENCRYPTION_KEY='0123456789abcdef0123456789abcdef'

# 以下变量按需配置；不用相关能力时可以先留空
export EFUNCARD_BASE_URL='https://card.efuncard.com'
export EFUNCARD_API_KEY='<your-efuncard-api-key>'
export CLOUDMAIL_BASE_URL='https://puax.cloud'
export CLOUDMAIL_API_TOKEN='<your-cloudmail-token>'
export DUCKMAIL_BASE_URL='https://api.duckmail.sbs'
```

然后启动服务：

```bash
go run ./cmd/api
```

成功启动后，默认可通过下面的地址检查服务健康状态：

```bash
curl http://localhost:8088/api/health
```

如果返回类似下面的数据，说明后端已经正常启动：

```json
{"data":{"status":"ok"}}
```

### 2. 启动前端管理后台

进入前端目录并安装依赖：

```bash
cd frontend
bun install
```

复制前端环境变量文件：

```bash
cp apps/web/.env.example apps/web/.env.local
```

默认情况下，前端会把 `/api` 代理到 `http://localhost:8088`。如果你的后端不是跑在 `8088`，需要修改 `apps/web/.env.local`：

```bash
VITE_API_BASE_URL=/api
VITE_API_PROXY_TARGET=http://localhost:你的后端端口
```

启动前端：

```bash
bun run dev
```

启动后访问：

```text
http://localhost:5173
```

### 3. 首次登录

当用户表为空时，后端会自动创建一个默认管理员：

- 用户名：`admin`
- 密码：`admin`

建议首次登录后立即修改或新增正式管理员账号。

## 关键环境变量说明

### 后端

| 变量名 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `PORT` | 否 | `8080` | API 监听端口。前端默认代理到 `8088`，本地联调建议显式设为 `8088`。 |
| `POSTGRES_DSN` | 是 | - | PostgreSQL 连接串。应用会基于它检测并创建业务库。 |
| `POSTGRES_DB_NAME` | 否 | `gpt_team` | 业务数据库名。 |
| `ACCOUNT_ENCRYPTION_KEY` | 是 | - | 必须是 32 字节明文，或 base64 编码后的 32 字节密钥。 |
| `EFUNCARD_BASE_URL` | 否 | `https://card.efuncard.com` | 卡密服务地址。 |
| `EFUNCARD_API_KEY` | 否 | - | 卡密激活、查询、账单、3DS 等能力依赖的 API Key。 |
| `CLOUDMAIL_BASE_URL` | 否 | `https://puax.cloud` | Cloudmail 服务地址。 |
| `CLOUDMAIL_API_TOKEN` | 否 | - | Cloudmail 鉴权 Token。也可以使用 `CLOUDMAIL_AUTHORIZATION`。 |
| `DUCKMAIL_BASE_URL` | 否 | `https://api.duckmail.sbs` | DuckMail 服务地址。 |

### 前端

| 变量名 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `VITE_API_BASE_URL` | 否 | `/api` | 前端请求 API 的基础路径。开发环境建议保留 `/api`。 |
| `VITE_API_PROXY_TARGET` | 否 | `http://localhost:8088` | Vite 开发代理目标地址。 |

## 常用命令

### 后端

```bash
cd go
go test ./...
go build ./cmd/api
```

### 前端

```bash
cd frontend
bun run test
bun run build
bun run typecheck
bun run lint
```

## 开发注意事项

- 前端开发环境默认通过 Vite 代理 `/api`，这样浏览器请求和 Session Cookie 行为最稳定。
- 如果后端端口不是 `8088`，记得同步修改 `apps/web/.env.local` 中的 `VITE_API_PROXY_TARGET`。
- `ACCOUNT_ENCRYPTION_KEY` 长度不对时，后端会直接启动失败。
- `POSTGRES_DSN` 对应的账号如果没有建库权限，后端也会在启动阶段报错。
- 部分页面依赖第三方服务：
  - 卡密相关操作依赖 `EFUNCARD_API_KEY`
  - Cloudmail / DuckMail 相关能力依赖对应服务配置

## 适合的本地开发方式

日常联调推荐开两个终端：

```bash
# 终端 1
cd /path/to/gpt-team/go
go run ./cmd/api
```

```bash
# 终端 2
cd /path/to/gpt-team/frontend
bun run dev
```

这样就可以通过前端页面直接调用本地后端接口进行开发。
