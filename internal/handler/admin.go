package handler

import (
	"net/http"
	"strconv"

	"mago/internal/model"
	"mago/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db             *gorm.DB
	articleService *service.ArticleService
	importService  *service.ImportService
	adminPass      string
}

func NewAdminHandler(db *gorm.DB, articleService *service.ArticleService, importService *service.ImportService, adminPass string) *AdminHandler {
	return &AdminHandler{
		db:             db,
		articleService: articleService,
		importService:  importService,
		adminPass:      adminPass,
	}
}

// Login 显示登录页面
func (h *AdminHandler) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/login.html", gin.H{
		"Title": "管理后台 - 登录",
	})
}

// DoLogin 处理登录
func (h *AdminHandler) DoLogin(c *gin.Context) {
	password := c.PostForm("password")

	if password != h.adminPass {
		c.HTML(http.StatusOK, "admin/login.html", gin.H{
			"Title": "管理后台 - 登录",
			"Error": "密码错误",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("authenticated", true)
	session.Save()

	c.Redirect(http.StatusFound, "/admin")
}

// Logout 退出登录
func (h *AdminHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/admin/login")
}

// Dashboard 管理后台首页
func (h *AdminHandler) Dashboard(c *gin.Context) {
	var articles []model.Article
	h.db.Order("created_at DESC").Find(&articles)

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"Title":    "管理后台",
		"Articles": articles,
	})
}

// NewArticle 显示新建文章表单
func (h *AdminHandler) NewArticle(c *gin.Context) {
	var categories []model.Category
	h.db.Find(&categories)

	var tags []model.Tag
	h.db.Find(&tags)

	c.HTML(http.StatusOK, "admin/article_form.html", gin.H{
		"Title":      "新建文章",
		"Categories": categories,
		"Tags":       tags,
		"Article":    nil,
	})
}

// CreateArticle 创建文章
func (h *AdminHandler) CreateArticle(c *gin.Context) {
	title := c.PostForm("title")
	slug := c.PostForm("slug")
	content := c.PostForm("content")

	// 验证必填字段
	if title == "" || slug == "" || content == "" {
		c.HTML(http.StatusBadRequest, "500.html", gin.H{"error": "标题、slug和内容不能为空"})
		return
	}

	// 验证并限制状态值
	status, _ := strconv.Atoi(c.PostForm("status"))
	if status != 0 && status != 1 {
		status = 0 // 默认草稿
	}

	// 验证置顶值
	pinned, _ := strconv.Atoi(c.PostForm("pinned"))
	if pinned < 0 {
		pinned = 0
	}

	// 验证分类ID
	categoryID, _ := strconv.Atoi(c.PostForm("category_id"))
	if categoryID < 0 {
		categoryID = 0
	}

	// 渲染 Markdown
	htmlContent, err := service.RenderMarkdown(content)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	// 生成摘要
	summary := service.GenerateSummary(content, 200)

	article := model.Article{
		Title:   title,
		Slug:    slug,
		Content: content,
		HTML:    htmlContent,
		Summary: summary,
		Status:  status,
		Pinned:  pinned,
	}

	if categoryID > 0 {
		catID := uint(categoryID)
		article.CategoryID = &catID
	}

	// 处理标签
	tagIDs := c.PostFormArray("tags")
	for _, tagIDStr := range tagIDs {
		tagID, err := strconv.Atoi(tagIDStr)
		if err == nil && tagID > 0 {
			var tag model.Tag
			if err := h.db.First(&tag, tagID).Error; err == nil {
				article.Tags = append(article.Tags, tag)
			}
		}
	}

	if err := h.db.Create(&article).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin")
}

// EditArticle 显示编辑文章表单
func (h *AdminHandler) EditArticle(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var article model.Article
	if err := h.db.Preload("Tags").First(&article, id).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"error": "文章不存在"})
		return
	}

	var categories []model.Category
	h.db.Find(&categories)

	var tags []model.Tag
	h.db.Find(&tags)

	// 获取文章的标签 ID
	var articleTagIDs []uint
	for _, tag := range article.Tags {
		articleTagIDs = append(articleTagIDs, tag.ID)
	}

	c.HTML(http.StatusOK, "admin/article_form.html", gin.H{
		"Title":         "编辑文章",
		"Article":       article,
		"Categories":    categories,
		"Tags":          tags,
		"ArticleTagIDs": articleTagIDs,
	})
}

// UpdateArticle 更新文章
func (h *AdminHandler) UpdateArticle(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var article model.Article
	if err := h.db.First(&article, id).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"error": "文章不存在"})
		return
	}

	title := c.PostForm("title")
	slug := c.PostForm("slug")
	content := c.PostForm("content")
	status, _ := strconv.Atoi(c.PostForm("status"))
	pinned, _ := strconv.Atoi(c.PostForm("pinned"))
	categoryID, _ := strconv.Atoi(c.PostForm("category_id"))

	// 渲染 Markdown
	htmlContent, err := service.RenderMarkdown(content)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	summary := service.GenerateSummary(content, 200)

	// 更新文章
	article.Title = title
	article.Slug = slug
	article.Content = content
	article.HTML = htmlContent
	article.Summary = summary
	article.Status = status
	article.Pinned = pinned

	if categoryID > 0 {
		catID := uint(categoryID)
		article.CategoryID = &catID
	} else {
		article.CategoryID = nil
	}

	// 更新标签
	var tags []model.Tag
	tagIDs := c.PostFormArray("tags")
	for _, tagIDStr := range tagIDs {
		tagID, _ := strconv.Atoi(tagIDStr)
		if tagID > 0 {
			var tag model.Tag
			h.db.First(&tag, tagID)
			tags = append(tags, tag)
		}
	}

	// 使用 Association 替换标签
	h.db.Model(&article).Association("Tags").Replace(tags)

	if err := h.db.Save(&article).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin")
}

// DeleteArticle 删除文章
func (h *AdminHandler) DeleteArticle(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.db.Delete(&model.Article{}, id).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin")
}

// Categories 管理分类
func (h *AdminHandler) Categories(c *gin.Context) {
	var categories []model.Category
	h.db.Find(&categories)

	c.HTML(http.StatusOK, "admin/categories.html", gin.H{
		"Title":      "分类管理",
		"Categories": categories,
	})
}

// CreateCategory 创建分类
func (h *AdminHandler) CreateCategory(c *gin.Context) {
	name := c.PostForm("name")
	slug := c.PostForm("slug")

	category := model.Category{Name: name, Slug: slug}
	if err := h.db.Create(&category).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/categories")
}

// DeleteCategory 删除分类
func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.db.Delete(&model.Category{}, id).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/categories")
}

// Tags 管理标签
func (h *AdminHandler) Tags(c *gin.Context) {
	var tags []model.Tag
	h.db.Find(&tags)

	c.HTML(http.StatusOK, "admin/tags.html", gin.H{
		"Title": "标签管理",
		"Tags":  tags,
	})
}

// CreateTag 创建标签
func (h *AdminHandler) CreateTag(c *gin.Context) {
	name := c.PostForm("name")
	slug := c.PostForm("slug")

	tag := model.Tag{Name: name, Slug: slug}
	if err := h.db.Create(&tag).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/tags")
}

// DeleteTag 删除标签
func (h *AdminHandler) DeleteTag(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.db.Delete(&model.Tag{}, id).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/tags")
}
