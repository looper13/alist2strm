package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

//go:generate stringer -type=MediaType

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeMovie  MediaType = "movie"  // 电影
	MediaTypeTVShow MediaType = "tvshow" // 电视剧
)

// 常用语言代码
var LanguageCodes = []string{
	"zh", "cn", "zh-cn", "zh-tw", "cht", "chs", // 中文
	"en", "eng", // 英文
	"ja", "jpn", // 日文
	"ko", "kor", // 韩文
	"fr", "fra", // 法语
	"de", "ger", // 德语
	"es", "spa", // 西班牙语
	"ru", "rus", // 俄语
	"pt", "por", // 葡萄牙语
	"it", "ita", // 意大利语
}

// MovieMetadata 电影元数据文件规则
var MovieMetadata = []string{
	"movie.nfo",  // 基本信息
	"poster.jpg", // 海报
	"poster.png",
	"backdrop.jpg", // 背景图
	"backdrop.png",
	"banner.jpg", // 横幅
	"banner.png",
	"clearart.png", // 透明背景艺术图
	"logo.png",     // 标志
	"disc.png",     // 光盘封面
	"thumb.jpg",    // 缩略图
	"thumb.png",
	"fanart.jpg", // 粉丝艺术图
	"fanart.png",
}

// TVShowMetadata 剧集元数据文件规则
var TVShowMetadata = map[string][]string{
	"show": {
		"tvshow.nfo", // 整个剧集信息
		"poster.jpg", // 剧集海报
		"poster.png",
		"banner.jpg", // 剧集横幅
		"banner.png",
		"fanart.jpg", // 剧集背景图
		"fanart.png",
	},
	"season": {
		"season.nfo",        // 季信息
		"season-poster.jpg", // 季海报
		"season-poster.png",
		"season-banner.jpg", // 季横幅
		"season-banner.png",
		"season-fanart.jpg", // 季背景图
		"season-fanart.png",
	},
	"episode": {
		".nfo",        // 单集信息（前缀为具体集数）
		"-thumb.jpg",  // 单集缩略图
		"-fanart.jpg", // 单集背景图
	},
}

// TaskConfig 任务配置
type TaskConfig struct {
	SourcePath string    `json:"sourcePath"` // Alist 中的源路径
	TargetPath string    `json:"targetPath"` // 本地目标路径
	MediaType  MediaType `json:"mediaType"`  // 媒体类型：movie 或 tvshow

	// 文件处理选项
	Overwrite  bool     `json:"overwrite"`            // 是否覆盖已存在的文件
	FileSuffix []string `json:"fileSuffix,omitempty"` // 视频文件后缀名列表

	// 元数据处理选项
	DownloadMetadata   bool     `json:"downloadMetadata"`   // 是否下载元数据
	MetadataExtensions []string `json:"metadataExtensions"` // 元数据文件扩展名列表

	// 字幕处理选项
	DownloadSubtitle   bool     `json:"downloadSubtitle"`   // 是否下载字幕
	SubtitleExtensions []string `json:"subtitleExtensions"` // 字幕文件扩展名列表
}

// FileService 文件服务，处理strm文件生成、元数据和字幕文件下载等
type FileService struct {
	alistClient *AlistClient
	logger      *zap.Logger
	config      *TaskConfig // 当前任务配置
}

// NewFileService 创建FileService实例
func NewFileService(logger *zap.Logger) *FileService {
	return &FileService{
		alistClient: GetAlistClient(logger),
		logger:      logger,
	}
}

// SetTaskConfig 设置任务配置
func (s *FileService) SetTaskConfig(config *TaskConfig) {
	s.config = config
}

// GenerateStrm 生成strm文件
func (s *FileService) GenerateStrm(sourcePath, filename, targetSubPath, sign string) error {
	if s.config == nil {
		return fmt.Errorf("任务配置未设置")
	}

	targetDir := filepath.Join(s.config.TargetPath, targetSubPath)

	// 确保目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建strm目录失败: %w", err)
	}

	// 构建strm文件内容（播放地址，包含签名如果有的话）
	content := s.alistClient.GetFileURL(sourcePath, filename, sign)

	// 创建strm文件
	baseName := strings.TrimSuffix(filename, filepath.Ext(filename))
	strmFile := filepath.Join(targetDir, baseName+".strm")

	// 检查文件是否存在且不需要覆盖
	if !s.config.Overwrite {
		if _, err := os.Stat(strmFile); err == nil {
			s.logger.Debug("strm文件已存在且不需要覆盖", zap.String("file", strmFile))
			return nil
		}
	}

	if err := os.WriteFile(strmFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入strm文件失败: %w", err)
	}

	return nil
}

