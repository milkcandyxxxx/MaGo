package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"mago/internal/config"
	"mago/internal/handler"
	"mago/internal/middleware"
	"mago/internal/model"
	"mago/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// loadTemplates 递归加载所有模板
func loadTemplates(dir string) *template.Template {
	funcMap := template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
	}

	tmpl := template.New("").Funcs(funcMap)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".html") {
			// 使用文件名作为模板名（去掉目录前缀）
			name := strings.TrimPrefix(path, dir)
			name = strings.TrimPrefix(name, "/")
			name = strings.TrimPrefix(name, "\\")
			name = strings.ReplaceAll(name, "\\", "/")

			// 读取文件内容
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			// 解析模板
			_, parseErr := tmpl.New(name).Parse(string(data))
			if parseErr != nil {
				return parseErr
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal("加载模板失败:", err)
	}

	return tmpl
}

func main() {
	if len(os.Args) < 2 {
		startServer()
		return
	}

	switch os.Args[1] {
	case "import":
		runImport()
	case "serve":
		startServer()
	default:
		fmt.Println("Usage: goblog [serve|import <dir>]")
	}
}

func startServer() {
	cfg := config.Load()

	// 初始化数据库
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 启用外键约束
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		log.Printf("警告: 启用外键约束失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&model.Article{}, &model.Category{}, &model.Tag{}, &model.Comment{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 创建复合排序索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_articles_sort ON articles(status, pinned DESC, created_at DESC)").Error; err != nil {
		log.Printf("警告: 创建索引失败: %v", err)
	}

	// 初始化服务
	articleService := service.NewArticleService(db)
	importService := service.NewImportService(db)

	// 初始化 handler
	homeHandler := handler.NewHomeHandler(articleService)
	postHandler := handler.NewPostHandler(db, articleService)
	archiveHandler := handler.NewArchiveHandler(articleService)
	categoryHandler := handler.NewCategoryHandler(articleService)
	tagHandler := handler.NewTagHandler(articleService)
	aboutHandler := handler.NewAboutHandler()
	searchHandler := handler.NewSearchHandler(db)
	rssHandler := handler.NewRSSHandler(db, "http://localhost:"+cfg.Port)
	commentHandler := handler.NewCommentHandler(db)
	adminHandler := handler.NewAdminHandler(db, articleService, importService, cfg.AdminPass)

	// 初始化 Gin
	r := gin.Default()

	// 注册自定义模板函数
	r.SetFuncMap(template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
	})

	// 加载模板
	tmpl := loadTemplates("templates")
	r.SetHTMLTemplate(tmpl)

	// 静态资源
	r.Static("/static", "./static")

	// Session 中间件
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	r.Use(sessions.Sessions("mago-session", store))

	// UTF-8 编码中间件
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Next()
	})

	// 前台路由
	r.GET("/", homeHandler.Index)
	r.GET("/post/:slug", postHandler.Show)
	r.GET("/archives", archiveHandler.Index)
	r.GET("/categories", categoryHandler.Index)
	r.GET("/category/:slug", categoryHandler.Show)
	r.GET("/tags", tagHandler.Index)
	r.GET("/tag/:slug", tagHandler.Show)
	r.GET("/about", aboutHandler.Index)
	r.GET("/search", searchHandler.Index)
	r.GET("/rss", rssHandler.Feed)
	r.POST("/comment", commentHandler.Create)

	// 管理后台路由
	admin := r.Group("/admin")
	{
		// 登录（不需要认证）
		admin.GET("/login", adminHandler.Login)
		admin.POST("/login", adminHandler.DoLogin)
		admin.GET("/logout", adminHandler.Logout)

		// 需要认证的路由
		auth := admin.Group("/")
		auth.Use(middleware.AuthRequired())
		{
			auth.GET("", adminHandler.Dashboard)

			// 文章管理
			auth.GET("/article/new", adminHandler.NewArticle)
			auth.POST("/article", adminHandler.CreateArticle)
			auth.GET("/article/:id/edit", adminHandler.EditArticle)
			auth.POST("/article/:id", adminHandler.UpdateArticle)
			auth.POST("/article/:id/delete", adminHandler.DeleteArticle)

			// 分类管理
			auth.GET("/categories", adminHandler.Categories)
			auth.POST("/category", adminHandler.CreateCategory)
			auth.POST("/category/:id/delete", adminHandler.DeleteCategory)

			// 标签管理
			auth.GET("/tags", adminHandler.Tags)
			auth.POST("/tag", adminHandler.CreateTag)
			auth.POST("/tag/:id/delete", adminHandler.DeleteTag)
		}
	}

	// 启动服务器
	log.Printf("服务器启动在 http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

func runImport() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: goblog import <dir>")
		os.Exit(1)
	}

	dir := os.Args[2]
	cfg := config.Load()

	// 初始化数据库
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		log.Printf("警告: 启用外键约束失败: %v", err)
	}

	if err := db.AutoMigrate(&model.Article{}, &model.Category{}, &model.Tag{}, &model.Comment{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 初始化导入服务
	importService := service.NewImportService(db)

	// 执行导入
	result, err := importService.ImportDir(dir)
	if err != nil {
		log.Fatal("导入失败:", err)
	}

	// 打印结果
	fmt.Printf("导入完成: 成功 %d 篇", result.Success)
	if result.Skipped > 0 {
		fmt.Printf(", 跳过 %d 篇", result.Skipped)
	}
	fmt.Println()

	if len(result.Errors) > 0 {
		fmt.Println("错误详情:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
}
