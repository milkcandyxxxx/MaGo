package service

import (
	"mago/internal/model"

	"gorm.io/gorm"
)

type ArticleService struct {
	db *gorm.DB
}

func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{db: db}
}

// GetPublishedArticles 获取已发布文章列表（分页）
func (s *ArticleService) GetPublishedArticles(page, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := s.db.Model(&model.Article{}).Where("status = ?", 1)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页修正
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	err := query.
		Preload("Category").
		Preload("Tags").
		Order("pinned DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&articles).Error

	return articles, total, err
}

// GetArticleBySlug 根据 slug 获取文章
func (s *ArticleService) GetArticleBySlug(slug string) (*model.Article, error) {
	var article model.Article
	err := s.db.
		Preload("Category").
		Preload("Tags").
		Where("slug = ? AND status = ?", slug, 1).
		First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetArchives 获取所有已发布文章（归档用）
func (s *ArticleService) GetArchives() ([]model.Article, error) {
	var articles []model.Article
	err := s.db.
		Where("status = ?", 1).
		Order("created_at DESC").
		Find(&articles).Error
	return articles, err
}

// GetArticlesByCategory 获取某分类下的文章
func (s *ArticleService) GetArticlesByCategory(categorySlug string, page, pageSize int) ([]model.Article, int64, *model.Category, error) {
	var category model.Category
	if err := s.db.Where("slug = ?", categorySlug).First(&category).Error; err != nil {
		return nil, 0, nil, err
	}

	var articles []model.Article
	var total int64

	query := s.db.Model(&model.Article{}).Where("status = ? AND category_id = ?", 1, category.ID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, nil, err
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	err := query.
		Preload("Tags").
		Order("pinned DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&articles).Error

	return articles, total, &category, err
}

// GetArticlesByTag 获取某标签下的文章
func (s *ArticleService) GetArticlesByTag(tagSlug string, page, pageSize int) ([]model.Article, int64, *model.Tag, error) {
	var tag model.Tag
	if err := s.db.Where("slug = ?", tagSlug).First(&tag).Error; err != nil {
		return nil, 0, nil, err
	}

	var articles []model.Article
	var total int64

	// 通过 article_tags 关联表查询
	query := s.db.Model(&model.Article{}).
		Joins("JOIN article_tags ON article_tags.article_id = articles.id").
		Where("articles.status = ? AND article_tags.tag_id = ?", 1, tag.ID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, nil, err
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	err := query.
		Preload("Category").
		Preload("Tags").
		Order("articles.pinned DESC, articles.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&articles).Error

	return articles, total, &tag, err
}

// GetAllCategories 获取所有分类
func (s *ArticleService) GetAllCategories() ([]model.Category, error) {
	var categories []model.Category
	err := s.db.Find(&categories).Error
	return categories, err
}

// GetAllTags 获取所有标签
func (s *ArticleService) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	err := s.db.Find(&tags).Error
	return tags, err
}
