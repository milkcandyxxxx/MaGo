package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"mago/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RSSHandler struct {
	db      *gorm.DB
	baseURL string
}

func NewRSSHandler(db *gorm.DB, baseURL string) *RSSHandler {
	return &RSSHandler{db: db, baseURL: baseURL}
}

func (h *RSSHandler) Feed(c *gin.Context) {
	var articles []model.Article
	h.db.Where("status = ?", 1).
		Order("created_at DESC").
		Limit(20).
		Find(&articles)

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, h.generateRSS(articles))
}

func (h *RSSHandler) generateRSS(articles []model.Article) string {
	rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>MaGo</title>
    <link>%s</link>
    <description>MaGo (マゴ) - 极简博客</description>
    <language>zh-CN</language>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href="%s/rss" rel="self" type="application/rss+xml"/>
`

	lastBuild := time.Now().Format(time.RFC1123Z)
	rss = fmt.Sprintf(rss, h.baseURL, lastBuild, h.baseURL)

	for _, article := range articles {
		// 转义 XML 特殊字符
		safeTitle := escapeXML(article.Title)
		safeSummary := escapeXML(article.Summary)

		item := fmt.Sprintf(`    <item>
      <title>%s</title>
      <link>%s/post/%s</link>
      <guid>%s/post/%s</guid>
      <pubDate>%s</pubDate>
      <description><![CDATA[%s]]></description>
    </item>
`, safeTitle, h.baseURL, article.Slug, h.baseURL, article.Slug,
			article.CreatedAt.Format(time.RFC1123Z), safeSummary)
		rss += item
	}

	rss += `  </channel>
</rss>`

	return rss
}

// escapeXML 转义 XML 特殊字符
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
