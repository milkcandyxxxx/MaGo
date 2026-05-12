package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
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

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// loadTemplates 从嵌入的文件系统加载模板
func loadTemplates() *template.Template {
	funcMap := template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
	}

	tmpl := template.New("").Funcs(funcMap)

	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		name := strings.TrimPrefix(path, "templates/")
		name = strings.ReplaceAll(name, "\\", "/")

		data, readErr := fs.ReadFile(templateFS, path)
		if readErr != nil {
			return readErr
		}

		_, parseErr := tmpl.New(name).Parse(string(data))
		if parseErr != nil {
			return parseErr
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
		fmt.Println("Usage: mago [serve|import <dir>]")
	}
}

func startServer() {
	cfg := config.Load()

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

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_articles_sort ON articles(status, pinned DESC, created_at DESC)").Error; err != nil {
		log.Printf("警告: 创建索引失败: %v", err)
	}

	articleService := service.NewArticleService(db)
	importService := service.NewImportService(db)

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

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
	})

	tmpl := loadTemplates()
	r.SetHTMLTemplate(tmpl)

	subStatic, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(subStatic))

	store := cookie.NewStore([]byte(cfg.SessionSecret))
	r.Use(sessions.Sessions("mago-session", store))

	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Next()
	})

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

	admin := r.Group("/admin")
	{
		admin.GET("/login", adminHandler.Login)
		admin.POST("/login", adminHandler.DoLogin)
		admin.GET("/logout", adminHandler.Logout)

		auth := admin.Group("/")
		auth.Use(middleware.AuthRequired())
		{
			auth.GET("", adminHandler.Dashboard)
			auth.GET("/article/new", adminHandler.NewArticle)
			auth.POST("/article", adminHandler.CreateArticle)
			auth.GET("/article/:id/edit", adminHandler.EditArticle)
			auth.POST("/article/:id", adminHandler.UpdateArticle)
			auth.POST("/article/:id/delete", adminHandler.DeleteArticle)
			auth.GET("/categories", adminHandler.Categories)
			auth.POST("/category", adminHandler.CreateCategory)
			auth.POST("/category/:id/delete", adminHandler.DeleteCategory)
			auth.GET("/tags", adminHandler.Tags)
			auth.POST("/tag", adminHandler.CreateTag)
			auth.POST("/tag/:id/delete", adminHandler.DeleteTag)
		}
	}

	log.Printf("服务器启动在 http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

func runImport() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: mago import <dir>")
		os.Exit(1)
	}

	dir := os.Args[2]
	cfg := config.Load()

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

	importService := service.NewImportService(db)

	result, err := importService.ImportDir(dir)
	if err != nil {
		log.Fatal("导入失败:", err)
	}

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
