## BingWP 必应每日壁纸

必应每日壁纸浏览与下载工具，基于 Go + Mithril.js 构建。

## 预览

![界面](./screenshot.png)

## 功能

- 按月份浏览必应每日壁纸（2009年7月至今）
- 缩略图自动下载与压缩（480x360 JPEG）
- 悬停查看壁纸详细描述
- 响应式布局，适配桌面与移动端

## 数据说明

- 壁纸数据来源于 [有图必应](https://bing.wilii.cn/) API
- 数据覆盖范围：**2009年7月 ~ 至今**
- **2012年8月14日**源头无数据，该日期之前的壁纸信息可能不完整
- **2014年5月1日**之前的壁纸无 description，Phase 2 会自动跳过
- 旧壁纸（Bing 已下架）的缩略图会自动从源站 fallback 下载；若源站也不可用则跳过

## 使用

```bash
# 构建
make one

# 更新壁纸数据并下载缩略图（三阶段流程：列表→描述→缩略图）
./bin/bwp up

# 使用8个并发线程下载缩略图
./bin/bwp up -w 8

# 启动 Web 服务（默认 127.0.0.1:8080）
./bin/bwp
# 或
./bin/bwp web
```

## 更新流程

`up` 命令执行三阶段数据更新：

1. **列表数据获取** — 从 API 分页拉取壁纸列表，写入数据库（批量事务写入）
   - 空库时从最后一页往前补全所有数据
   - 有数据时检测日期断档，从断点处继续补充
   - 无断档时只拉取最新数据

2. **描述更新** — 根据 GUID 逐条请求详情接口，获取 description 并批量写入（仅处理 2014-05-01 之后的记录）

3. **缩略图下载** — 检查本地缩略图完整性（缺失/尺寸错误），多线程下载并生成缩略图
   - 优先从 Bing CDN 下载，404 时自动 fallback 到源站 URL
   - 小批量（≤50张）使用 UHD 画质，同时记录原图文件大小
   - 大批量使用多线程并发下载

## 配置

通过 `.env` 文件或环境变量配置：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| HTTP_HOST | 127.0.0.1 | 监听地址 |
| HTTP_PORT | 8080 | 监听端口 |
| IMAGE_DIR | ./images | 缩略图目录 |
| DB_DSN | bingwp.db | SQLite 数据库路径 |
| DB_TYPE | sqlite | 数据库类型 |
| LOG_FILE | ./logs/access.log | 日志文件路径 |
| WORKERS | 8 | 下载线程数 |

## 项目结构

```
├── main.go              # 入口，命令行解析
├── server.go            # HTTP 服务
├── update.go            # 数据更新（三阶段流程 + 文件日志）
├── update_test.go       # 更新流程集成测试
├── models/
│   ├── tables.go        # 数据表定义（Wallpaper struct）
│   ├── fields.go        # Entity 接口实现（ScanDest/UpdatePairs）
│   ├── conn.go          # 数据库连接与自动迁移
│   └── api.go           # API 响应类型
├── services/
│   ├── config.go        # 配置加载
│   ├── database.go      # 数据库查询与批量操作
│   ├── fetcher.go       # API 请求、图片下载与处理（含 fallback）
│   └── downloader.go    # 文件操作工具
├── handlers/
│   └── api.go           # HTTP API 处理
├── static/
│   ├── app.js           # Mithril.js 前端
│   └── style.css        # 样式（磨砂玻璃效果）
└── views/
    └── index.html        # SPA 入口
```

## 技术栈

- **后端**：Go / SQLite / goent ORM
- **前端**：Mithril.js / 原生 CSS
- **图片处理**：imaging（缩略图裁剪与压缩）
