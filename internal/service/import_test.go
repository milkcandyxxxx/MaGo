package service

import (
	"os"
	"path/filepath"
	"testing"

	"mago/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFM   *FrontMatter
		wantBody string
		wantErr  bool
	}{
		{
			name: "valid front-matter",
			input: `---
title: "Test Title"
slug: "test-slug"
category: "Test Category"
tags: ["tag1", "tag2"]
pinned: 5
---

This is the body content.`,
			wantFM: &FrontMatter{
				Title:    "Test Title",
				Slug:     "test-slug",
				Category: "Test Category",
				Tags:     []string{"tag1", "tag2"},
				Pinned:   5,
			},
			wantBody: "This is the body content.",
			wantErr:  false,
		},
		{
			name: "minimal front-matter",
			input: `---
title: "Minimal"
slug: "minimal"
---

Body here.`,
			wantFM: &FrontMatter{
				Title: "Minimal",
				Slug:  "minimal",
			},
			wantBody: "Body here.",
			wantErr:  false,
		},
		{
			name:    "missing front-matter",
			input:   "No front-matter here",
			wantErr: true,
		},
		{
			name: "unclosed front-matter",
			input: `---
title: "Test"
slug: "test"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := parseFrontMatter([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantFM.Title, fm.Title)
			assert.Equal(t, tt.wantFM.Slug, fm.Slug)
			assert.Equal(t, tt.wantFM.Category, fm.Category)
			assert.Equal(t, tt.wantFM.Tags, fm.Tags)
			assert.Equal(t, tt.wantFM.Pinned, fm.Pinned)
			assert.Equal(t, tt.wantBody, body)
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple english",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "with underscores",
			input: "hello_world",
			want:  "hello-world",
		},
		{
			name:  "with special chars",
			input: "Hello! World?",
			want:  "hello-world",
		},
		{
			name:  "chinese",
			input: "中文测试",
			want:  "中文测试",
		},
		{
			name:  "mixed",
			input: "Hello 世界",
			want:  "hello-世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSlug(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	tests := []struct {
		name    string
		content string
		maxLen  int
		want    string
	}{
		{
			name:    "short content",
			content: "Hello World",
			maxLen:  100,
			want:    "Hello World",
		},
		{
			name:    "long content",
			content: "This is a long content that should be truncated at some point because it exceeds the maximum length",
			maxLen:  20,
			want:    "This is a long conte...",
		},
		{
			name:    "with markdown",
			content: "# Title\n\n**Bold** and *italic* text",
			maxLen:  100,
			want:    " Title\n\nBold and italic text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSummary(tt.content, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple paragraph",
			content: "Hello World",
			want:    "<p>Hello World</p>\n",
		},
		{
			name:    "heading",
			content: "# Title",
			want:    "<h1 id=\"title\">Title</h1>\n",
		},
		{
			name:    "list",
			content: "- item1\n- item2",
			want:    "<ul>\n<li>item1</li>\n<li>item2</li>\n</ul>\n",
		},
		{
			name:    "code block",
			content: "```go\nfmt.Println(\"Hello\")\n```",
			want:    "<pre><code class=\"language-go\">fmt.Println(&quot;Hello&quot;)\n</code></pre>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderMarkdown(tt.content)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestImportDir(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建测试 Markdown 文件
	content := `---
title: "Test Article"
slug: "test-article"
category: "Test"
tags: ["tag1", "tag2"]
---

This is a test article.`

	err := os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte(content), 0644)
	require.NoError(t, err)

	// 创建测试数据库
	db := setupTestDB(t)

	// 创建表
	err = db.AutoMigrate(&model.Article{}, &model.Category{}, &model.Tag{})
	require.NoError(t, err)

	// 创建导入服务
	service := NewImportService(db)

	// 执行导入
	result, err := service.ImportDir(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 1, result.Success)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImportDirWithErrors(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建无效的 Markdown 文件
	invalidContent := `---
title: ""
slug: ""
---

This is invalid.`

	err := os.WriteFile(filepath.Join(tmpDir, "invalid.md"), []byte(invalidContent), 0644)
	require.NoError(t, err)

	// 创建测试数据库
	db := setupTestDB(t)

	// 创建导入服务
	service := NewImportService(db)

	// 执行导入
	result, err := service.ImportDir(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 1, result.Total)
	assert.Equal(t, 0, result.Success)
	assert.Equal(t, 1, result.Skipped)
	assert.Len(t, result.Errors, 1)
}
