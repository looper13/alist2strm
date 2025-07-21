package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/MccRay-s/alist2strm/clouddrive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

const DEFAULT_BUFFER_SIZE = 8192 // 默认缓冲区大小

type Client struct {
	addr              string
	conn              *grpc.ClientConn
	cd                clouddrive.CloudDriveFileSrvClient
	contextWithHeader context.Context
	username          string
	password          string
	OfflineFolder     string
	UploadFolder      string
}

func NewClient(addr, username, password, offlineFolder, uploadFolder string) *Client {
	c := Client{
		addr:              addr,
		conn:              nil,
		cd:                nil,
		contextWithHeader: nil,
		username:          username,
		password:          password,
		OfflineFolder:     offlineFolder,
		UploadFolder:      uploadFolder,
	}
	return &c
}

func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *Client) Login() error {
	conn, err := grpc.Dial(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.conn = conn
	c.cd = clouddrive.NewCloudDriveFileSrvClient(c.conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	r, err := c.cd.GetToken(ctx, &clouddrive.GetTokenRequest{UserName: c.username, Password: c.password})
	if err != nil {
		return err
	}
	header := metadata.New(map[string]string{
		"authorization": "Bearer " + r.GetToken(),
	})
	c.contextWithHeader = metadata.NewOutgoingContext(context.Background(), header)
	return nil
}

func (c *Client) Set115Cookie(ck string) error {
	res, err := c.cd.APILogin115Editthiscookie(c.contextWithHeader, &clouddrive.Login115EditthiscookieRequest{EditThiscookieString: ck})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) Get115QrCode(platformString string) (string, error) {
	res, err := c.cd.APILogin115QRCode(c.contextWithHeader, &clouddrive.Login115QrCodeRequest{PlatformString: &platformString})
	if err != nil {
		return "", err
	}
	msg, err := res.Recv()
	if err != nil {
		return "", err
	}
	return msg.GetMessage(), nil
}

func (c *Client) AddOfflineFiles(url string) ([]string, error) {
	res, err := c.cd.AddOfflineFiles(c.contextWithHeader, &clouddrive.AddOfflineFileRequest{Urls: url, ToFolder: c.OfflineFolder})
	if err != nil {
		return nil, err
	}
	if !res.Success {
		return nil, errors.New(res.ErrorMessage)
	}
	return res.GetResultFilePaths(), nil
}

func (c *Client) ListAllOfflineFiles(cloudName, cloudAccountId string, page uint32) ([]*clouddrive.OfflineFile, error) {
	res, err := c.cd.ListAllOfflineFiles(c.contextWithHeader, &clouddrive.OfflineFileListAllRequest{
		CloudName:      cloudName,
		CloudAccountId: cloudAccountId,
		Page:           page,
	})
	if err != nil {
		return nil, err
	}
	return res.GetOfflineFiles(), nil
}

func (c *Client) Upload(filePath, fileName string) error {
	var createFileResult *clouddrive.CreateFileResult
	var file *os.File
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	defer func() {
		if file != nil {
			_ = file.Close()
		}
		if createFileResult != nil {
			_, _ = c.cd.CloseFile(c.contextWithHeader, &clouddrive.CloseFileRequest{FileHandle: createFileResult.FileHandle})
		}
	}()
	createFileResult, err := c.cd.CreateFile(c.contextWithHeader, &clouddrive.CreateFileRequest{ParentPath: c.UploadFolder, FileName: fileName})
	if err != nil {
		return err
	}
	// 打开文件
	file, err = os.Open(filePath)
	if err != nil {
		return err
	}
	// 如果传入了文件对象
	if file != nil {
		offset := uint64(0)
		// 循环读取文件内容并写入到云端文件
		for {
			reader := bufio.NewReader(file)
			data := make([]byte, DEFAULT_BUFFER_SIZE)
			n, err := reader.Read(data)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}
			// 将文件内容写入到云端文件
			_, err = c.cd.WriteToFile(c.contextWithHeader, &clouddrive.WriteFileRequest{FileHandle: createFileResult.FileHandle, StartPos: offset, Length: uint64(n), Buffer: data[:n], CloseFile: false})
			if err != nil {
				return err
			}
			//fmt.Println(res.GetBytesWritten())
			offset += uint64(n)
		}
	}

	return nil
}

func (c *Client) GetSubFiles(path string, forceRefresh bool, checkExpires bool) (*clouddrive.SubFilesReply, error) {

	res, err := c.cd.GetSubFiles(c.contextWithHeader, &clouddrive.ListSubFileRequest{Path: path, ForceRefresh: forceRefresh, CheckExpires: &checkExpires})
	if err != nil {
		return nil, err
	}
	subFilesReply, err := res.Recv()
	if err != nil {
		return nil, err
	}
	return subFilesReply, err
}

// GetImmediateSubFiles getImmediateSubFiles 辅助函数：获取指定路径下的所有直接子文件和子目录。
// 它处理云盘服务可能返回的流式数据。
func (c *Client) GetImmediateSubFiles(path string, forceRefresh bool, checkExpires bool) ([]*clouddrive.CloudDriveFile, error) {
	// 调用云盘服务API，获取子文件流
	stream, err := c.cd.GetSubFiles(c.contextWithHeader, &clouddrive.ListSubFileRequest{
		Path:         path,          // 要查询的目录路径
		ForceRefresh: forceRefresh,  // 是否强制刷新缓存
		CheckExpires: &checkExpires, // 是否检查过期
	})
	if err != nil {
		return nil, fmt.Errorf("调用云盘GetSubFiles API失败: %w", err)
	}
	var collectedFiles []*clouddrive.CloudDriveFile
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("从SubFiles流接收数据失败: %w", err)
		}
		if reply != nil && reply.SubFiles != nil {
			collectedFiles = append(collectedFiles, reply.SubFiles...)
		}
	}
	return collectedFiles, nil
}

