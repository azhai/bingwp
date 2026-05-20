# Bing Wallpaper 项目重构设计文档

**日期**: 2026-05-20
**版本**: v1.0
**状态**: 待批准

## 1. 项目概述

### 1.1 目标
将现有的 Bing 壁纸展示项目重构为一个简洁、现代化的壁纸浏览应用。

### 1.2 核心需求
- 展示一个月的 Bing 壁纸缩略图（按时间从早到晚排列）
- 使用 SQLite3 存储数据（移除 URL 字段，保留文件大小）
- 仅下载缩略图到本地（不保存原始高清图）
- 支持增量更新功能

## 2. 视觉设计规范

### 2.1 布局风格
- **方案**: 响应式瀑布流网格 (CSS Grid)
- **特点**: 自动适应屏幕宽度，图片等比缩放

### 2.2 配色方案
- **背景色**: `#ffffff` (纯白)
- **边框色**: `#e5e7eb` (浅灰)
- **文字主色**: `#111827` (深灰黑)
- **辅助文字**: `#6b7280` (中灰)
- **阴影**: `0 2px 8px rgba(0,0,0,0.08)` (柔和)

### 2.3 卡片样式
- **布局**: 图片 + 下方信息区
- **圆角**: `6px`
- **间距**: 卡片间 `12px`，图片与信息区 `8px`
- **信息内容**: 标题、日期、文件大小

### 2.4 缩略图规格
- **统一尺寸**: 宽度自适应，高度按比例缩放（建议 16:9）
- **存储路径**: `images/{yyyymm}/{dd}.jpg`
  - 示例: `images/202604/01.jpg`

## 3. 数据模型设计

### 3.1 数据源结构
**来源**: `https://bing.npanuhin.me/CN-zh.{YYYY}.{MM}.json`

**原始字段**:
```json
{
  "title": "粉色牵牛花里的日本树蛙",
  "caption": "跃入四月",
  "subtitle": "日本树蛙",
  "copyright": "© Tetsuya Tanooka/Getty Images",
  "description": "详细描述文本...",
  "date": "2026-04-01",
  "bing_url": "https://bing.com/th?id=OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg",
  "url": "https://bing.npanuhin.me/CN/zh/2026-04-01.jpg"
}
```

### 3.2 数据库表结构 (SQLite3)

**表名**: `wallpapers`

| 字段名 | 类型 | 说明 | 示例 |
|--------|------|------|------|
| id | INTEGER PRIMARY KEY | 自增ID | 1 |
| date | TEXT UNIQUE | 日期 | '2026-04-01' |
| title | TEXT | 标题 | '粉色牵牛花里的日本树蛙' |
| caption | TEXT | 副标题 | '跃入四月' |
| subtitle | TEXT | 子标题 | '日本树蛙' |
| copyright | TEXT | 版权信息 | '© Tetsuya Tanooka/Getty Images' |
| description | TEXT | 详细描述 | '...' |
| bing_file | TEXT | Bing文件标识 | 'OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg' |
| file_size | INTEGER | 文件大小(字节) | 245760 |
| local_path | TEXT | 本地存储路径 | 'images/202604/01.jpg' |

**字段说明**:
- `bing_file`: 从 `bing_url` 中提取 `id=` 之后的内容
  - 输入: `https://bing.com/th?id=OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg`
  - 提取: `OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg`
- **已移除字段**: `url`, `bing_url` (不存储URL)

### 3.3 数据处理规则

#### 3.3.1 URL 转换为 bing_file
```python
def extract_bing_file(bing_url):
    """从 bing_url 中提取文件标识"""
    if 'id=' in bing_url:
        return bing_url.split('id=')[-1]
    return None
```

#### 3.3.2 本地路径生成
```python
def generate_local_path(date_str):
    """生成本地存储路径"""
    # date_str: '2026-04-01'
    yyyymm = date_str.replace('-', '')[:6]  # '202604'
    dd = date_str.split('-')[-1]           # '01'
    return f"images/{yyyymm}/{dd}.jpg"
```

