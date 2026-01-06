package config

import (
	"Discrepancies/internal/models"
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = "config.json"
const configDirName = ".discrepancies"

// 默认排除规则
var defaultExcludeRules = []models.ExcludeRule{
	// 目录规则
	{Pattern: "obj", Type: "glob", IsDir: true, Enabled: true, Comment: ".NET 编译输出目录"},
	{Pattern: "bin", Type: "glob", IsDir: true, Enabled: true, Comment: ".NET 编译输出目录"},
	{Pattern: ".idea", Type: "glob", IsDir: true, Enabled: true, Comment: "JetBrains IDE 配置"},
	{Pattern: ".vs", Type: "glob", IsDir: true, Enabled: true, Comment: "Visual Studio 配置"},
	{Pattern: ".vscode", Type: "glob", IsDir: true, Enabled: true, Comment: "VS Code 配置"},
	{Pattern: "node_modules", Type: "glob", IsDir: true, Enabled: true, Comment: "Node.js 依赖"},
	{Pattern: "My Project", Type: "glob", IsDir: true, Enabled: true, Comment: "VB.NET 项目文件夹"},
	{Pattern: "Service References", Type: "glob", IsDir: true, Enabled: true, Comment: "服务引用"},
	{Pattern: "Properties", Type: "glob", IsDir: true, Enabled: true, Comment: ".NET 属性文件夹"},
	// 文件规则
	{Pattern: "*.vbproj", Type: "glob", IsDir: false, Enabled: true, Comment: "VB.NET 项目文件"},
	{Pattern: "*.vbproj.user", Type: "glob", IsDir: false, Enabled: true, Comment: "VB.NET 用户配置"},
	{Pattern: "*.csproj", Type: "glob", IsDir: false, Enabled: true, Comment: "C# 项目文件"},
	{Pattern: "*.csproj.user", Type: "glob", IsDir: false, Enabled: true, Comment: "C# 用户配置"},
	{Pattern: "*.suo", Type: "glob", IsDir: false, Enabled: true, Comment: "VS 解决方案用户选项"},
	{Pattern: "*.user", Type: "glob", IsDir: false, Enabled: true, Comment: "用户配置文件"},
	{Pattern: ".DS_Store", Type: "glob", IsDir: false, Enabled: true, Comment: "macOS 系统文件"},
	{Pattern: "Thumbs.db", Type: "glob", IsDir: false, Enabled: true, Comment: "Windows 缩略图"},
}

// Manager 配置管理器
type Manager struct {
	configPath string
	config     *models.Config
}

// NewManager 创建新的配置管理器
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, configDirName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		configPath: filepath.Join(configDir, configFileName),
		config:     &models.Config{},
	}

	// 尝试加载现有配置
	if err := m.Load(); err != nil || m.config == nil {
		// 如果配置文件不存在，使用默认配置
		m.config = &models.Config{
			ExcludeRules: defaultExcludeRules,
		}
		m.Save()
	}

	// 如果排除规则为空，使用默认规则
	if len(m.config.ExcludeRules) == 0 {
		m.config.ExcludeRules = defaultExcludeRules
		m.Save()
	}

	return m, nil
}

// Load 加载配置
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 配置文件不存在，使用默认值
		}
		return err
	}

	return json.Unmarshal(data, m.config)
}

// Save 保存配置
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Get 获取当前配置
func (m *Manager) Get() models.Config {
	if m.config == nil {
		return models.Config{}
	}
	return *m.config
}

// Set 设置配置
func (m *Manager) Set(cfg models.Config) error {
	m.config = &cfg
	return m.Save()
}

// SetLastZipPath 设置上次选择的 ZIP 文件路径
func (m *Manager) SetLastZipPath(path string) error {
	m.config.LastZipPath = path
	return m.Save()
}

// SetLastWorkDir 设置上次选择的工作目录
func (m *Manager) SetLastWorkDir(path string) error {
	m.config.LastWorkDir = path
	return m.Save()
}

// SetLastOutputDir 设置上次选择的输出目录
func (m *Manager) SetLastOutputDir(path string) error {
	m.config.LastOutputDir = path
	return m.Save()
}

// GetDefaultOutputDir 获取默认输出目录
func (m *Manager) GetDefaultOutputDir() string {
	if m.config.LastOutputDir != "" {
		return m.config.LastOutputDir
	}
	// 默认使用用户文档目录
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Documents", "Discrepancies_Output")
}

// GetExcludeRules 获取排除规则
func (m *Manager) GetExcludeRules() []models.ExcludeRule {
	if m.config == nil || len(m.config.ExcludeRules) == 0 {
		return defaultExcludeRules
	}
	return m.config.ExcludeRules
}

// SetExcludeRules 设置排除规则
func (m *Manager) SetExcludeRules(rules []models.ExcludeRule) error {
	m.config.ExcludeRules = rules
	return m.Save()
}

// AddExcludeRule 添加排除规则
func (m *Manager) AddExcludeRule(rule models.ExcludeRule) error {
	m.config.ExcludeRules = append(m.config.ExcludeRules, rule)
	return m.Save()
}

// RemoveExcludeRule 删除排除规则（按索引）
func (m *Manager) RemoveExcludeRule(index int) error {
	if index < 0 || index >= len(m.config.ExcludeRules) {
		return nil
	}
	m.config.ExcludeRules = append(m.config.ExcludeRules[:index], m.config.ExcludeRules[index+1:]...)
	return m.Save()
}

// ResetExcludeRules 重置为默认排除规则
func (m *Manager) ResetExcludeRules() error {
	m.config.ExcludeRules = defaultExcludeRules
	return m.Save()
}