// DownloadMetadata 下载并保存元数据文件
func (s *FileService) DownloadMetadata(sourcePath, filename, targetSubPath, sign string) error {
	if s.config == nil {
		return fmt.Errorf("任务配置未设置")
	}

	if !s.config.DownloadMetadata {
		return nil
	}

	// 检查文件扩展名是否匹配
	extWithoutDot := s.getExtWithoutDot(filename)
	matched := false
	for _, allowedExt := range s.config.MetadataExtensions {
		// 将不带点号的扩展名与配置的扩展名比较
		allowedExtWithoutDot := strings.TrimPrefix(s.normalizeExtension(allowedExt), ".")
		if strings.EqualFold(extWithoutDot, allowedExtWithoutDot) {
			matched = true
			break
		}
	}
	if !matched {
		s.logger.Debug("跳过不匹配的元数据文件",
			zap.String("file", filename),
			zap.String("ext", "."+extWithoutDot))
		return nil
	}

	var targetDir string
	if targetSubPath == "." {
		// 如果是根目录，直接使用目标路径
		targetDir = s.config.TargetPath
	} else {
		targetDir = filepath.Join(s.config.TargetPath, targetSubPath)
	}

	// 确保目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建元数据目录失败: %w", err)
	}

	targetFile := filepath.Join(targetDir, filename)

	// 检查文件是否存在且不需要覆盖
	if !s.config.Overwrite {
		if _, err := os.Stat(targetFile); err == nil {
			s.logger.Debug("元数据文件已存在且不需要覆盖", zap.String("file", targetFile))
			return nil
		}
	}

	// 构建下载URL（包含签名如果有的话）
	url := s.alistClient.GetFileURL(sourcePath, filename, sign)

	// 下载文件
	if err := s.downloadFile(url, targetFile); err != nil {
		return fmt.Errorf("下载元数据文件失败: %w", err)
	}

	return nil
}

// DownloadSubtitle 下载并保存字幕文件
func (s *FileService) DownloadSubtitle(sourcePath, filename, targetSubPath, sign string) error {
	if s.config == nil {
		return fmt.Errorf("任务配置未设置")
	}

	if !s.config.DownloadSubtitle {
		return nil
	}

	// 检查文件扩展名是否匹配
	extWithoutDot := s.getExtWithoutDot(filename)
	matched := false
	for _, allowedExt := range s.config.SubtitleExtensions {
		// 将不带点号的扩展名与配置的扩展名比较
		allowedExtWithoutDot := strings.TrimPrefix(s.normalizeExtension(allowedExt), ".")
		if strings.EqualFold(extWithoutDot, allowedExtWithoutDot) {
			matched = true
			break
		}
	}
	if !matched {
		s.logger.Debug("跳过不匹配的字幕文件",
			zap.String("file", filename),
			zap.String("ext", "."+extWithoutDot))
		return nil
	}

	targetDir := filepath.Join(s.config.TargetPath, targetSubPath)

	// 确保目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建字幕目录失败: %w", err)
	}

	targetFile := filepath.Join(targetDir, filename)

	// 检查文件是否存在且不需要覆盖
	if !s.config.Overwrite {
		if _, err := os.Stat(targetFile); err == nil {
			s.logger.Debug("字幕文件已存在且不需要覆盖", zap.String("file", targetFile))
			return nil
		}
	}

	// 构建下载URL（包含签名如果有的话）
	url := s.alistClient.GetFileURL(sourcePath, filename, sign)

	// 下载文件
	if err := s.downloadFile(url, targetFile); err != nil {
		return fmt.Errorf("下载字幕文件失败: %w", err)
	}

	return nil
}

// downloadFile 通用的文件下载函数
// isSubtitleExt 检查是否是字幕文件扩展名
func (s *FileService) isSubtitleExt(ext string) bool {
	for _, allowedExt := range s.config.SubtitleExtensions {
		if s.normalizeExtension(ext) == s.normalizeExtension(allowedExt) {
			return true
		}
	}
	return false
}

