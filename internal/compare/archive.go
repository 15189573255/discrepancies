package compare

import (
	"archive/zip"
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// ZipEntry 表示 ZIP 文件中的一个条目
type ZipEntry struct {
	RelPath string // 相对路径
	IsDir   bool   // 是否是目录
	Size    int64  // 文件大小
}

// ZipReader 封装 ZIP 读取操作
type ZipReader struct {
	path   string
	reader *zip.ReadCloser
}

// NewZipReader 创建新的 ZIP 读取器
func NewZipReader(zipPath string) (*ZipReader, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	return &ZipReader{path: zipPath, reader: reader}, nil
}

// Close 关闭 ZIP 读取器
func (z *ZipReader) Close() error {
	if z.reader != nil {
		return z.reader.Close()
	}
	return nil
}

// GetRootFolder 获取 ZIP 中的根文件夹名称
// 通常 ZIP 文件会有一个根目录，例如 project-v1.0/
func (z *ZipReader) GetRootFolder() string {
	if len(z.reader.File) == 0 {
		return ""
	}

	// 获取第一个条目的路径
	firstPath := z.reader.File[0].Name
	parts := strings.Split(strings.TrimPrefix(firstPath, "/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ListFiles 列出 ZIP 中的所有文件（不包含目录）
// 返回相对于根目录的路径
func (z *ZipReader) ListFiles() (map[string]*zip.File, error) {
	files := make(map[string]*zip.File)
	rootFolder := z.GetRootFolder()

	for _, f := range z.reader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// 获取相对路径（去除根目录前缀）
		relPath := f.Name
		if rootFolder != "" && strings.HasPrefix(relPath, rootFolder+"/") {
			relPath = strings.TrimPrefix(relPath, rootFolder+"/")
		}

		// 统一使用正斜杠
		relPath = filepath.ToSlash(relPath)
		if relPath != "" {
			files[relPath] = f
		}
	}

	return files, nil
}

// ListDirs 列出 ZIP 中的所有目录
func (z *ZipReader) ListDirs() (map[string]bool, error) {
	dirs := make(map[string]bool)
	rootFolder := z.GetRootFolder()

	for _, f := range z.reader.File {
		if !f.FileInfo().IsDir() {
			continue
		}

		relPath := strings.TrimSuffix(f.Name, "/")
		if rootFolder != "" && strings.HasPrefix(relPath, rootFolder+"/") {
			relPath = strings.TrimPrefix(relPath, rootFolder+"/")
		} else if relPath == rootFolder {
			continue // 跳过根目录本身
		}

		relPath = filepath.ToSlash(relPath)
		if relPath != "" {
			dirs[relPath] = true
		}
	}

	return dirs, nil
}

// GetFileHash 计算 ZIP 中指定文件的 MD5 哈希
func (z *ZipReader) GetFileHash(relPath string) ([]byte, error) {
	files, err := z.ListFiles()
	if err != nil {
		return nil, err
	}

	f, exists := files[relPath]
	if !exists {
		return nil, fmt.Errorf("file not found in zip: %s", relPath)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer rc.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, rc); err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hash.Sum(nil), nil
}

// ReadFileContent 读取 ZIP 中指定文件的内容
func (z *ZipReader) ReadFileContent(relPath string) ([]byte, error) {
	files, err := z.ListFiles()
	if err != nil {
		return nil, err
	}

	f, exists := files[relPath]
	if !exists {
		return nil, fmt.Errorf("file not found in zip: %s", relPath)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// GetFileSize 获取 ZIP 中指定文件的大小
func (z *ZipReader) GetFileSize(relPath string) (int64, error) {
	files, err := z.ListFiles()
	if err != nil {
		return 0, err
	}

	f, exists := files[relPath]
	if !exists {
		return 0, fmt.Errorf("file not found in zip: %s", relPath)
	}

	return int64(f.UncompressedSize64), nil
}
