package handler

import (
	"net/http"
	"strconv"

	"mago/internal/service"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	articleService *service.ArticleService
}

func NewTagHandler(articleService *service.ArticleService) *TagHandler {
	return &TagHandler{articleService: articleService}
}

// Index 显示所有标签
func (h *TagHandler) Index(c *gin.Context) {
	tags, err := h.articleService.GetAllTags()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "tags.html", gin.H{
		"Title": "标签",
		"Tags":  tags,
	})
}

// Show 显示某标签下的文章
func (h *TagHandler) Show(c *gin.Context) {
	slug := c.Param("slug")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 10

	articles, total, tag, err := h.articleService.GetArticlesByTag(slug, page, pageSize)
	if err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"error": "标签不存在"})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	if page > totalPages && totalPages > 0 {
		page = totalPages
		articles, _, _, err = h.articleService.GetArticlesByTag(slug, page, pageSize)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}
	}

	c.HTML(http.StatusOK, "tag.html", gin.H{
		"Title":      tag.Name,
		"Tag":        tag,
		"Articles":   articles,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
	})
}
