package compare

import (
	"Discrepancies/internal/models"
	"os"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// TextDiffer 文本差异比较器
type TextDiffer struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

// NewTextDiffer 创建新的文本差异比较器
func NewTextDiffer() *TextDiffer {
	return &TextDiffer{
		dmp: diffmatchpatch.New(),
	}
}

// CompareTexts 比较两段文本并返回差异结果
func (d *TextDiffer) CompareTexts(oldText, newText string) *models.TextDiff {
	diffs := d.dmp.DiffMain(oldText, newText, true)
	diffs = d.dmp.DiffCleanupSemantic(diffs)

	result := &models.TextDiff{
		OldContent: oldText,
		NewContent: newText,
		Lines:      make([]models.DiffLine, 0),
	}

	// 转换为行级别的差异
	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for i, line := range lines {
			// 跳过最后一个空行（如果是换行符产生的）
			if i == len(lines)-1 && line == "" {
				continue
			}

			var diffType string
			switch diff.Type {
			case diffmatchpatch.DiffInsert:
				diffType = "insert"
			case diffmatchpatch.DiffDelete:
				diffType = "delete"
			default:
				diffType = "equal"
			}

			result.Lines = append(result.Lines, models.DiffLine{
				Type:    diffType,
				Content: line,
			})
		}
	}

	return result
}

// CompareFiles 比较 ZIP 中的文件和工作目录中的文件
func (d *TextDiffer) CompareFiles(zipReader *ZipReader, relPath, workFilePath string) (*models.TextDiff, error) {
	// 读取 ZIP 中的文件内容
	oldContent, err := zipReader.ReadFileContent(relPath)
	if err != nil {
		return nil, err
	}

	// 读取工作目录中的文件内容
	newContent, err := os.ReadFile(workFilePath)
	if err != nil {
		return nil, err
	}

	return d.CompareTexts(string(oldContent), string(newContent)), nil
}

// GetPrettyDiff 获取格式化的差异文本（用于终端显示）
func (d *TextDiffer) GetPrettyDiff(oldText, newText string) string {
	diffs := d.dmp.DiffMain(oldText, newText, true)
	diffs = d.dmp.DiffCleanupSemantic(diffs)
	return d.dmp.DiffPrettyText(diffs)
}

// IsTextFile 判断文件是否是文本文件（基于扩展名）
func IsTextFile(filename string) bool {
	textExtensions := map[string]bool{
		".txt":   true,
		".md":    true,
		".json":  true,
		".xml":   true,
		".html":  true,
		".htm":   true,
		".css":   true,
		".js":    true,
		".ts":    true,
		".go":    true,
		".py":    true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".hpp":   true,
		".cs":    true,
		".vb":    true,
		".sql":   true,
		".sh":    true,
		".bat":   true,
		".ps1":   true,
		".yaml":  true,
		".yml":   true,
		".toml":  true,
		".ini":   true,
		".cfg":   true,
		".conf":  true,
		".log":   true,
		".csv":   true,
		".tsv":   true,
		".svg":   true,
		".vue":   true,
		".jsx":   true,
		".tsx":   true,
		".svelte": true,
	}

	// 获取文件扩展名（转小写）
	ext := strings.ToLower(getFileExt(filename))
	return textExtensions[ext]
}

// getFileExt 获取文件扩展名
func getFileExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
		if filename[i] == '/' || filename[i] == '\\' {
			break
		}
	}
	return ""
}
