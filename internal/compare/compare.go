package compare

import (
	"Discrepancies/internal/models"
	"archive/zip"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ExcludeMatcher 排除规则匹配器
type ExcludeMatcher struct {
	rules         []models.ExcludeRule
	regexCache    map[string]*regexp.Regexp
	compiledRules []compiledRule
}

type compiledRule struct {
	rule    models.ExcludeRule
	regex   *regexp.Regexp
	pattern string
}

// NewExcludeMatcher 创建新的排除匹配器
func NewExcludeMatcher(rules []models.ExcludeRule) *ExcludeMatcher {
	m := &ExcludeMatcher{
		rules:         rules,
		regexCache:    make(map[string]*regexp.Regexp),
		compiledRules: make([]compiledRule, 0),
	}
	m.compile()
	return m
}

// compile 编译所有规则
func (m *ExcludeMatcher) compile() {
	for _, rule := range m.rules {
		if !rule.Enabled {
			continue
		}

		cr := compiledRule{rule: rule}

		if rule.Type == "regex" {
			// 正则表达式模式
			if re, err := regexp.Compile(rule.Pattern); err == nil {
				cr.regex = re
			}
		} else {
			// Glob 模式，转换为正则表达式
			cr.pattern = rule.Pattern
			regexPattern := globToRegex(rule.Pattern)
			if re, err := regexp.Compile(regexPattern); err == nil {
				cr.regex = re
			}
		}

		m.compiledRules = append(m.compiledRules, cr)
	}
}

// globToRegex 将 glob 模式转换为正则表达式
func globToRegex(pattern string) string {
	// 转义正则特殊字符
	result := regexp.QuoteMeta(pattern)
	// 替换 glob 通配符
	result = strings.ReplaceAll(result, `\*\*`, `.*`)    // ** 匹配任意路径
	result = strings.ReplaceAll(result, `\*`, `[^/]*`)  // * 匹配单级路径中的任意字符
	result = strings.ReplaceAll(result, `\?`, `.`)      // ? 匹配单个字符
	return "^" + result + "$"
}

// ShouldExclude 检查路径是否应该被排除
func (m *ExcludeMatcher) ShouldExclude(path string, isDir bool) bool {
	// 统一使用正斜杠
	path = filepath.ToSlash(path)

	for _, cr := range m.compiledRules {
		if cr.regex == nil {
			continue
		}

		// 如果规则仅匹配目录，跳过文件
		if cr.rule.IsDir && !isDir {
			// 但仍需检查路径中是否包含该目录
			if m.pathContainsDir(path, cr) {
				return true
			}
			continue
		}

		// 对于目录规则，检查路径中的每个部分
		if cr.rule.IsDir {
			parts := strings.Split(path, "/")
			for _, part := range parts {
				if cr.regex.MatchString(part) {
					return true
				}
			}
		} else {
			// 对于文件规则，匹配文件名或完整路径
			fileName := filepath.Base(path)
			if cr.regex.MatchString(fileName) || cr.regex.MatchString(path) {
				return true
			}
		}
	}

	return false
}

// pathContainsDir 检查路径中是否包含匹配的目录
func (m *ExcludeMatcher) pathContainsDir(path string, cr compiledRule) bool {
	parts := strings.Split(path, "/")
	for _, part := range parts[:len(parts)-1] { // 排除最后一个（文件名）
		if cr.regex.MatchString(part) {
			return true
		}
	}
	return false
}

// Comparer 负责比较 ZIP 文件和工作目录
type Comparer struct {
	zipPath        string
	workDir        string
	zipReader      *ZipReader
	excludeMatcher *ExcludeMatcher
	OnProgress     func(current, total int, message string)
}

// NewComparer 创建新的比较器
func NewComparer(zipPath, workDir string) *Comparer {
	return &Comparer{
		zipPath: zipPath,
		workDir: workDir,
	}
}

// SetExcludeRules 设置排除规则
func (c *Comparer) SetExcludeRules(rules []models.ExcludeRule) {
	c.excludeMatcher = NewExcludeMatcher(rules)
}

// Compare 执行比较并返回差异结果
func (c *Comparer) Compare() (*models.CompareResult, error) {
	// 打开 ZIP 文件
	var err error
	c.zipReader, err = NewZipReader(c.zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer c.zipReader.Close()

	// 获取 ZIP 中的文件列表
	zipFiles, err := c.zipReader.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list zip files: %w", err)
	}

	// 获取工作目录的文件列表
	workFiles, _, err := getAllFilesAndDirs(c.workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list work directory files: %w", err)
	}

	result := &models.CompareResult{
		Items: make([]models.DiffItem, 0),
	}

	totalFiles := len(zipFiles) + len(workFiles)
	processed := 0

	// 比较 ZIP 中的文件与工作目录
	for relPath, zipFile := range zipFiles {
		if c.shouldExclude(relPath, false) {
			continue
		}

		processed++
		c.emitProgress(processed, totalFiles, fmt.Sprintf("检查: %s", relPath))

		workFilePath, exists := workFiles[relPath]
		if !exists {
			// 文件在工作目录中不存在（已删除）
			result.Items = append(result.Items, models.DiffItem{
				RelPath:    relPath,
				Type:       "deleted",
				Selected:   true,
				SourcePath: "",
			})
			result.Deleted++
		} else {
			// 比较文件内容
			zipHash, err := c.getZipFileHash(zipFile)
			if err != nil {
				continue
			}
			workHash, err := fileHash(workFilePath)
			if err != nil {
				continue
			}

			if !bytes.Equal(zipHash, workHash) {
				// 文件已修改
				result.Items = append(result.Items, models.DiffItem{
					RelPath:    relPath,
					Type:       "modified",
					Selected:   true,
					SourcePath: workFilePath,
				})
				result.Modified++
			}
		}
	}

	// 查找工作目录中新增的文件
	for relPath, workFilePath := range workFiles {
		if c.shouldExclude(relPath, false) {
			continue
		}

		processed++
		c.emitProgress(processed, totalFiles, fmt.Sprintf("检查: %s", relPath))

		// 统一路径分隔符
		normalizedPath := filepath.ToSlash(relPath)
		if _, exists := zipFiles[normalizedPath]; !exists {
			// 这是新文件
			result.Items = append(result.Items, models.DiffItem{
				RelPath:    relPath,
				Type:       "added",
				Selected:   true,
				SourcePath: workFilePath,
			})
			result.Added++
		}
	}

	result.TotalFiles = len(result.Items)
	return result, nil
}

// shouldExclude 检查路径是否应该被排除
func (c *Comparer) shouldExclude(path string, isDir bool) bool {
	if c.excludeMatcher != nil {
		return c.excludeMatcher.ShouldExclude(path, isDir)
	}
	// 如果没有设置排除规则，使用默认逻辑
	return defaultShouldExclude(path)
}

// defaultShouldExclude 默认排除逻辑（向后兼容）
func defaultShouldExclude(path string) bool {
	path = filepath.ToSlash(path)
	pathParts := strings.Split(path, "/")

	for _, part := range pathParts {
		switch part {
		case "obj", "bin", ".idea", ".vs", "My Project", "Service References", "Properties":
			return true
		}
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".vbproj", ".csproj":
		return true
	}

	if strings.HasSuffix(path, ".vbproj.user") {
		return true
	}

	return false
}

// getZipFileHash 计算 ZIP 中文件的哈希
func (c *Comparer) getZipFileHash(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, rc); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// emitProgress 发送进度事件
func (c *Comparer) emitProgress(current, total int, message string) {
	if c.OnProgress != nil {
		c.OnProgress(current, total, message)
	}
}

// getAllFilesAndDirs 获取目录下的所有文件和子目录
func getAllFilesAndDirs(root string) (map[string]string, map[string]bool, error) {
	files := make(map[string]string)
	dirs := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "." {
			return nil
		}

		// 统一使用正斜杠
		relPath = filepath.ToSlash(relPath)

		if info.IsDir() {
			dirs[relPath] = true
		} else {
			files[relPath] = path
		}
		return nil
	})

	return files, dirs, err
}

