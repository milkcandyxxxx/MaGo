package handler

import (
	"net/http"
	"strconv"

	"mago/internal/service"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	articleService *service.ArticleService
}

func NewCategoryHandler(articleService *service.ArticleService) *CategoryHandler {
	return &CategoryHandler{articleService: articleService}
}

// Index 显示所有分类
func (h *CategoryHandler) Index(c *gin.Context) {
	categories, err := h.articleService.GetAllCategories()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "categories.html", gin.H{
		"Title":      "分类",
		"Categories": categories,
	})
}

// Show 显示某分类下的文章
func (h *CategoryHandler) Show(c *gin.Context) {
	slug := c.Param("slug")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 10

	articles, total, category, err := h.articleService.GetArticlesByCategory(slug, page, pageSize)
	if err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"error": "分类不存在"})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	if page > totalPages && totalPages > 0 {
		page = totalPages
		articles, _, _, err = h.articleService.GetArticlesByCategory(slug, page, pageSize)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}
	}

	c.HTML(http.StatusOK, "category.html", gin.H{
		"Title":      category.Name,
		"Category":   category,
		"Articles":   articles,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
	})
}
