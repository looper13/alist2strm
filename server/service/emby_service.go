package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/MccRay-s/alist2strm/model/configs"
	"github.com/MccRay-s/alist2strm/model/configs/response"
	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/utils"
)

type EmbyService struct{}

// 包级别的全局实例
var Emby = &EmbyService{}

// Emby 媒体库信息
type EmbyLibrary struct {
	ID                 string      `json:"Id"`                           // 媒体库ID
	Name               string      `json:"Name"`                         // 媒体库名称
	Locations          []string    `json:"Locations"`                    // 媒体库位置列表
	CollectionType     string      `json:"CollectionType,omitempty"`     // 集合类型(movies,tvshows,music等)
	ItemId             string      `json:"ItemId,omitempty"`             // 项目ID(兼容旧版)
	Guid               string      `json:"Guid,omitempty"`               // 全局唯一标识符
	PrimaryImageItemId string      `json:"PrimaryImageItemId,omitempty"` // 主图像项目ID
	PrimaryImageTag    string      `json:"PrimaryImageTag,omitempty"`    // 主图像标签
	RefreshProgress    float64     `json:"RefreshProgress,omitempty"`    // 刷新进度
	RefreshStatus      string      `json:"RefreshStatus,omitempty"`      // 刷新状态
	ItemCount          int         `json:"ItemCount"`                    // 项目数量
	LibraryOptions     interface{} `json:"LibraryOptions,omitempty"`     // 库配置选项
	MediaType          string      `json:"Type,omitempty"`               // 媒体类型
	LastUpdate         string      `json:"LastUpdate,omitempty"`         // 最后更新时间
}

