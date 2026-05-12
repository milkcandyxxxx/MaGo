> [!CAUTION]
> 本项目 80% 内容由 AI 生成，介意请勿使用。

# MaGo (マゴ)

极简博客系统。

## 技术栈

Go + Gin + SQLite + GORM + Goldmark

## 使用

### 下载发行版

从 [Releases](https://github.com/yourusername/mago/releases) 下载对应平台的二进制文件。

### 从源码编译

```bash
git clone https://github.com/yourusername/mago.git
cd mago
go build -o mago .
```

### 运行

```bash
ADMIN_PASSWORD=your_password ./mago
```

访问 http://localhost:8080

### 导入文章

```bash
./mago import ./content/
```

## 配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | 8080 | 端口 |
| `DB_PATH` | blog.db | 数据库路径 |
| `ADMIN_PASSWORD` | - | 管理密码 |
| `SESSION_SECRET` | 随机 | 会话密钥 |