func (c *Client) FindFileByPath(path string) (*clouddrive.CloudDriveFile, error) {
	return c.cd.FindFileByPath(c.contextWithHeader, &clouddrive.FindFileByPathRequest{Path: path})
}

func (c *Client) CreateFolder(parentPath, folderName string) (*clouddrive.CloudDriveFile, error) {
	res, err := c.cd.CreateFolder(c.contextWithHeader, &clouddrive.CreateFolderRequest{ParentPath: parentPath, FolderName: folderName})
	if err != nil {
		return nil, err
	}
	if !res.GetResult().GetSuccess() {
		return nil, errors.New(res.GetResult().GetErrorMessage())
	}
	return res.GetFolderCreated(), nil
}

func (c *Client) RenameFile(theFilePath, newName string) error {
	res, err := c.cd.RenameFile(c.contextWithHeader, &clouddrive.RenameFileRequest{TheFilePath: theFilePath, NewName: newName})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) MoveFile(theFilePaths []string, destPath string) error {
	res, err := c.cd.MoveFile(c.contextWithHeader, &clouddrive.MoveFileRequest{TheFilePaths: theFilePaths, DestPath: destPath})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) DeleteFile(path string) error {
	res, err := c.cd.DeleteFile(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) DeleteFiles(paths []string) error {
	res, err := c.cd.DeleteFiles(c.contextWithHeader, &clouddrive.MultiFileRequest{Path: paths})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) RemoveOfflineFiles(infoHashes []string, cloudAccountId string) error {
	res, err := c.cd.RemoveOfflineFiles(c.contextWithHeader, &clouddrive.RemoveOfflineFilesRequest{InfoHashes: infoHashes, CloudAccountId: cloudAccountId})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) ListOfflineFilesByPath(path string) (*clouddrive.OfflineFileListResult, error) {
	return c.cd.ListOfflineFilesByPath(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
}

func (c *Client) GetSpaceInfo(path string) (*clouddrive.SpaceInfo, error) {
	return c.cd.GetSpaceInfo(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
}

func (c *Client) GetAllCloudApis() (*clouddrive.CloudAPIList, error) {
	return c.cd.GetAllCloudApis(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetMountPoints() (*clouddrive.GetMountPointsResult, error) {
	return c.cd.GetMountPoints(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetSystemInfo() (*clouddrive.CloudDriveSystemInfo, error) {
	// This is a public method, no token needed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.cd.GetSystemInfo(ctx, &emptypb.Empty{})
}

func (c *Client) Logout() error {
	res, err := c.cd.Logout(c.contextWithHeader, &clouddrive.UserLogoutRequest{})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) GetAccountStatus() (*clouddrive.AccountStatusResult, error) {
	return c.cd.GetAccountStatus(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetRuntimeInfo() (*clouddrive.RuntimeInfo, error) {
	return c.cd.GetRuntimeInfo(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetRunningInfo() (*clouddrive.RunInfo, error) {
	return c.cd.GetRunningInfo(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) ChangePassword(oldPassword, newPassword string) error {
	res, err := c.cd.ChangePassword(c.contextWithHeader, &clouddrive.ChangePasswordRequest{OldPassword: oldPassword, NewPassword: newPassword})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) RenameFiles(renameFiles []*clouddrive.RenameFileRequest) error {
	res, err := c.cd.RenameFiles(c.contextWithHeader, &clouddrive.RenameFilesRequest{RenameFiles: renameFiles})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) DeleteFilePermanently(path string) error {
	res, err := c.cd.DeleteFilePermanently(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) DeleteFilesPermanently(paths []string) error {
	res, err := c.cd.DeleteFilesPermanently(c.contextWithHeader, &clouddrive.MultiFileRequest{Path: paths})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) GetMetaData(path string) (*clouddrive.FileMetaData, error) {
	return c.cd.GetMetaData(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
}

func (c *Client) GetOriginalPath(path string) (string, error) {
	res, err := c.cd.GetOriginalPath(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
	if err != nil {
		return "", err
	}
	return res.GetResult(), nil
}

func (c *Client) GetAllTasksCount() (*clouddrive.GetAllTasksCountResult, error) {
	return c.cd.GetAllTasksCount(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetDownloadFileList() (*clouddrive.GetDownloadFileListResult, error) {
	return c.cd.GetDownloadFileList(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetUploadFileList(req *clouddrive.GetUploadFileListRequest) (*clouddrive.GetUploadFileListResult, error) {
	return c.cd.GetUploadFileList(c.contextWithHeader, req)
}

func (c *Client) CancelUploadFiles(keys []string) error {
	_, err := c.cd.CancelUploadFiles(c.contextWithHeader, &clouddrive.MultpleUploadFileKeyRequest{Keys: keys})
	return err
}

func (c *Client) PauseUploadFiles(keys []string) error {
	_, err := c.cd.PauseUploadFiles(c.contextWithHeader, &clouddrive.MultpleUploadFileKeyRequest{Keys: keys})
	return err
}

func (c *Client) ResumeUploadFiles(keys []string) error {
	_, err := c.cd.ResumeUploadFiles(c.contextWithHeader, &clouddrive.MultpleUploadFileKeyRequest{Keys: keys})
	return err
}

func (c *Client) GetSystemSettings() (*clouddrive.SystemSettings, error) {
	return c.cd.GetSystemSettings(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) SetSystemSettings(settings *clouddrive.SystemSettings) error {
	_, err := c.cd.SetSystemSettings(c.contextWithHeader, settings)
	return err
}

func (c *Client) RestartService() error {
	_, err := c.cd.RestartService(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) ShutdownService() error {
	_, err := c.cd.ShutdownService(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) APILoginAliyundriveOAuth(req *clouddrive.LoginAliyundriveOAuthRequest) error {
	res, err := c.cd.APILoginAliyundriveOAuth(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) APILoginAliyundriveRefreshtoken(req *clouddrive.LoginAliyundriveRefreshtokenRequest) error {
	res, err := c.cd.APILoginAliyundriveRefreshtoken(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) APILoginAliyunDriveQRCode(useOpenAPI bool) (clouddrive.CloudDriveFileSrv_APILoginAliyunDriveQRCodeClient, error) {
	return c.cd.APILoginAliyunDriveQRCode(c.contextWithHeader, &clouddrive.LoginAliyundriveQRCodeRequest{UseOpenAPI: useOpenAPI})
}

func (c *Client) APILoginBaiduPanOAuth(req *clouddrive.LoginBaiduPanOAuthRequest) error {
	res, err := c.cd.APILoginBaiduPanOAuth(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) APILoginOneDriveOAuth(req *clouddrive.LoginOneDriveOAuthRequest) error {
	res, err := c.cd.APILoginOneDriveOAuth(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) ApiLoginGoogleDriveOAuth(req *clouddrive.LoginGoogleDriveOAuthRequest) error {
	res, err := c.cd.ApiLoginGoogleDriveOAuth(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) ApiLoginGoogleDriveRefreshToken(req *clouddrive.LoginGoogleDriveRefreshTokenRequest) error {
	res, err := c.cd.ApiLoginGoogleDriveRefreshToken(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) ApiLoginXunleiOAuth(req *clouddrive.LoginXunleiOAuthRequest) error {
	res, err := c.cd.ApiLoginXunleiOAuth(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) APILogin189QRCode() (clouddrive.CloudDriveFileSrv_APILogin189QRCodeClient, error) {
	return c.cd.APILogin189QRCode(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) APILoginPikPak(username, password string) error {
	res, err := c.cd.APILoginPikPak(c.contextWithHeader, &clouddrive.UserLoginRequest{UserName: username, Password: password})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) APILoginWebDav(req *clouddrive.LoginWebDavRequest) error {
	res, err := c.cd.APILoginWebDav(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) APIAddLocalFolder(path string) error {
	res, err := c.cd.APIAddLocalFolder(c.contextWithHeader, &clouddrive.AddLocalFolderRequest{LocalFolderPath: path})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) RemoveCloudAPI(cloudName, userName string, permanentRemove bool) error {
	res, err := c.cd.RemoveCloudAPI(c.contextWithHeader, &clouddrive.RemoveCloudAPIRequest{CloudName: cloudName, UserName: userName, PermanentRemove: permanentRemove})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) GetCloudAPIConfig(cloudName, userName string) (*clouddrive.CloudAPIConfig, error) {
	return c.cd.GetCloudAPIConfig(c.contextWithHeader, &clouddrive.GetCloudAPIConfigRequest{CloudName: cloudName, UserName: userName})
}

func (c *Client) SetCloudAPIConfig(cloudName, userName string, config *clouddrive.CloudAPIConfig) error {
	_, err := c.cd.SetCloudAPIConfig(c.contextWithHeader, &clouddrive.SetCloudAPIConfigRequest{CloudName: cloudName, UserName: userName, Config: config})
	return err
}

// Mount Point Management
func (c *Client) CanAddMoreMountPoints() (bool, error) {
	res, err := c.cd.CanAddMoreMountPoints(c.contextWithHeader, &emptypb.Empty{})
	if err != nil {
		return false, err
	}
	if !res.Success {
		return false, errors.New(res.ErrorMessage)
	}
	return true, nil
}

func (c *Client) AddMountPoint(req *clouddrive.MountOption) error {
	res, err := c.cd.AddMountPoint(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.FailReason)
	}
	return nil
}

func (c *Client) RemoveMountPoint(mountPoint string) error {
	res, err := c.cd.RemoveMountPoint(c.contextWithHeader, &clouddrive.MountPointRequest{MountPoint: mountPoint})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.FailReason)
	}
	return nil
}

func (c *Client) Mount(mountPoint string) error {
	res, err := c.cd.Mount(c.contextWithHeader, &clouddrive.MountPointRequest{MountPoint: mountPoint})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.FailReason)
	}
	return nil
}

func (c *Client) Unmount(mountPoint string) error {
	res, err := c.cd.Unmount(c.contextWithHeader, &clouddrive.MountPointRequest{MountPoint: mountPoint})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.FailReason)
	}
	return nil
}

func (c *Client) UpdateMountPoint(req *clouddrive.UpdateMountPointRequest) error {
	res, err := c.cd.UpdateMountPoint(c.contextWithHeader, req)
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.FailReason)
	}
	return nil
}

func (c *Client) GetAvailableDriveLetters() (*clouddrive.GetAvailableDriveLettersResult, error) {
	return c.cd.GetAvailableDriveLetters(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) HasDriveLetters() (*clouddrive.HasDriveLettersResult, error) {
	return c.cd.HasDriveLetters(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) LocalGetSubFiles(req *clouddrive.LocalGetSubFilesRequest) (clouddrive.CloudDriveFileSrv_LocalGetSubFilesClient, error) {
	return c.cd.LocalGetSubFiles(c.contextWithHeader, req)
}

// Backup & Restore
func (c *Client) BackupGetAll() (*clouddrive.BackupList, error) {
	return c.cd.BackupGetAll(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) BackupAdd(backup *clouddrive.Backup) error {
	_, err := c.cd.BackupAdd(c.contextWithHeader, backup)
	return err
}

func (c *Client) BackupRemove(sourcePath string) error {
	_, err := c.cd.BackupRemove(c.contextWithHeader, &clouddrive.StringValue{Value: sourcePath})
	return err
}

func (c *Client) BackupUpdate(backup *clouddrive.Backup) error {
	_, err := c.cd.BackupUpdate(c.contextWithHeader, backup)
	return err
}

func (c *Client) BackupAddDestination(req *clouddrive.BackupModifyRequest) error {
	_, err := c.cd.BackupAddDestination(c.contextWithHeader, req)
	return err
}

func (c *Client) BackupRemoveDestination(req *clouddrive.BackupModifyRequest) error {
	_, err := c.cd.BackupRemoveDestination(c.contextWithHeader, req)
	return err
}

func (c *Client) BackupSetEnabled(req *clouddrive.BackupSetEnabledRequest) error {
	_, err := c.cd.BackupSetEnabled(c.contextWithHeader, req)
	return err
}

func (c *Client) BackupSetFileSystemWatchEnabled(req *clouddrive.BackupModifyRequest) error {
	_, err := c.cd.BackupSetFileSystemWatchEnabled(c.contextWithHeader, req)
	return err
}

func (c *Client) BackupRestartWalkingThrough(sourcePath string) error {
	_, err := c.cd.BackupRestartWalkingThrough(c.contextWithHeader, &clouddrive.StringValue{Value: sourcePath})
	return err
}

func (c *Client) CanAddMoreBackups() (bool, error) {
	res, err := c.cd.CanAddMoreBackups(c.contextWithHeader, &emptypb.Empty{})
	if err != nil {
		return false, err
	}
	if !res.Success {
		return false, errors.New(res.ErrorMessage)
	}
	return true, nil
}

// User Management
func (c *Client) Register(username, password string) error {
	// This is a public method, no token needed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := c.cd.Register(ctx, &clouddrive.UserRegisterRequest{UserName: username, Password: password})
	if err != nil {
		return err
	}
	if !res.Success {
		return errors.New(res.ErrorMessage)
	}
	return nil
}

func (c *Client) SendConfirmEmail() error {
	_, err := c.cd.SendConfirmEmail(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) ConfirmEmail(code string) error {
	_, err := c.cd.ConfirmEmail(c.contextWithHeader, &clouddrive.ConfirmEmailRequest{ConfirmCode: code})
	return err
}

func (c *Client) SendResetAccountEmail(email string) error {
	// This is a public method, no token needed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := c.cd.SendResetAccountEmail(ctx, &clouddrive.SendResetAccountEmailRequest{Email: email})
	return err
}

func (c *Client) ResetAccount(resetCode, newPassword string) error {
	// This is a public method, no token needed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := c.cd.ResetAccount(ctx, &clouddrive.ResetAccountRequest{ResetCode: resetCode, NewPassword: newPassword})
	return err
}

// Promotions, Plans, and Balance
func (c *Client) GetPromotions() (*clouddrive.GetPromotionsResult, error) {
	return c.cd.GetPromotions(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) UpdatePromotionResult() error {
	_, err := c.cd.UpdatePromotionResult(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) GetCloudDrivePlans() (*clouddrive.GetCloudDrivePlansResult, error) {
	return c.cd.GetCloudDrivePlans(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) JoinPlan(req *clouddrive.JoinPlanRequest) (*clouddrive.JoinPlanResult, error) {
	res, err := c.cd.JoinPlan(c.contextWithHeader, req)
	if err != nil {
		return nil, err
	}
	if !res.Success {
		// This one has a paymentInfo field, so we should probably return it even on failure
		return res, errors.New("failed to join plan")
	}
	return res, nil
}

func (c *Client) BindCloudAccount(req *clouddrive.BindCloudAccountRequest) error {
	_, err := c.cd.BindCloudAccount(c.contextWithHeader, req)
	return err
}

func (c *Client) TransferBalance(req *clouddrive.TransferBalanceRequest) error {
	_, err := c.cd.TransferBalance(c.contextWithHeader, req)
	return err
}

func (c *Client) ChangeEmail(req *clouddrive.ChangeUserNameEmailRequest) error {
	_, err := c.cd.ChangeEmail(c.contextWithHeader, req)
	return err
}

func (c *Client) GetBalanceLog() (*clouddrive.BalanceLogResult, error) {
	return c.cd.GetBalanceLog(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) CheckActivationCode(code string) (*clouddrive.CheckActivationCodeResult, error) {
	return c.cd.CheckActivationCode(c.contextWithHeader, &clouddrive.StringValue{Value: code})
}

func (c *Client) ActivatePlan(code string) (*clouddrive.JoinPlanResult, error) {
	res, err := c.cd.ActivatePlan(c.contextWithHeader, &clouddrive.StringValue{Value: code})
	if err != nil {
		return nil, err
	}
	if !res.Success {
		return nil, errors.New("failed to activate plan")
	}
	return res, nil
}

func (c *Client) CheckCouponCode(req *clouddrive.CheckCouponCodeRequest) (*clouddrive.CouponCodeResult, error) {
	return c.cd.CheckCouponCode(c.contextWithHeader, req)
}

func (c *Client) GetReferralCode() (string, error) {
	res, err := c.cd.GetReferralCode(c.contextWithHeader, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return res.GetValue(), nil
}

// System Updates
func (c *Client) HasUpdate() (*clouddrive.UpdateResult, error) {
	return c.cd.HasUpdate(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) CheckUpdate() (*clouddrive.UpdateResult, error) {
	return c.cd.CheckUpdate(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) DownloadUpdate() error {
	_, err := c.cd.DownloadUpdate(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) UpdateSystem() error {
	_, err := c.cd.UpdateSystem(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) GetSearchResults(req *clouddrive.SearchRequest) (clouddrive.CloudDriveFileSrv_GetSearchResultsClient, error) {
	return c.cd.GetSearchResults(c.contextWithHeader, req)
}

func (c *Client) GetFileDetailProperties(path string) (*clouddrive.FileDetailProperties, error) {
	return c.cd.GetFileDetailProperties(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
}

func (c *Client) GetCloudMemberships(path string) (*clouddrive.CloudMemberships, error) {
	return c.cd.GetCloudMemberships(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
}

func (c *Client) GetDownloadFileCount() (*clouddrive.GetDownloadFileCountResult, error) {
	return c.cd.GetDownloadFileCount(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetUploadFileCount() (*clouddrive.GetUploadFileCountResult, error) {
	return c.cd.GetUploadFileCount(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) CancelAllUploadFiles() error {
	_, err := c.cd.CancelAllUploadFiles(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) PauseAllUploadFiles() error {
	_, err := c.cd.PauseAllUploadFiles(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) ResumeAllUploadFiles() error {
	_, err := c.cd.ResumeAllUploadFiles(c.contextWithHeader, &emptypb.Empty{})
	return err
}

func (c *Client) CanAddMoreCloudApis() (bool, error) {
	res, err := c.cd.CanAddMoreCloudApis(c.contextWithHeader, &emptypb.Empty{})
	if err != nil {
		return false, err
	}
	if !res.Success {
		return false, errors.New(res.ErrorMessage)
	}
	return true, nil
}

func (c *Client) SetDirCacheTimeSecs(req *clouddrive.SetDirCacheTimeRequest) error {
	_, err := c.cd.SetDirCacheTimeSecs(c.contextWithHeader, req)
	return err
}

func (c *Client) GetEffectiveDirCacheTimeSecs(path string) (*clouddrive.GetEffectiveDirCacheTimeResult, error) {
	return c.cd.GetEffectiveDirCacheTimeSecs(c.contextWithHeader, &clouddrive.GetEffectiveDirCacheTimeRequest{Path: path})
}

func (c *Client) GetOpenFileTable(includeDir bool) (*clouddrive.OpenFileTable, error) {
	return c.cd.GetOpenFileTable(c.contextWithHeader, &clouddrive.GetOpenFileTableRequest{IncludeDir: includeDir})
}

func (c *Client) GetDirCacheTable() (*clouddrive.DirCacheTable, error) {
	return c.cd.GetDirCacheTable(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetReferencedEntryPaths(path string) (*clouddrive.StringList, error) {
	return c.cd.GetReferencedEntryPaths(c.contextWithHeader, &clouddrive.FileRequest{Path: path})
}

func (c *Client) GetTempFileTable() (*clouddrive.TempFileTable, error) {
	return c.cd.GetTempFileTable(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) PushTaskChange() (clouddrive.CloudDriveFileSrv_PushTaskChangeClient, error) {
	return c.cd.PushTaskChange(c.contextWithHeader, &emptypb.Empty{})
}

func (c *Client) GetCloudDrive1UserData() (string, error) {
	res, err := c.cd.GetCloudDrive1UserData(c.contextWithHeader, &emptypb.Empty{})
	if err != nil {
		return "", err
	}
	return res.GetResult(), nil
}

func (c *Client) WriteToFileStream() (clouddrive.CloudDriveFileSrv_WriteToFileStreamClient, error) {
	return c.cd.WriteToFileStream(c.contextWithHeader)
}

// Deprecated: use BackupUpdate instead
func (c *Client) BackupUpdateStrategies(req *clouddrive.BackupModifyRequest) error {
	_, err := c.cd.BackupUpdateStrategies(c.contextWithHeader, req)
	return err
}