// normalizeExtension 标准化文件扩展名（确保以点号开头）
func (s *FileService) normalizeExtension(ext string) string {
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}

// getExtWithoutDot 获取不带点号的扩展名
func (s *FileService) getExtWithoutDot(filename string) string {
	return strings.TrimPrefix(filepath.Ext(filename), ".")
}

func (s *FileService) downloadFile(url, destPath string) error {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", s.alistClient.config.Token)

	// 发送请求
	resp, err := s.alistClient.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建目标文件
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 读取响应内容并写入文件
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, &buf)
	return err
}

// ProcessDirectory 处理目录，生成strm文件并下载相关文件
func (s *FileService) ProcessDirectory(sourcePath string) error {
	if s.config == nil {
		return fmt.Errorf("任务配置未设置")
	}
	// log
	s.logger.Info("开始处理目录",
		zap.String("sourcePath", sourcePath),
		zap.String("targetPath", s.config.TargetPath),
		zap.String("mediaType", string(s.config.MediaType)),
		zap.Bool("downloadMetadata", s.config.DownloadMetadata),
		zap.Bool("downloadSubtitle", s.config.DownloadSubtitle))

	files, err := s.alistClient.ListFiles(sourcePath)
	if err != nil {
		return fmt.Errorf("获取目录文件列表失败: %w", err)
	}

	// 计算相对于 SourcePath 的路径
	relPath := ""
	if s.config.SourcePath == "" {
		s.logger.Error("未设置 SourcePath，将使用完整路径",
			zap.String("sourcePath", sourcePath))
		relPath = sourcePath
	} else {
		// 确保 sourcePath 是以 SourcePath 开头的
		if !strings.HasPrefix(sourcePath, s.config.SourcePath) {
			s.logger.Error("源文件路径不在配置的基础路径下",
				zap.String("sourcePath", sourcePath),
				zap.String("configSourcePath", s.config.SourcePath))
			return fmt.Errorf("源文件路径 %s 不在配置的基础路径 %s 下", sourcePath, s.config.SourcePath)
		}

		// 获取相对路径部分（去掉 SourcePath 前缀）
		relPath = strings.TrimPrefix(sourcePath, s.config.SourcePath)
	}

	// 去掉开头的斜杠并规范化路径分隔符
	relPath = strings.TrimPrefix(relPath, "/")
	// 如果结果是空字符串，就保持为空（代表根目录）
	if relPath != "" {
		relPath = filepath.FromSlash(relPath)
	}

	for _, file := range files {
		if file.IsDir {
			// 递归处理子目录
			subPath := filepath.Join(sourcePath, file.Name)
			if err := s.ProcessDirectory(subPath); err != nil {
				s.logger.Error("处理子目录失败",
					zap.String("path", subPath),
					zap.Error(err))
			}
			continue
		}

		// 获取文件名相关信息
		baseName := strings.TrimSuffix(file.Name, filepath.Ext(file.Name))
		extWithoutDot := s.getExtWithoutDot(file.Name)

		// 检查是否是支持的视频文件类型
		isVideo := false
		for _, suffix := range s.config.FileSuffix {
			suffixWithoutDot := strings.TrimPrefix(s.normalizeExtension(suffix), ".")
			if strings.EqualFold(extWithoutDot, suffixWithoutDot) {
				isVideo = true
				break
			}
		}

		// 处理视频文件，生成strm文件
		if isVideo {
			// log
			s.logger.Info("处理视频文件",
				zap.String("path", sourcePath),
				zap.String("file", file.Name),
				zap.String("relPath", relPath))
			if err := s.GenerateStrm(sourcePath, file.Name, relPath, file.Sign); err != nil {
				s.logger.Error("生成strm文件失败",
					zap.String("path", sourcePath),
					zap.String("file", file.Name),
					zap.Error(err))
			}
		}

		// 检查并下载字幕文件
		if s.config.DownloadSubtitle {
			// log
			s.logger.Info("处理字幕文件",
				zap.String("path", sourcePath),
				zap.String("file", file.Name),
				zap.String("relPath", relPath))
			// 如果当前文件是字幕文件，直接下载
			// 使用 getExtWithoutDot 处理扩展名
			extWithoutDot := s.getExtWithoutDot(file.Name)
			if s.isSubtitleExt("." + extWithoutDot) {
				if err := s.DownloadSubtitle(sourcePath, file.Name, relPath, file.Sign); err != nil {
					s.logger.Error("下载字幕文件失败",
						zap.String("path", sourcePath),
						zap.String("file", file.Name),
						zap.Error(err))
				}
			}

			// 如果是视频文件，尝试下载关联的字幕
			if isVideo {
				// 尝试下载不带语言代码的字幕
				for _, subtitleExt := range s.config.SubtitleExtensions {
					subtitleName := baseName + subtitleExt
					if err := s.DownloadSubtitle(sourcePath, subtitleName, relPath, ""); err != nil {
						s.logger.Debug("尝试下载无语言代码字幕失败",
							zap.String("path", sourcePath),
							zap.String("file", subtitleName))
					}
				}
				// 尝试下载带语言代码的字幕
				for _, lang := range LanguageCodes {
					for _, subtitleExt := range s.config.SubtitleExtensions {
						subtitleName := fmt.Sprintf("%s.%s%s", baseName, lang, subtitleExt)
						if err := s.DownloadSubtitle(sourcePath, subtitleName, relPath, ""); err != nil {
							s.logger.Debug("尝试下载带语言字幕失败",
								zap.String("path", sourcePath),
								zap.String("file", subtitleName),
								zap.String("lang", lang))
						}
					}
				}
			}
		}

		// 根据媒体类型处理元数据
		if s.config.DownloadMetadata {
			// 计算当前文件的相对路径
			fileRelPath := relPath
			if fileRelPath == "" {
				// 如果在根目录，使用文件名作为相对路径
				fileRelPath = "."
			}

			switch s.config.MediaType {
			case MediaTypeMovie:
				for _, metaFile := range MovieMetadata {
					// 构建元数据文件名
					metaFileName := metaFile
					// 检查元数据文件是否存在
					if file.Name == metaFileName {
						s.logger.Debug("发现电影元数据文件",
							zap.String("sourcePath", sourcePath),
							zap.String("file", file.Name),
							zap.String("relPath", fileRelPath))
						if err := s.DownloadMetadata(sourcePath, file.Name, fileRelPath, file.Sign); err != nil {
							s.logger.Error("下载电影元数据文件失败",
								zap.String("path", sourcePath),
								zap.String("file", file.Name),
								zap.Error(err))
						}
					}
				}

			case MediaTypeTVShow:
				// 检查是否为剧集级别的元数据
				for _, metaFile := range TVShowMetadata["show"] {
					if file.Name == metaFile {
						s.logger.Debug("发现剧集元数据文件",
							zap.String("sourcePath", sourcePath),
							zap.String("file", file.Name),
							zap.String("relPath", fileRelPath))
						if err := s.DownloadMetadata(sourcePath, file.Name, fileRelPath, file.Sign); err != nil {
							s.logger.Error("下载剧集元数据文件失败",
								zap.String("path", sourcePath),
								zap.String("file", file.Name),
								zap.Error(err))
						}
					}
				}

				// 检查是否为季级别的元数据
				for _, metaFile := range TVShowMetadata["season"] {
					if file.Name == metaFile {
						s.logger.Debug("发现季度元数据文件",
							zap.String("sourcePath", sourcePath),
							zap.String("file", file.Name),
							zap.String("relPath", fileRelPath))
						if err := s.DownloadMetadata(sourcePath, file.Name, fileRelPath, file.Sign); err != nil {
							s.logger.Error("下载季度元数据文件失败",
								zap.String("path", sourcePath),
								zap.String("file", file.Name),
								zap.Error(err))
						}
					}
				}

				// 检查是否为单集元数据
				for _, metaFile := range TVShowMetadata["episode"] {
					if strings.HasSuffix(file.Name, metaFile) {
						s.logger.Debug("发现单集元数据文件",
							zap.String("sourcePath", sourcePath),
							zap.String("file", file.Name),
							zap.String("relPath", fileRelPath))
						if err := s.DownloadMetadata(sourcePath, file.Name, fileRelPath, file.Sign); err != nil {
							s.logger.Error("下载单集元数据文件失败",
								zap.String("path", sourcePath),
								zap.String("file", file.Name),
								zap.Error(err))
						}
					}
				}
			}
		}
	}

	return nil
}