#### 3.3.3 文件大小获取
在下载缩略图时，通过 HTTP Header 或本地文件系统获取实际大小。

## 4. 功能模块设计

### 4.1 数据获取模块

**功能**:
- 从 JSON API 获取指定月份的壁纸数据
- 解析并转换数据格式
- 计算或获取文件大小

**接口**:
```go
func FetchMonthData(year, month int) ([]WallpaperData, error)
```

### 4.2 图片下载模块

**功能**:
- 从 `url` 字段下载缩略图
- 统一保存到 `images/{yyyymm}/{dd}.jpg`
- 返回实际文件大小
- 避免重复下载（检查文件是否存在）

**接口**:
```go
func DownloadThumbnail(url, localPath string) (int64, error)
func GetFileSize(localPath string) int64
```

### 4.3 数据库操作模块

**功能**:
- 创建/初始化数据库表
- 插入新数据（忽略已存在的记录）
- 查询指定月份的数据（按日期升序）
- 获取最后更新日期

**接口**:
```go
func InitDB(dbPath string) error
func InsertWallpaper(wp *Wallpaper) error
func GetWallpapersByMonth(year, month int) ([]Wallpaper, error)
func GetLastUpdateDate() (string, error)
```

### 4.4 Update 更新命令

**触发方式**: `./bingwp update`

**逻辑流程**:
1. 查询数据库中最后一条记录的日期 (`last_date`)
2. 计算 `last_date` 到今天之间的所有日期
3. 对每个缺失的月份:
   - 调用 API 获取该月数据
   - 过滤出 `last_date` 之后的新数据
   - 下载每条数据的缩略图
   - 写入数据库
4. 输出更新统计信息

**伪代码**:
```
function update():
    last_date = db.get_last_update_date()
    today = current_date()

    for each month from last_date to today:
        data = fetch_month_data(month.year, month.month)

        for each item in data:
            if item.date > last_date:
                # 下载缩略图
                local_path = generate_path(item.date)
                file_size = download_thumbnail(item.url, local_path)

                # 构建数据库记录
                wallpaper = {
                    date: item.date,
                    title: item.title,
                    ...
                    bing_file: extract_bing_file(item.bing_url),
                    file_size: file_size,
                    local_path: local_path
                }

                # 写入数据库
                db.insert(wallpaper)

    print(f"Updated {count} new wallpapers")
```

### 4.5 Web 服务模块

**路由设计**:
- `GET /` - 显示当前月壁纸
- `GET /{YYYYMM}` - 显示指定月壁纸（如 `/202604`）
- `GET /images/*` - 静态文件服务（缩略图）

**页面渲染**:
- 使用 Go template 引擎
- 瀑布流网格布局
- 卡片包含：缩略图、标题、日期、文件大小
- 按时间从早到晚排序

## 5. 技术实现细节

### 5.1 项目结构（简化后）

```
bingwp/
├── main.go              # 入口文件 + CLI 参数解析
├── server.go            # HTTP 服务器配置
├── update.go            # update 命令实现
├── models/
│   └── wallpaper.go     # 数据模型定义
├── services/
│   ├── fetcher.go       # 数据获取服务
│   ├── downloader.go    # 图片下载服务
│   └── database.go      # 数据库操作服务
├── handlers/
│   └── page.go          # 页面处理器
├── views/
│   ├── index.html       # 主页面模板
│   └── card.html        # 卡片组件模板
├── images/              # 缩略图存储目录
│   └── {yyyymm}/
│       └── {dd}.jpg
├── static/
│   └── style.css        # 样式文件
└── bingwp.db            # SQLite3 数据库文件
```

### 5.2 依赖精简

**移除的依赖**:
- Tabler UI 框架 (CSS/JS)
- Bootstrap
- 复杂的 ORM 库（使用标准库 database/sql）

