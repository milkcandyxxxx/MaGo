package handler

import (
	"net/http"
	"strconv"

	"mago/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

// Create 创建评论
func (h *CommentHandler) Create(c *gin.Context) {
	articleIDStr := c.PostForm("article_id")
	nickname := c.PostForm("nickname")
	email := c.PostForm("email")
	content := c.PostForm("content")
	parentIDStr := c.PostForm("parent_id")
	articleSlug := c.PostForm("article_slug")

	// 验证必填字段
	if nickname == "" || content == "" {
		c.HTML(http.StatusBadRequest, "404.html", gin.H{"error": "昵称和内容不能为空"})
		return
	}

	// 验证 article_id
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil || articleID <= 0 {
		c.HTML(http.StatusBadRequest, "404.html", gin.H{"error": "无效的文章ID"})
		return
	}

	// 验证文章是否存在
	var article model.Article
	if err := h.db.First(&article, articleID).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"error": "文章不存在"})
		return
	}

	comment := model.Comment{
		ArticleID: uint(articleID),
		Nickname:  nickname,
		Email:     email,
		Content:   content,
		Status:    1,   // 默认直接发布
		ParentID:  nil, // 默认无父评论
	}

	// 处理回复
	if parentIDStr != "" {
		parentID, err := strconv.Atoi(parentIDStr)
		if err == nil && parentID > 0 {
			// 验证父评论是否存在
			var parentComment model.Comment
			if err := h.db.First(&parentComment, parentID).Error; err == nil {
				pid := uint(parentID)
				comment.ParentID = &pid
			}
		}
	}

	if err := h.db.Create(&comment).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/post/"+articleSlug+"#comments")
}