// Emby 近期入库信息
type EmbyLatestMedia struct {
	// 基础信息
	ID            string `json:"Id"`                      // 资源 ID
	Name          string `json:"Name"`                    // 资源名称
	OriginalTitle string `json:"OriginalTitle,omitempty"` // 原始标题
	Path          string `json:"Path"`                    // 文件路径
	Type          string `json:"Type"`                    // 类型
	MediaType     string `json:"MediaType,omitempty"`     // 媒体类型
	IsFolder      bool   `json:"IsFolder,omitempty"`      // 是否是文件夹

	// 时间信息
	DateCreated  time.Time `json:"DateCreated"`            // 创建日期/入库日期
	PremiereDate time.Time `json:"PremiereDate,omitempty"` // 首播日期
	EndDate      time.Time `json:"EndDate,omitempty"`      // 结束日期

	// 分类信息
	ProductionYear  int     `json:"ProductionYear,omitempty"`  // 制作年份
	OfficialRating  string  `json:"OfficialRating,omitempty"`  // 官方评级
	CommunityRating float64 `json:"CommunityRating,omitempty"` // 社区评分
	CriticRating    float64 `json:"CriticRating,omitempty"`    // 评论家评分

	// 详情
	Overview string   `json:"Overview,omitempty"` // 概述
	Genres   []string `json:"Genres,omitempty"`   // 流派
	Studios  []struct {
		Name string `json:"Name"` // 工作室名称
		ID   string `json:"Id"`   // 工作室 ID
	} `json:"Studios,omitempty"` // 工作室

	// 剧集相关
	SeriesName        string `json:"SeriesName,omitempty"`        // 剧集名称
	SeriesId          string `json:"SeriesId,omitempty"`          // 剧集ID
	SeasonId          string `json:"SeasonId,omitempty"`          // 季ID
	SeasonName        string `json:"SeasonName,omitempty"`        // 季名称
	IndexNumber       int    `json:"IndexNumber,omitempty"`       // 索引号(集号)
	ParentIndexNumber int    `json:"ParentIndexNumber,omitempty"` // 父索引号(季号)

	// 文件信息
	RunTimeTicks int64  `json:"RunTimeTicks,omitempty"` // 运行时间刻度
	Size         int64  `json:"Size,omitempty"`         // 文件大小
	Container    string `json:"Container,omitempty"`    // 容器格式

	// 状态
	LocationType string `json:"LocationType,omitempty"` // 位置类型

	// 用户数据
	UserData *struct {
		PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`    // 播放位置
		PlayCount             int    `json:"PlayCount"`                // 播放次数
		IsFavorite            bool   `json:"IsFavorite"`               // 是否为收藏
		Played                bool   `json:"Played"`                   // 是否已播放
		LastPlayedDate        string `json:"LastPlayedDate,omitempty"` // 最后播放日期
	} `json:"UserData,omitempty"` // 用户数据
}

// EmbyUser 表示Emby用户信息
type EmbyUser struct {
	ID                    string      `json:"Id"`
	Name                  string      `json:"Name"`
	HasPassword           bool        `json:"HasPassword"`
	HasConfiguredPassword bool        `json:"HasConfiguredPassword"`
	IsAdministrator       bool        `json:"IsAdministrator"` // 兼容旧版API直接返回的属性
	PrimaryImageTag       string      `json:"PrimaryImageTag,omitempty"`
	Policy                *EmbyPolicy `json:"Policy,omitempty"` // 用户策略
	ConnectUserName       string      `json:"ConnectUserName,omitempty"`
	ServerId              string      `json:"ServerId,omitempty"`
	LastActivityDate      string      `json:"LastActivityDate,omitempty"`
	LastLoginDate         string      `json:"LastLoginDate,omitempty"`
}

// EmbyPolicy 表示Emby用户策略
type EmbyPolicy struct {
	IsAdministrator       bool `json:"IsAdministrator"`       // 是否是管理员
	IsHidden              bool `json:"IsHidden"`              // 是否隐藏
	IsDisabled            bool `json:"IsDisabled"`            // 是否禁用
	EnableRemoteAccess    bool `json:"EnableRemoteAccess"`    // 是否允许远程访问
	EnableLiveTvAccess    bool `json:"EnableLiveTvAccess"`    // 是否允许直播电视访问
	EnableMediaPlayback   bool `json:"EnableMediaPlayback"`   // 是否允许媒体播放
	EnableAllChannels     bool `json:"EnableAllChannels"`     // 是否允许所有频道
	EnableAllFolders      bool `json:"EnableAllFolders"`      // 是否允许所有文件夹
	EnableAllDevices      bool `json:"EnableAllDevices"`      // 是否允许所有设备
	EnableContentDeletion bool `json:"EnableContentDeletion"` // 是否允许内容删除
}

// 获取 Emby 配置
func (s *EmbyService) getEmbyConfig() (*configs.EmbyConfig, error) {
	config, err := repository.Config.GetByCode("EMBY")
	if err != nil {
		return nil, fmt.Errorf("获取 Emby 配置失败: %w", err)
	}

	if config == nil || config.Value == "" {
		return nil, errors.New("Emby 配置不存在或为空")
	}

	var embyConfig configs.EmbyConfig
	if err := json.Unmarshal([]byte(config.Value), &embyConfig); err != nil {
		return nil, fmt.Errorf("解析 Emby 配置失败: %w", err)
	}

	// 验证配置
	if embyConfig.EmbyServer == "" || embyConfig.EmbyToken == "" {
		return nil, errors.New("Emby 服务器地址或 API 密钥未配置")
	}

	// 确保服务器地址没有尾部斜杠
	embyConfig.EmbyServer = strings.TrimRight(embyConfig.EmbyServer, "/")

	return &embyConfig, nil
}

// 发送 HTTP 请求到 Emby 服务器
func (s *EmbyService) doEmbyRequest(method, path string, body interface{}) ([]byte, error) {
	embyConfig, err := s.getEmbyConfig()
	if err != nil {
		return nil, err
	}

	// 确保路径格式正确
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Emby API 路径统一添加 /emby 前缀
	if !strings.HasPrefix(path, "/emby") {
		path = "/emby" + path
	}

	url := fmt.Sprintf("%s%s", embyConfig.EmbyServer, path)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("编码请求体失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 添加认证头
	req.Header.Set("X-Emby-Token", embyConfig.EmbyToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("emby 服务器返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应数据失败: %w", err)
	}

	return responseData, nil
}

// QueryResult 查询结果包装结构
type QueryResult struct {
	Items            []EmbyLibrary `json:"Items"`
	TotalRecordCount int           `json:"TotalRecordCount"`
}

// GetLibraries 获取 Emby 媒体库列表
func (s *EmbyService) GetLibraries() ([]EmbyLibrary, error) {
	// 使用正确的 Emby API 端点
	responseData, err := s.doEmbyRequest("GET", "/Library/VirtualFolders/Query", nil)
	if err != nil {
		return nil, err
	}

	// 尝试解析为 QueryResult 结构
	var queryResult QueryResult
	err = json.Unmarshal(responseData, &queryResult)
	if err == nil && len(queryResult.Items) > 0 {
		// 成功解析为 QueryResult
		utils.InfoLogger.Infof("成功获取到 %d 个媒体库", queryResult.TotalRecordCount)
		return queryResult.Items, nil
	}

	// 如果解析为 QueryResult 失败，尝试直接解析为数组
	var libraries []EmbyLibrary
	if err := json.Unmarshal(responseData, &libraries); err != nil {
		return nil, fmt.Errorf("解析媒体库数据失败: %w", err)
	}

	return libraries, nil
}

// RefreshLibrary 刷新指定的 Emby 媒体库
func (s *EmbyService) RefreshLibrary(libraryID string) error {
	if libraryID == "" {
		return errors.New("媒体库 ID 不能为空")
	}

	// Emby API 刷新特定媒体库的正确方法
	refreshPath := fmt.Sprintf("/Items/%s/Refresh", libraryID)
	_, err := s.doEmbyRequest("POST", refreshPath, nil)
	if err != nil {
		utils.ErrorLogger.Errorf("刷新媒体库失败: %v", err)
		return fmt.Errorf("刷新媒体库失败: %w", err)
	}

	utils.InfoLogger.Infof("成功触发媒体库 ID %s 刷新", libraryID)
	return nil
}

// RefreshAllLibraries 刷新所有 Emby 媒体库
func (s *EmbyService) RefreshAllLibraries() error {
	// 可以使用两种方式：
	// 1. 调用全局刷新API
	// 2. 遍历各个媒体库分别刷新

	// 方法1：调用全局刷新API
	_, err := s.doEmbyRequest("POST", "/Library/Refresh", nil)
	if err != nil {
		return fmt.Errorf("全局刷新媒体库失败: %w", err)
	}

	utils.InfoLogger.Info("已成功触发所有媒体库的刷新操作")

	/*
		// 方法2：逐个刷新各个媒体库（备用方法）
		libraries, err := s.GetLibraries()
		if err != nil {
			return fmt.Errorf("获取媒体库列表失败: %w", err)
		}

		for _, library := range libraries {
			if err := s.RefreshLibrary(library.ID); err != nil {
				// 记录错误但继续刷新其他媒体库
				utils.ErrorLogger.Errorf("刷新媒体库 %s (%s) 失败: %v", library.Name, library.ID, err)
			}
		}
	*/

	return nil
}

// GetLatestMedia 获取 Emby 最新入库的媒体信息
func (s *EmbyService) GetLatestMedia(limit int) ([]EmbyLatestMedia, error) {
	if limit <= 0 {
		limit = 10 // 默认获取10条
		utils.InfoLogger.Infof("使用默认限制值: %d 条记录", limit)
	}

	// 1.获取管理员用户
	adminUser, err := s.getAdminUser()
	if err != nil {
		utils.ErrorLogger.Errorf("获取管理员用户失败: %v", err)
		return nil, fmt.Errorf("获取管理员用户失败: %w", err)
	}

	// 2.使用管理员用户的 Id 进行请求
	// 3.获取最新入库的媒体信息
	// 添加Fields  ProductionYear,Path,BackdropImageTags,IndexNumber
	path := fmt.Sprintf("/Users/%s/Items/Latest?Limit=%d&Fields=ProductionYear,Path,BackdropImageTags,IndexNumber", adminUser.ID, limit)
	utils.InfoLogger.Infof("使用管理员用户 %s (ID: %s) 获取最新入库媒体，限制为 %d 条",
		adminUser.Name, adminUser.ID, limit)

	responseData, err := s.doEmbyRequest("GET", path, nil)
	if err != nil {
		utils.ErrorLogger.Errorf("使用管理员ID获取最新媒体失败: %v", err)
		return nil, fmt.Errorf("使用管理员ID获取最新媒体失败: %w", err)
	}
	utils.InfoLogger.Info("成功获取最新入库媒体数据", string(responseData))
	var latestMedia []EmbyLatestMedia
	if err := json.Unmarshal(responseData, &latestMedia); err != nil {
		utils.ErrorLogger.Errorf("解析最新媒体数据失败: %v, 响应数据: %s", err, string(responseData))
		return nil, fmt.Errorf("解析最新媒体数据失败: %w", err)
	}

	// 记录获取到的媒体数量
	utils.InfoLogger.Infof("通过管理员用户成功获取 %d 条最新入库媒体", len(latestMedia))

	// 处理路径映射
	embyConfig, err := s.getEmbyConfig()
	if err != nil {
		utils.WarnLogger.Warnf("获取Emby配置失败，无法进行路径映射: %v", err)
	} else if embyConfig != nil && len(embyConfig.PathMappings) > 0 {
		originalPaths := make([]string, len(latestMedia))
		for i := range latestMedia {
			originalPaths[i] = latestMedia[i].Path
			latestMedia[i].Path = s.mapEmbyPathToLocal(latestMedia[i].Path, embyConfig.PathMappings)
			if originalPaths[i] != latestMedia[i].Path {
				utils.DebugLogger.Debugf("路径映射: %s -> %s", originalPaths[i], latestMedia[i].Path)
			}
		}
	}

	return latestMedia, nil
}

// mapEmbyPathToLocal 将Emby路径映射到本地路径
func (s *EmbyService) mapEmbyPathToLocal(embyPath string, mappings []configs.PathMapping) string {
	if embyPath == "" || len(mappings) == 0 {
		return embyPath
	}

	// 规范化路径分隔符
	embyPath = filepath.ToSlash(embyPath)

	// 尝试每个映射
	for _, mapping := range mappings {
		if mapping.EmbyPath != "" && mapping.Path != "" && strings.HasPrefix(embyPath, mapping.EmbyPath) {
			// 替换前缀
			localPath := strings.Replace(embyPath, mapping.EmbyPath, mapping.Path, 1)
			return localPath
		}
	}

	return embyPath
}

// MapLocalPathToEmby 将本地路径映射到Emby路径
// 导出此函数使其可以被其他包使用，避免未使用警告
func (s *EmbyService) MapLocalPathToEmby(localPath string) (string, error) {
	if localPath == "" {
		return "", nil
	}

	embyConfig, err := s.getEmbyConfig()
	if err != nil {
		return localPath, err
	}

	// 规范化路径分隔符
	localPath = filepath.ToSlash(localPath)

	// 尝试每个映射
	for _, mapping := range embyConfig.PathMappings {
		if mapping.Path != "" && mapping.EmbyPath != "" && strings.HasPrefix(localPath, mapping.Path) {
			// 替换前缀
			embyPath := strings.Replace(localPath, mapping.Path, mapping.EmbyPath, 1)
			return embyPath, nil
		}
	}

	return localPath, nil
}

// GetEmbySystemInfo 获取Emby系统信息
func (s *EmbyService) GetEmbySystemInfo() (map[string]interface{}, error) {
	// 使用正确的 API 端点
	responseData, err := s.doEmbyRequest("GET", "/System/Info", nil)
	if err != nil {
		return nil, err
	}

	var systemInfo map[string]interface{}
	if err := json.Unmarshal(responseData, &systemInfo); err != nil {
		return nil, fmt.Errorf("解析系统信息失败: %w", err)
	}

	return systemInfo, nil
}

// TestConnection 测试与Emby服务器的连接
func (s *EmbyService) TestConnection() (*response.EmbyConnectionTestResult, error) {
	result := &response.EmbyConnectionTestResult{
		Connected: false,
	}

	_, err := s.getEmbyConfig()
	if err != nil {
		result.Error = err.Error()
		return result, nil // 返回结果但不返回错误
	}

	systemInfo, err := s.GetEmbySystemInfo()
	if err != nil {
		result.Error = fmt.Sprintf("连接Emby服务器失败: %s", err.Error())
		return result, nil
	}

	result.Connected = true
	result.Version = fmt.Sprintf("%v", systemInfo["Version"])
	result.ServerName = fmt.Sprintf("%v", systemInfo["ServerName"])

	// 尝试获取系统信息中可能存在的其他字段
	if operatingSystem, ok := systemInfo["OperatingSystem"].(string); ok {
		result.OperatingSystem = operatingSystem
	}

	return result, nil
}

// GetImage 获取 Emby 图片
// itemId: 项目ID
// imageType: 图片类型，如 Primary, Backdrop 等
// tag: 可选的图片标签
// maxWidth: 可选的最大宽度
// maxHeight: 可选的最大高度
// quality: 可选的质量参数
func (s *EmbyService) GetImage(itemId, imageType string, tag string, maxWidth, maxHeight, quality int) ([]byte, string, error) {
	if itemId == "" || imageType == "" {
		return nil, "", fmt.Errorf("项目ID和图片类型不能为空")
	}

	// 构建图片请求路径
	path := fmt.Sprintf("/Items/%s/Images/%s", itemId, imageType)

	// 添加可选参数
	params := make([]string, 0)
	if tag != "" {
		params = append(params, fmt.Sprintf("tag=%s", tag))
	}
	if maxWidth > 0 {
		params = append(params, fmt.Sprintf("maxWidth=%d", maxWidth))
	}
	if maxHeight > 0 {
		params = append(params, fmt.Sprintf("maxHeight=%d", maxHeight))
	}
	if quality > 0 {
		params = append(params, fmt.Sprintf("quality=%d", quality))
	}

	if len(params) > 0 {
		path = path + "?" + strings.Join(params, "&")
	}

	// 发送请求
	embyConfig, err := s.getEmbyConfig()
	if err != nil {
		return nil, "", err
	}

	// 确保路径格式正确
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Emby API 路径统一添加 /emby 前缀
	if !strings.HasPrefix(path, "/emby") {
		path = "/emby" + path
	}

	url := fmt.Sprintf("%s%s", embyConfig.EmbyServer, path)

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 添加认证头
	req.Header.Set("X-Emby-Token", embyConfig.EmbyToken)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("emby 服务器返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取图片数据失败: %w", err)
	}

	// 获取内容类型
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// 如果服务器没有返回内容类型，尝试根据图片数据猜测
		contentType = http.DetectContentType(imageData)
	}

	return imageData, contentType, nil
}

// EmbyQueryResult 表示Emby用户查询结果
type EmbyQueryResult struct {
	Items            []EmbyUser `json:"Items"`
	TotalRecordCount int        `json:"TotalRecordCount"`
}

// getAdminUser 获取管理员用户
func (s *EmbyService) getAdminUser() (*EmbyUser, error) {
	// 调用Emby API获取用户列表
	// 参考: https://dev.emby.media/reference/RestAPI/UserService/getUsersQuery.html
	path := "/Users/Query"
	responseData, err := s.doEmbyRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 解析用户查询结果
	var queryResult EmbyQueryResult
	if err := json.Unmarshal(responseData, &queryResult); err != nil {
		// 如果解析成 QueryResult 失败，尝试直接解析为数组（兼容旧版API）
		var users []EmbyUser
		if jsonErr := json.Unmarshal(responseData, &users); jsonErr != nil {
			return nil, fmt.Errorf("解析用户列表失败: %w", err)
		}

		// 查找管理员用户
		for _, user := range users {
			// 检查用户是否为管理员，处理 Policy 字段可能为空的情况
			isAdmin := user.IsAdministrator || (user.Policy != nil && user.Policy.IsAdministrator)
			if isAdmin {
				utils.InfoLogger.Infof("找到Emby管理员用户: %s (ID: %s)", user.Name, user.ID)
				return &user, nil
			}
		}

		return nil, fmt.Errorf("未找到管理员用户")
	}

	// 查找管理员用户
	for _, user := range queryResult.Items {
		// 检查用户是否为管理员，处理 Policy 字段可能为空的情况
		isAdmin := user.IsAdministrator || (user.Policy != nil && user.Policy.IsAdministrator)
		if isAdmin {
			utils.InfoLogger.Infof("找到Emby管理员用户: %s (ID: %s)", user.Name, user.ID)
			return &user, nil
		}
	}

	return nil, fmt.Errorf("未找到管理员用户")
}