// fileHash 计算文件的 MD5 哈希值
func fileHash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

// ExportDiffs 导出差异文件到输出目录
func ExportDiffs(items []models.DiffItem, outputDir string, onProgress func(current, total int, message string)) error {
	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	selectedItems := make([]models.DiffItem, 0)
	for _, item := range items {
		if item.Selected && item.Type != "deleted" {
			selectedItems = append(selectedItems, item)
		}
	}

	for i, item := range selectedItems {
		if onProgress != nil {
			onProgress(i+1, len(selectedItems), fmt.Sprintf("导出: %s", item.RelPath))
		}

		destPath := filepath.Join(outputDir, item.RelPath)
		if err := copyFile(item.SourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", item.RelPath, err)
		}
	}

	return nil
}

// copyFile 复制文件到目标路径
func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// CreateZip 创建 ZIP 压缩包
func CreateZip(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if path == sourceDir {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// 使用正斜杠
		relPath = filepath.ToSlash(relPath)

		if info.IsDir() {
			_, err := writer.Create(relPath + "/")
			return err
		}

		// 创建文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(w, file)
		return err
	})
}

// GenerateZipName 生成 ZIP 文件名
func GenerateZipName(baseName string) string {
	currentTime := time.Now()
	return fmt.Sprintf("%s_差分_%s.zip", baseName, currentTime.Format("2006年01月02日"))
}
