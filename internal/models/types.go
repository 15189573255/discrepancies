package models

// DiffItem 表示一个差异项
type DiffItem struct {
	RelPath    string `json:"relPath"`    // 相对路径
	Type       string `json:"type"`       // "added" | "modified" | "deleted"
	Selected   bool   `json:"selected"`   // 是否选中
	SourcePath string `json:"sourcePath"` // 源文件完整路径（工作目录中的路径）
}

// DiffLine 表示一行差异
type DiffLine struct {
	Type    string `json:"type"`    // "equal" | "insert" | "delete"
	Content string `json:"content"` // 行内容
}

// TextDiff 表示文本差异结果
type TextDiff struct {
	OldContent string     `json:"oldContent"` // 原始内容
	NewContent string     `json:"newContent"` // 新内容
	Lines      []DiffLine `json:"lines"`      // 差异行
}

// CompareResult 表示比较结果
type CompareResult struct {
	Items      []DiffItem `json:"items"`      // 差异项列表
	TotalFiles int        `json:"totalFiles"` // 总文件数
	Added      int        `json:"added"`      // 新增文件数
	Modified   int        `json:"modified"`   // 修改文件数
	Deleted    int        `json:"deleted"`    // 删除文件数
}

// ExcludeRule 排除规则
type ExcludeRule struct {
	Pattern  string `json:"pattern"`  // 匹配模式
	Type     string `json:"type"`     // "glob" | "regex"
	IsDir    bool   `json:"isDir"`    // 是否仅匹配目录
	Enabled  bool   `json:"enabled"`  // 是否启用
	Comment  string `json:"comment"`  // 备注说明
}

// Config 应用配置
type Config struct {
	LastZipPath   string        `json:"lastZipPath"`   // 上次选择的 ZIP 文件路径
	LastWorkDir   string        `json:"lastWorkDir"`   // 上次选择的工作目录
	LastOutputDir string        `json:"lastOutputDir"` // 上次选择的输出目录
	ExcludeRules  []ExcludeRule `json:"excludeRules"`  // 排除规则列表
}

// ProgressEvent 进度事件
type ProgressEvent struct {
	Current int    `json:"current"` // 当前进度
	Total   int    `json:"total"`   // 总数
	Message string `json:"message"` // 进度消息
}
