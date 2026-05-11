package handler

import (
	"html/template"
	"net/http"

	"mago/internal/model"
	"mago/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostHandler struct {
	db             *gorm.DB
	articleService *service.ArticleService
}

func NewPostHandler(db *gorm.DB, articleService *service.ArticleService) *PostHandler {
	return &PostHandler{db: db, articleService: articleService}
}

func (h *PostHandler) Show(c *gin.Context) {
	slug := c.Param("slug")

	article, err := h.articleService.GetArticleBySlug(slug)
	if err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"error": "文章不存在"})
		return
	}

	// 加载评论（只加载顶级评论，回复通过关联加载）
	var comments []model.Comment
	h.db.Where("article_id = ? AND parent_id IS NULL AND status = ?", article.ID, 1).
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", 1).Order("created_at ASC")
		}).
		Order("created_at DESC").
		Find(&comments)

	// 将 HTML 字符串转换为 template.HTML 类型以正确渲染
	articleHTML := template.HTML(article.HTML)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.HTML(http.StatusOK, "post.html", gin.H{
		"Title":    article.Title,
		"Article":  article,
		"HTML":     articleHTML,
		"Comments": comments,
	})
}
