package handler

import (
	"net/http"

	"mago/internal/service"

	"github.com/gin-gonic/gin"
)

type ArchiveHandler struct {
	articleService *service.ArticleService
}

func NewArchiveHandler(articleService *service.ArticleService) *ArchiveHandler {
	return &ArchiveHandler{articleService: articleService}
}

func (h *ArchiveHandler) Index(c *gin.Context) {
	articles, err := h.articleService.GetArchives()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "archive.html", gin.H{
		"Title":    "归档",
		"Articles": articles,
	})
}
