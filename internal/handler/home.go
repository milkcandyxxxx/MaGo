package handler

import (
	"net/http"
	"strconv"

	"mago/internal/service"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	articleService *service.ArticleService
}

func NewHomeHandler(articleService *service.ArticleService) *HomeHandler {
	return &HomeHandler{articleService: articleService}
}

func (h *HomeHandler) Index(c *gin.Context) {
	// 解析页码，自动修正非法值
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 10

	articles, total, err := h.articleService.GetPublishedArticles(page, pageSize)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	// 计算总页数
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	// 修正超出范围的页码
	if page > totalPages && totalPages > 0 {
		page = totalPages
		articles, _, err = h.articleService.GetPublishedArticles(page, pageSize)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}
	}

	c.HTML(http.StatusOK, "home.html", gin.H{
		"Title":      "首页",
		"Articles":   articles,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
	})
}
