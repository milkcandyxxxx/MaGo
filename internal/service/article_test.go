package service

import (
	"testing"

	"mago/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupArticleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&model.Article{}, &model.Category{}, &model.Tag{}, &model.Comment{})
	require.NoError(t, err)

	return db
}

func createTestArticle(t *testing.T, db *gorm.DB, title string, slug string, status int) *model.Article {
	article := &model.Article{
		Title:   title,
		Slug:    slug,
		Content: "Test content for " + title,
		HTML:    "<p>Test content for " + title + "</p>",
		Summary: "Test summary",
		Status:  status,
	}
	err := db.Create(article).Error
	require.NoError(t, err)
	return article
}

func createTestCategory(t *testing.T, db *gorm.DB, name string, slug string) *model.Category {
	category := &model.Category{
		Name: name,
		Slug: slug,
	}
	err := db.Create(category).Error
	require.NoError(t, err)
	return category
}

func createTestTag(t *testing.T, db *gorm.DB, name string, slug string) *model.Tag {
	tag := &model.Tag{
		Name: name,
		Slug: slug,
	}
	err := db.Create(tag).Error
	require.NoError(t, err)
	return tag
}

func TestGetPublishedArticles(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建测试文章
	createTestArticle(t, db, "Published Article", "published", 1)
	createTestArticle(t, db, "Draft Article", "draft", 0)

	// 获取已发布文章
	articles, total, err := service.GetPublishedArticles(1, 10)
	require.NoError(t, err)

	assert.Equal(t, int64(1), total)
	assert.Len(t, articles, 1)
	assert.Equal(t, "Published Article", articles[0].Title)
}

func TestGetPublishedArticlesPagination(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建多篇文章
	for i := 0; i < 15; i++ {
		createTestArticle(t, db, "Article "+string(rune('A'+i)), "article-"+string(rune('a'+i)), 1)
	}

	// 第一页
	articles, total, err := service.GetPublishedArticles(1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, articles, 10)

	// 第二页
	articles, total, err = service.GetPublishedArticles(2, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, articles, 5)
}

func TestGetArticleBySlug(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建测试文章
	createTestArticle(t, db, "Test Article", "test-article", 1)

	// 获取文章
	article, err := service.GetArticleBySlug("test-article")
	require.NoError(t, err)
	assert.Equal(t, "Test Article", article.Title)
	assert.Equal(t, "test-article", article.Slug)
}

func TestGetArticleBySlugNotFound(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 获取不存在的文章
	_, err := service.GetArticleBySlug("non-existent")
	assert.Error(t, err)
}

func TestGetArchives(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建测试文章
	createTestArticle(t, db, "Article 1", "article-1", 1)
	createTestArticle(t, db, "Article 2", "article-2", 1)
	createTestArticle(t, db, "Draft", "draft", 0)

	// 获取归档
	articles, err := service.GetArchives()
	require.NoError(t, err)
	assert.Len(t, articles, 2) // 只返回已发布文章
}

func TestGetArticlesByCategory(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建分类
	category := createTestCategory(t, db, "Test Category", "test-category")

	// 创建文章并关联分类
	article := &model.Article{
		Title:      "Categorized Article",
		Slug:       "categorized",
		Content:    "Content",
		HTML:       "<p>Content</p>",
		Summary:    "Summary",
		Status:     1,
		CategoryID: &category.ID,
	}
	err := db.Create(article).Error
	require.NoError(t, err)

	// 获取分类下的文章
	articles, total, cat, err := service.GetArticlesByCategory("test-category", 1, 10)
	require.NoError(t, err)

	assert.Equal(t, int64(1), total)
	assert.Len(t, articles, 1)
	assert.Equal(t, "Test Category", cat.Name)
	assert.Equal(t, "Categorized Article", articles[0].Title)
}

func TestGetArticlesByTag(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建标签
	tag := createTestTag(t, db, "Test Tag", "test-tag")

	// 创建文章
	article := createTestArticle(t, db, "Tagged Article", "tagged", 1)

	// 关联标签
	err := db.Model(article).Association("Tags").Append(tag)
	require.NoError(t, err)

	// 获取标签下的文章
	articles, total, tagResult, err := service.GetArticlesByTag("test-tag", 1, 10)
	require.NoError(t, err)

	assert.Equal(t, int64(1), total)
	assert.Len(t, articles, 1)
	assert.Equal(t, "Test Tag", tagResult.Name)
	assert.Equal(t, "Tagged Article", articles[0].Title)
}

func TestGetAllCategories(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建分类
	createTestCategory(t, db, "Category 1", "category-1")
	createTestCategory(t, db, "Category 2", "category-2")

	// 获取所有分类
	categories, err := service.GetAllCategories()
	require.NoError(t, err)
	assert.Len(t, categories, 2)
}

func TestGetAllTags(t *testing.T) {
	db := setupArticleTestDB(t)
	service := NewArticleService(db)

	// 创建标签
	createTestTag(t, db, "Tag 1", "tag-1")
	createTestTag(t, db, "Tag 2", "tag-2")

	// 获取所有标签
	tags, err := service.GetAllTags()
	require.NoError(t, err)
	assert.Len(t, tags, 2)
}
