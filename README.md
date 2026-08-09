# NetView · 网络媒体收藏管理系统

自部署的图片/视频收藏管理器：把网上看到的有趣图片和视频单独保存，与手机相册隔离。运行在本地或局域网，支持多人使用。

![CI](https://github.com/ac227/netview/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/ac227/netview/actions/workflows/release.yml/badge.svg)

## 功能

- **多来源收集**：粘贴链接（自动抓取标题/封面）、上传文件、拖拽导入、**浏览器插件**（Chrome/Edge 右键一键收藏）
- **媒体库**：网格浏览、内置看图器/视频播放器（支持拖动进度）、收藏、搜索与筛选（关键词/标签/分类/类型）
- **视频下载**：直链直接下载；其他网站（B站/抖音/YouTube 等）接入 yt-dlp 兜底，后台任务队列
- **组织管理**：标签、分类、收藏
- **AI 自动打标签**：对接任意 OpenAI 兼容 API（OpenAI / DeepSeek / 通义等），自动生成标题/描述/标签
- **多人访问**：单共享密码 + JWT 认证，局域网内输入密码即可使用
- **API 文档**：完整 OpenAPI 3.0 规范（`docs/openapi.yaml`），方便对接浏览器插件等客户端

## 快速开始

### 前置依赖

- **PostgreSQL 16**（本地安装，非容器）：`brew install postgresql@16 && brew services start postgresql@16`
- **ffmpeg**（生成视频缩略图）：`brew install ffmpeg`
- **Go 1.22+、Node.js 20+**（仅编译时需要）
- **yt-dlp**（可选，下载 B站/YouTube 等平台视频）：`brew install yt-dlp`

### 编译与运行

```bash
./scripts/setup_db.sh   # 首次初始化 PostgreSQL 数据库
./scripts/build.sh      # 编译前端并嵌入，产出 bin/netview 单个可执行文件
./bin/netview           # 运行，访问 http://localhost:8080
```

整个程序只有一个二进制 `bin/netview`（约 25MB），前端已内嵌其中，复制到任意位置都能跑。

> 数据默认保存在 `<可执行文件所在目录>/data`，可通过 `NETVIEW_DATA_DIR` 覆盖。
> 无论从哪里启动，数据都落在同一位置，不会因工作目录不同而"丢失"。

### 从 Release 下载二进制

在 [Releases](https://github.com/ac227/netview/releases) 页下载对应平台的二进制即可，无需本地编译：

| 平台 | 文件 |
|---|---|
| macOS (Apple Silicon) | `netview-darwin-arm64` |
| macOS (Intel) | `netview-darwin-amd64` |
| Linux (x86_64 / ARM64) | `netview-linux-amd64` / `netview-linux-arm64` |
| Windows | `netview-windows-amd64.exe` |
| 浏览器插件 | `netview-extension.zip` |

macOS 首次运行如提示"无法验证开发者"，在「系统设置 → 隐私与安全性」中点「仍要打开」。

## 备份与恢复

完整备份 = 数据库（条目/标签/分类/设置/密码哈希）+ 媒体文件 + 配置。

```bash
./scripts/backup.sh                          # 备份到 ./backups（保留最近 10 份，自动删旧）
./scripts/backup.sh ~/mybackup 30            # 自定义目录 + 保留 30 份
./scripts/restore.sh backups/netview_20260809_130000   # 从备份恢复

# macOS 定时备份（每天 03:00，保留 30 份）
./scripts/setup-scheduled-backup.sh          # 安装定时任务
./scripts/remove-scheduled-backup.sh         # 卸载
```

> 建议首次部署后就执行一次 `./scripts/backup.sh` 建立基线。

## 浏览器插件（Chrome / Edge）

安装：打开 `chrome://extensions` → 开启「开发者模式」→「加载已解压的扩展程序」→ 选择 `extension/` 目录。

- 右键**图片**：保存链接 / 下载原图到 NetView
- 右键**视频/链接/页面**：一键收藏到媒体库
- 点工具栏图标：扫描当前页面图片，逐张保存
- 首次在扩展**选项页**填写服务器地址和密码

详见 `extension/README.md`。

## 配置（环境变量）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `NETVIEW_PORT` | `8080` | 端口 |
| `NETVIEW_HOST` | `0.0.0.0` | 监听地址（局域网共享保持默认） |
| `NETVIEW_DB_DSN` | `postgres://netview:netview_dev@localhost:5432/netview` | 数据库连接串 |
| `NETVIEW_DATA_DIR` | `<二进制目录>/data` | 媒体文件存储目录（默认跟随可执行文件位置） |
| `NETVIEW_JWT_SECRET` | 开发用固定值 | 生产环境务必修改 |
| `NETVIEW_YTDLP_PATH` | `yt-dlp` | yt-dlp 可执行文件路径 |
| `NETVIEW_AI_BASE_URL` / `NETVIEW_AI_API_KEY` / `NETVIEW_AI_MODEL` | - | AI 默认配置（也可网页「设置」页配置） |

## 开发

```bash
./scripts/dev.sh   # 后端 go run + 前端 vite 热更新（localhost:5173）
```

## 发布（CI/CD）

项目使用 GitHub Actions 自动构建与发布：

- **CI**（`.github/workflows/ci.yml`）：每次 push / PR 自动执行 `go build`、`go vet`、前端构建，并在 PostgreSQL 上做冒烟测试。
- **Release**（`.github/workflows/release.yml`）：**推送 `v*` 版本标签**时自动交叉编译全平台二进制、打包浏览器插件，并生成 GitHub Release。

发布一个新版本：

```bash
git tag v1.0.0
git push origin v1.0.0   # 触发 Actions 自动构建并创建 Release
```

## 项目结构

```
NetView/
├── .github/workflows/    # GitHub Actions：ci.yml / release.yml
├── bin/netview           # 编译产物（单个可执行文件）
├── backend/              # Go 后端
│   ├── cmd/server/       # 入口（-version 可查看版本）
│   └── internal/
│       ├── api/          # HTTP 路由与处理器
│       ├── web/          # 内嵌的前端静态文件（编译时填充）
│       ├── auth/         # 共享密码 + JWT
│       ├── media/        # 条目 CRUD、缩略图
│       ├── meta/         # 网页 OG 元数据抓取
│       ├── download/     # 直链 / yt-dlp 下载
│       ├── ai/           # OpenAI 兼容打标签
│       ├── storage/      # 本地文件管理
│       └── db/           # pgx 连接与迁移
├── frontend/             # React + Vite + AntD
├── docs/openapi.yaml     # API 文档（OpenAPI 3.0）
├── extension/            # Chrome/Edge 浏览器插件
└── scripts/              # setup_db / build / start / backup / restore / 定时备份 / dev
```

## API 文档

`docs/openapi.yaml` 是完整的 OpenAPI 3.0 规范。认证流程：

```
POST /api/auth/status     # 是否已设置密码
POST /api/auth/login      # 密码 → JWT token
# 之后所有业务请求带: Authorization: Bearer <token>
```

核心接口速览：

- `GET/POST /api/items` — 列表（搜索/筛选）/ 创建链接条目
- `POST /api/items/upload` — 上传文件
- `GET /api/items/{id}/file` — 媒体流（支持 Range）
- `GET /api/items/{id}/thumbnail` — 缩略图
- `POST /api/items/{id}/download` — 触发下载
- `POST /api/items/{id}/ai-tag` — AI 打标签
- `GET /api/tags`、`GET/POST /api/categories` — 标签/分类
- `GET/PUT /api/settings`、`GET /api/system/stats` — 配置/统计

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Gin + pgx |
| 数据库 | PostgreSQL 16 |
| 前端 | React 19 + Vite + TypeScript + Ant Design |
| 存储 | 本地磁盘（默认 `data/`，可配置） |
| 视频缩略图 | ffmpeg |
| 视频下载 | 直链 + yt-dlp（可选） |

## License

[MIT](LICENSE)
