package handler

import (
	"net/http"
	"strings"

	"mago/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SearchHandler struct {
	db *gorm.DB
}

func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

func (h *SearchHandler) Index(c *gin.Context) {
	keyword := c.Query("q")
	var articles []model.Article

	if keyword != "" {
		// 转义 SQL 通配符防止注入
		escapedKeyword := strings.ReplaceAll(keyword, "%", "\\%")
		escapedKeyword = strings.ReplaceAll(escapedKeyword, "_", "\\_")

		h.db.Where("status = ? AND (title LIKE ? OR content LIKE ?)", 1,
			"%"+escapedKeyword+"%", "%"+escapedKeyword+"%").
			Order("created_at DESC").
			Limit(50).
			Find(&articles)
	}

	c.HTML(http.StatusOK, "search.html", gin.H{
		"Title":    "搜索",
		"Keyword":  keyword,
		"Articles": articles,
	})
}
