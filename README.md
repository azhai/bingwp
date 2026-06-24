# Bing Wallpaper

一个用来收集和浏览 Bing 每日壁纸的小工具。它会自动抓取 Bing 壁纸数据，保存到本地数据库，下载缩略图，并提供一个按月份浏览的网页。

![截图](screenshot-bingwp.png)

## 能做什么

- 抓取 Bing 历史上的每日壁纸（标题、描述、版权信息等）
- 下载 1920×1080 或 UHD 分辨率的壁纸图片到本地
- 启动本地 Web 服务，按月份浏览所有壁纸
- 增量更新：只抓取数据库中缺失或新增的数据

## 快速开始

### 1. 安装

需要安装 [Go](https://go.dev/)（1.21 或更高版本）。

```bash
go install github.com/azhai/bingwp@latest
```

安装完成后，可执行文件名为 `bingwp`，会自动放到 `$GOPATH/bin` 或 `$HOME/go/bin` 下，请确保该目录在 `PATH` 中。

> 注意：本项目 `go.mod` 中使用了 `replace` 指向本地仓库。如果直接 `go install` 失败，请先将 `github.com/azhai/goent` 和 `github.com/azhai/gobus` 放到对应的本地路径，或克隆本仓库后去掉 `replace` 再安装。

### 2. 准备配置

在运行目录下创建 `.env` 配置文件，可以从仓库中的 `example.env` 复制：

```bash
curl -O https://raw.githubusercontent.com/azhai/bingwp/main/example.env
mv example.env .env
```

### 3. 更新壁纸数据

首次运行需要抓取历史数据并下载缩略图：

```bash
bingwp up
```

默认下载 1920×1080 图片。如需下载 UHD 版本：

```bash
bingwp up --uhd
```

可以通过 `-w` 指定下载并发数：

```bash
bingwp up -w 16
```

### 4. 启动 Web 服务

```bash
bingwp web
```

打开浏览器访问 [http://127.0.0.1:8080](http://127.0.0.1:8080) 即可浏览壁纸。

直接运行 `bingwp` 不带参数时，也会启动 Web 服务。

## 配置说明

`.env` 文件可以从 [example.env](https://raw.githubusercontent.com/azhai/bingwp/main/example.env) 获取，修改其中配置即可生效。常用配置项如下：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `HTTP_HOST` | Web 服务监听地址 | `127.0.0.1` |
| `HTTP_PORT` | Web 服务端口 | `8080` |
| `IMAGE_DIR` | 壁纸原图存放目录 | `./images` |
| `THUMB_DIR` | 缩略图存放目录 | `./thumbs` |
| `LOG_FILE` | 日志文件路径 | `./logs/access.log` |
| `DB_DSN` | SQLite 数据库文件名 | `bingwp.db` |

## 启用 HTTPS

如果需要使用 HTTPS，将证书文件放到 `CERT_DIR` 指定的目录（默认 `./certs`）：

- `cert.pem`
- `key.pem`

启动服务时会自动检测并使用 HTTPS。

## 常用命令

```bash
# 安装
go install github.com/azhai/bingwp@latest

# 更新壁纸数据
bingwp up

# 更新 UHD 壁纸
bingwp up --uhd

# 启动 Web 服务
bingwp web
```

## 运行目录结构

首次运行后，当前目录下会生成以下文件和目录：

```
.
├── images/     # 下载的壁纸原图
├── thumbs/     # 缩略图
├── logs/       # 日志文件
├── bingwp.db   # SQLite 数据库
└── .env        # 配置文件
```

## 许可证

[LICENSE](LICENSE)