**保留/新增的依赖**:
- `database/sql` + `mattn/go-sqlite3` (SQLite驱动)
- `net/http` (HTTP客户端和服务端)
- `encoding/json` (JSON解析)

### 5.3 配置项

通过环境变量或 `.env` 文件配置:

```env
DB_PATH=./bingwp.db          # 数据库路径
IMAGE_DIR=./images            # 图片存储目录
PORT=8080                     # 服务端口
DATA_API_BASE=https://bing.npanuhin.me/CN-zh.{YYYY}.{MM}.json
```

## 6. 用户交互流程

### 6.1 初始化流程
```bash
# 1. 首次运行：下载历史数据
./bingwp update

# 2. 启动 Web 服务
./bingwp serve
# 访问 http://localhost:8080
```

### 6.2 日常更新
```bash
# 定期执行增量更新
./bingwp update
# 自动从最后一条记录更新到当天
```

### 6.3 浏览壁纸
- 打开浏览器访问 `http://localhost:8080`
- 默认显示当前月份壁纸
- 可通过 URL 切换月份：`/202604`, `/202605`
- 点击卡片查看大图（可选功能）

## 7. 性能考虑

### 7.1 图片优化
- 统一缩略图尺寸，避免过大图片
- 使用浏览器缓存（设置合适的 Cache-Control 头）
- 考虑懒加载（滚动时再加载图片）

### 7.2 数据库优化
- 为 `date` 字段建立唯一索引
- 查询时按日期排序（利用索引）
- 批量插入提高写入性能

### 7.3 并发控制
- 下载图片时限制并发数（建议 3-5 个并发）
- 避免对同一资源重复请求

## 8. 错误处理

### 8.1 网络错误
- API 请求失败：重试 3 次，间隔递增
- 图片下载失败：记录日志，跳过该条数据

### 8.2 数据验证
- 日期格式校验
- 必填字段检查（title, date, url, bing_url）
- 文件大小合理性校验（> 0 且 < 10MB）

### 8.3 文件系统错误
- 目录不存在：自动创建
- 磁盘空间不足：提前检测并提示
- 权限问题：明确错误信息

## 9. 测试策略

### 9.1 单元测试
- 数据解析测试
- 路径生成测试
- bing_file 提取测试
- 数据库 CRUD 测试

### 9.2 集成测试
- 完整的 update 流程测试
- Web 页面渲染测试
- API mock 测试

### 9.3 手动测试
- 不同月份的数据展示
- 增量更新功能
- 边界情况（跨年、闰月等）

## 10. 未来扩展（可选）

- [ ] 点击查看大图详情页
- [ ] 按年份归档浏览
- [ ] 搜索功能（按标题/描述）
- [ ] 壁纸收藏功能
- [ ] RSS 订阅输出
- [ ] Docker 化部署

---

## 附录 A: 数据示例

### 输入数据 (JSON API)
```json
{
  "title": "粉色牵牛花里的日本树蛙",
  "caption": "跃入四月",
  "subtitle": "日本树蛙",
  "copyright": "© Tetsuya Tanooka/Getty Images",
  "description": "四月是全国青蛙月...",
  "date": "2026-04-01",
  "bing_url": "https://bing.com/th?id=OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg",
  "url": "https://bing.npanuhin.me/CN/zh/2026-04-01.jpg"
}
```

### 存储数据 (SQLite)
```sql
INSERT INTO wallpapers (
  date, title, caption, subtitle, copyright, description,
  bing_file, file_size, local_path
) VALUES (
  '2026-04-01',
  '粉色牵牛花里的日本树蛙',
  '跃入四月',
  '日本树蛙',
  '© Tetsuya Tanooka/Getty Images',
  '四月是全国青蛙月...',
  'OHR.JapaneseTreeFrog_ZH-CN6467379766_UHD.jpg',
  245760,
  'images/202604/01.jpg'
);
```

---

**文档状态**: ✅ 已完成初稿
**下一步**: 请用户审核并批准此设计方案
