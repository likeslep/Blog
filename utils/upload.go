// utils/upload.go
package utils

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadConfig 上传配置
type UploadConfig struct {
	UploadPath  string   // 上传路径
	MaxSize     int64    // 最大文件大小（字节）
	AllowTypes  []string // 允许的文件类型
	Thumbnail   bool     // 是否生成缩略图
	ThumbWidth  int      // 缩略图宽度
	ThumbHeight int      // 缩略图高度
}

// UploadResult 上传结果
type UploadResult struct {
	Filename     string `json:"filename"`
	OriginalName string `json:"original_name"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	Type         string `json:"type"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	ThumbPath    string `json:"thumb_path,omitempty"`
	URL          string `json:"url"`
}

// UploadFile 上传文件
func UploadFile(file *multipart.FileHeader, config UploadConfig, userID uint) (*UploadResult, error) {
	// 1. 验证文件大小
	if file.Size > config.MaxSize {
		return nil, fmt.Errorf("文件大小超过限制（最大 %d MB）", config.MaxSize/1024/1024)
	}

	// 2. 检测文件类型
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	mime, err := mimetype.DetectReader(src)
	if err != nil {
		return nil, err
	}

	// 重置文件指针
	src.Seek(0, io.SeekStart)

	mimeType := mime.String()

	// 3. 验证文件类型
	allowed := false
	fileType := "file"
	for _, allowType := range config.AllowTypes {
		if strings.HasPrefix(mimeType, allowType) {
			allowed = true
			if strings.HasPrefix(mimeType, "image/") {
				fileType = "image"
			} else if strings.HasPrefix(mimeType, "video/") {
				fileType = "video"
			}
			break
		}
	}

	if !allowed {
		return nil, fmt.Errorf("不支持的文件类型: %s", mimeType)
	}

	// 4. 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	uniqueName := uuid.New().String() + ext
	relativePath := filepath.Join("uploads", time.Now().Format("2006/01"), uniqueName)
	fullPath := filepath.Join(config.UploadPath, relativePath)

	// 5. 创建目录
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	// 6. 保存文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	result := &UploadResult{
		Filename:     uniqueName,
		OriginalName: file.Filename,
		Path:         relativePath,
		Size:         file.Size,
		MimeType:     mimeType,
		Type:         fileType,
	}

	// 7. 如果是图片，获取尺寸并生成缩略图
	if fileType == "image" && config.Thumbnail {
		// 获取图片尺寸
		src.Seek(0, io.SeekStart)
		img, format, err := image.Decode(src)
		if err == nil {
			bounds := img.Bounds()
			result.Width = bounds.Dx()
			result.Height = bounds.Dy()

			// 生成缩略图
			thumbName := "thumb_" + uniqueName
			thumbPath := filepath.Join(filepath.Dir(fullPath), thumbName)
			thumbRelative := filepath.Join(filepath.Dir(relativePath), thumbName)

			// 创建缩略图
			thumb := imaging.Thumbnail(img, config.ThumbWidth, config.ThumbHeight, imaging.Lanczos)

			// 保存缩略图
			thumbFile, err := os.Create(thumbPath)
			if err == nil {
				defer thumbFile.Close()

				// 根据原图格式保存
				switch format {
				case "jpeg":
					err = jpeg.Encode(thumbFile, thumb, &jpeg.Options{Quality: 85})
				case "png":
					err = imaging.Encode(thumbFile, thumb, imaging.PNG)
				default:
					err = imaging.Encode(thumbFile, thumb, imaging.JPEG)
				}

				if err == nil {
					result.ThumbPath = thumbRelative
				}
			}
		}
	}

	return result, nil
}

// DeleteFile 删除文件
func DeleteFile(path string) error {
	// 删除原文件
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	// 尝试删除缩略图
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	thumbPath := filepath.Join(dir, "thumb_"+filename)
	os.Remove(thumbPath)

	return nil
}

// GetFileURL 获取文件访问URL
func GetFileURL(path string, baseURL string) string {
	// 将路径中的反斜杠替换为正斜杠
	path = strings.ReplaceAll(path, "\\", "/")
	return baseURL + "/" + path
}

// GetFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
