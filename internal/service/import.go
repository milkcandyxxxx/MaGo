package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mago/internal/model"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type ImportService struct {
	db     *gorm.DB
	md     goldmark.Markdown
}

type FrontMatter struct {
	Title    string   `yaml:"title"`
	Slug     string   `yaml:"slug"`
	Category string   `yaml:"category"`
	Tags     []string `yaml:"tags"`
	Pinned   int      `yaml:"pinned"`
	Status   *int     `yaml:"status"` // 指针，nil 时默认为 1（已发布）
}

type ImportResult struct {
	Total    int
	Success  int
	Skipped  int
	Errors   []string
}

func NewImportService(db *gorm.DB) *ImportService {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,           // GitHub 风格 Markdown
			extension.Table,         // 表格
			extension.Strikethrough, // 删除线
			extension.Linkify,       // 自动链接
			extension.TaskList,      // 任务列表
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // 自动标题 ID
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // 换行转 <br>
			html.WithUnsafe(),    // 允许原始 HTML
		),
	)

	return &ImportService{db: db, md: md}
}

// ImportDir 导入目录下的所有 Markdown 文件
func (s *ImportService) ImportDir(dir string) (*ImportResult, error) {
	result := &ImportResult{}

	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	result.Total = len(files)

	for _, file := range files {
		if err := s.importFile(file); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", filepath.Base(file), err))
			continue
		}
		result.Success++
	}

	return result, nil
}

// importFile 导入单个 Markdown 文件
func (s *ImportService) importFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 解析 front-matter 和正文
	fm, content, err := parseFrontMatter(data)
	if err != nil {
		return fmt.Errorf("解析 front-matter 失败: %w", err)
	}

	// 验证必填字段
	if fm.Title == "" {
		return fmt.Errorf("缺少 title 字段")
	}
	if fm.Slug == "" {
		return fmt.Errorf("缺少 slug 字段")
	}

	// 渲染 Markdown 为 HTML
	var buf bytes.Buffer
	if err := s.md.Convert([]byte(content), &buf); err != nil {
		return fmt.Errorf("渲染 Markdown 失败: %w", err)
	}
	htmlContent := buf.String()

	// 生成摘要（取前 200 字符）
	summary := generateSummary(content, 200)

	// 处理分类
	var categoryID *uint
	if fm.Category != "" {
		cat, err := s.getOrCreateCategory(fm.Category)
		if err != nil {
			return fmt.Errorf("处理分类失败: %w", err)
		}
		categoryID = &cat.ID
	}

	// 处理标签
	var tags []model.Tag
	for _, tagName := range fm.Tags {
		tag, err := s.getOrCreateTag(tagName)
		if err != nil {
			return fmt.Errorf("处理标签失败: %w", err)
		}
		tags = append(tags, *tag)
	}

	// 确定状态
	status := 1 // 默认已发布
	if fm.Status != nil {
		status = *fm.Status
	}

	// Upsert 文章
	article := model.Article{
		Title:      fm.Title,
		Slug:       fm.Slug,
		Content:    content,
		HTML:       htmlContent,
		Summary:    summary,
		Status:     status,
		Pinned:     fm.Pinned,
		CategoryID: categoryID,
		Tags:       tags,
	}

	// 按 slug 查找，存在则更新，不存在则插入
	var existing model.Article
	result := s.db.Where("slug = ?", fm.Slug).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		// 新增
		return s.db.Create(&article).Error
	} else if result.Error != nil {
		return result.Error
	}

	// 更新
	article.ID = existing.ID
	return s.db.Save(&article).Error
}

// parseFrontMatter 解析 YAML front-matter
func parseFrontMatter(data []byte) (*FrontMatter, string, error) {
	content := string(data)

	// 检查是否以 --- 开头
	if !strings.HasPrefix(content, "---") {
		return nil, "", fmt.Errorf("缺少 front-matter（---）")
	}

	// 找到第二个 ---
	endIdx := strings.Index(content[3:], "---")
	if endIdx == -1 {
		return nil, "", fmt.Errorf("front-matter 未闭合")
	}

	// 提取 YAML 部分
	yamlPart := content[3 : endIdx+3]
	// 提取正文部分（跳过第二个 --- 后面的内容）
	bodyPart := strings.TrimSpace(content[endIdx+3+3:])

	// 解析 YAML
	var fm FrontMatter
	if err := yaml.Unmarshal([]byte(yamlPart), &fm); err != nil {
		return nil, "", fmt.Errorf("YAML 解析失败: %w", err)
	}

	return &fm, bodyPart, nil
}

// getOrCreateCategory 获取或创建分类
func (s *ImportService) getOrCreateCategory(name string) (*model.Category, error) {
	slug := generateSlug(name)

	var category model.Category
	err := s.db.Where("slug = ?", slug).First(&category).Error
	if err == gorm.ErrRecordNotFound {
		category = model.Category{Name: name, Slug: slug}
		if err := s.db.Create(&category).Error; err != nil {
			return nil, err
		}
		return &category, nil
	}
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// getOrCreateTag 获取或创建标签
func (s *ImportService) getOrCreateTag(name string) (*model.Tag, error) {
	slug := generateSlug(name)

	var tag model.Tag
	err := s.db.Where("slug = ?", slug).First(&tag).Error
	if err == gorm.ErrRecordNotFound {
		tag = model.Tag{Name: name, Slug: slug}
		if err := s.db.Create(&tag).Error; err != nil {
			return nil, err
		}
		return &tag, nil
	}
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// generateSlug 生成 URL 友好的 slug
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// 移除特殊字符，保留中文、字母、数字、连字符
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r > 127 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// RenderMarkdown 渲染 Markdown 为 HTML（导出函数）
func RenderMarkdown(content string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateSummary 生成摘要（导出函数）
func GenerateSummary(content string, maxLen int) string {
	return generateSummary(content, maxLen)
}

// generateSummary 生成摘要
func generateSummary(content string, maxLen int) string {
	// 移除 Markdown 格式
	content = strings.ReplaceAll(content, "#", "")
	content = strings.ReplaceAll(content, "**", "")
	content = strings.ReplaceAll(content, "*", "")
	content = strings.ReplaceAll(content, "`", "")
	content = strings.ReplaceAll(content, "[", "")
	content = strings.ReplaceAll(content, "]", "")
	content = strings.ReplaceAll(content, "(", "")
	content = strings.ReplaceAll(content, ")", "")

	// 取前 maxLen 个字符
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}
