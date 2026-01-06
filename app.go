package main

import (
	"Discrepancies/internal/compare"
	"Discrepancies/internal/config"
	"Discrepancies/internal/models"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx       context.Context
	configMgr *config.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化配置管理器
	var err error
	a.configMgr, err = config.NewManager()
	if err != nil {
		runtime.LogError(ctx, fmt.Sprintf("Failed to initialize config manager: %v", err))
	}
}

// SelectZipFile 打开文件选择对话框选择 ZIP 文件
func (a *App) SelectZipFile() (string, error) {
	defaultDir := ""
	if a.configMgr != nil {
		cfg := a.configMgr.Get()
		if cfg.LastZipPath != "" {
			defaultDir = filepath.Dir(cfg.LastZipPath)
		}
	}

	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "选择原始 ZIP 压缩包",
		DefaultDirectory: defaultDir,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "ZIP 文件 (*.zip)",
				Pattern:     "*.zip",
			},
		},
	})

	if err != nil {
		return "", err
	}

	if path != "" && a.configMgr != nil {
		a.configMgr.SetLastZipPath(path)
	}

	return path, nil
}

// SelectWorkDir 打开目录选择对话框选择工作目录
func (a *App) SelectWorkDir() (string, error) {
	defaultDir := ""
	if a.configMgr != nil {
		cfg := a.configMgr.Get()
		if cfg.LastWorkDir != "" {
			defaultDir = cfg.LastWorkDir
		}
	}

	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "选择工作目录",
		DefaultDirectory: defaultDir,
	})

	if err != nil {
		return "", err
	}

	if path != "" && a.configMgr != nil {
		a.configMgr.SetLastWorkDir(path)
	}

	return path, nil
}

// SelectOutputDir 打开目录选择对话框选择输出目录
func (a *App) SelectOutputDir() (string, error) {
	defaultDir := ""
	if a.configMgr != nil {
		defaultDir = a.configMgr.GetDefaultOutputDir()
	}

	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "选择输出目录",
		DefaultDirectory: defaultDir,
	})

	if err != nil {
		return "", err
	}

	if path != "" && a.configMgr != nil {
		a.configMgr.SetLastOutputDir(path)
	}

	return path, nil
}

// Compare 比较 ZIP 文件和工作目录
func (a *App) Compare(zipPath, workDir string) (*models.CompareResult, error) {
	if zipPath == "" {
		return nil, fmt.Errorf("请选择 ZIP 文件")
	}
	if workDir == "" {
		return nil, fmt.Errorf("请选择工作目录")
	}

	// 检查文件和目录是否存在
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("ZIP 文件不存在: %s", zipPath)
	}
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("工作目录不存在: %s", workDir)
	}

	comparer := compare.NewComparer(zipPath, workDir)

	// 设置排除规则
	if a.configMgr != nil {
		comparer.SetExcludeRules(a.configMgr.GetExcludeRules())
	}

	// 设置进度回调
	comparer.OnProgress = func(current, total int, message string) {
		runtime.EventsEmit(a.ctx, "backend:progress", models.ProgressEvent{
			Current: current,
			Total:   total,
			Message: message,
		})
	}

	result, err := comparer.Compare()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetTextDiff 获取文件的文本差异
func (a *App) GetTextDiff(zipPath, workDir, relPath string) (*models.TextDiff, error) {
	// 检查是否是文本文件
	if !compare.IsTextFile(relPath) {
		return nil, fmt.Errorf("不支持预览非文本文件")
	}

	// 打开 ZIP 文件
	zipReader, err := compare.NewZipReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	// 获取工作目录中的文件路径
	workFilePath := filepath.Join(workDir, relPath)

	// 比较文件
	differ := compare.NewTextDiffer()
	return differ.CompareFiles(zipReader, relPath, workFilePath)
}

// ExportDiffs 导出差异文件
func (a *App) ExportDiffs(items []models.DiffItem, outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf("请选择输出目录")
	}

	return compare.ExportDiffs(items, outputDir, func(current, total int, message string) {
		runtime.EventsEmit(a.ctx, "backend:progress", models.ProgressEvent{
			Current: current,
			Total:   total,
			Message: message,
		})
	})
}

// CreateZip 创建 ZIP 压缩包
func (a *App) CreateZip(sourceDir, baseName string) (string, error) {
	if sourceDir == "" {
		return "", fmt.Errorf("请指定源目录")
	}

	zipName := compare.GenerateZipName(baseName)
	zipPath := filepath.Join(filepath.Dir(sourceDir), zipName)

	if err := compare.CreateZip(sourceDir, zipPath); err != nil {
		return "", err
	}

	return zipPath, nil
}

// GetConfig 获取配置
func (a *App) GetConfig() models.Config {
	if a.configMgr == nil {
		return models.Config{}
	}
	return a.configMgr.Get()
}

// SaveConfig 保存配置
func (a *App) SaveConfig(cfg models.Config) error {
	if a.configMgr == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	return a.configMgr.Set(cfg)
}

// GetZipRootFolder 获取 ZIP 文件的根目录名称
func (a *App) GetZipRootFolder(zipPath string) (string, error) {
	zipReader, err := compare.NewZipReader(zipPath)
	if err != nil {
		return "", err
	}
	defer zipReader.Close()

	return zipReader.GetRootFolder(), nil
}

// GetExcludeRules 获取排除规则
func (a *App) GetExcludeRules() []models.ExcludeRule {
	if a.configMgr == nil {
		return []models.ExcludeRule{}
	}
	return a.configMgr.GetExcludeRules()
}

// SetExcludeRules 设置排除规则
func (a *App) SetExcludeRules(rules []models.ExcludeRule) error {
	if a.configMgr == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	return a.configMgr.SetExcludeRules(rules)
}

// AddExcludeRule 添加排除规则
func (a *App) AddExcludeRule(rule models.ExcludeRule) error {
	if a.configMgr == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	return a.configMgr.AddExcludeRule(rule)
}

// RemoveExcludeRule 删除排除规则
func (a *App) RemoveExcludeRule(index int) error {
	if a.configMgr == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	return a.configMgr.RemoveExcludeRule(index)
}

// ResetExcludeRules 重置为默认排除规则
func (a *App) ResetExcludeRules() error {
	if a.configMgr == nil {
		return fmt.Errorf("配置管理器未初始化")
	}
	return a.configMgr.ResetExcludeRules()
}
