package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// User 用户信息
type User struct{}

// log出力使用
const (
	UserProcessName              = "User"
	ActionLogin                  = "Login"
	ActionFindUserByEmail        = "FindUserByEmail"
	ActionFindRelatedUsers       = "FindRelatedUsers"
	ActionFindUsers              = "FindUsers"
	ActionFindUser               = "FindUser"
	ActionFindDefaultUser        = "FindDefaultUser"
	ActionAddUser                = "AddUser"
	ActionModifyUser             = "ModifyUser"
	ActionDeleteUser             = "DeleteUser"
	ActionDeleteSelectUsers      = "DeleteSelectUsers"
	ActionHardDeleteUsers        = "HardDeleteUsers"
	ActionRecoverSelectUsers     = "RecoverSelectUsers"
	ActionUnlockSelectUsers      = "UnlockSelectUsers"
	ActionAddUserCollectionIndex = "AddUserCollectionIndex"
	ActionUpload                 = "Upload"
	ActionDownload               = "Download"
)

// Login 登录，返回用户信息
func (u *User) Login(ctx context.Context, req *user.LoginRequest, rsp *user.LoginResponse) error {
	utils.InfoLog(ActionLogin, utils.MsgProcessStarted)

	res, err := model.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		rsp.Error = err.Error()
		return nil
	}

	rsp.User = res.ToProto()

	utils.InfoLog(ActionLogin, utils.MsgProcessEnded)
	return nil
}

// FindUserByEmail 通过用户通知邮件查询返回用户信息
func (u *User) FindUserByEmail(ctx context.Context, req *user.EmailRequest, rsp *user.EmailResponse) error {
	utils.InfoLog(ActionFindUserByEmail, utils.MsgProcessStarted)

	res, err := model.FindUserByEmail(ctx, req.GetEmail())
	if err != nil {
		utils.ErrorLog(ActionFindUserByEmail, err.Error())
		return err
	}

	rsp.User = res.ToProto()

	utils.InfoLog(ActionFindUserByEmail, utils.MsgProcessEnded)
	return nil
}

// FindRelatedUsers 查找用户组&关联用户组的多个用户记录
func (u *User) FindRelatedUsers(ctx context.Context, req *user.FindRelatedUsersRequest, rsp *user.FindRelatedUsersResponse) error {
	utils.InfoLog(ActionFindRelatedUsers, utils.MsgProcessStarted)

	users, err := model.FindRelatedUsers(ctx, req.GetDatabase(), req.GetDomain(), req.GetInvalidatedIn(), req.GetGroupIDs())
	if err != nil {
		utils.ErrorLog(ActionFindRelatedUsers, err.Error())
		return err
	}

	res := &user.FindRelatedUsersResponse{}
	for _, u := range users {
		res.Users = append(res.Users, u.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindRelatedUsers, utils.MsgProcessEnded)
	return nil
}

// FindUsers 查找多个用户记录
func (u *User) FindUsers(ctx context.Context, req *user.FindUsersRequest, rsp *user.FindUsersResponse) error {
	utils.InfoLog(ActionFindUsers, utils.MsgProcessStarted)

	users, err := model.FindUsers(ctx, req.GetDatabase(), req.GetUserName(), req.GetEmail(), req.GetGroup(), req.GetApp(), req.GetRole(), req.GetDomain(), req.GetInvalidatedIn(), req.GetErrorCount())
	if err != nil {
		utils.ErrorLog(ActionFindUsers, err.Error())
		return err
	}

	res := &user.FindUsersResponse{}
	for _, u := range users {
		res.Users = append(res.Users, u.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindUsers, utils.MsgProcessEnded)
	return nil
}

// FindUser 查找单个用户记录
func (u *User) FindUser(ctx context.Context, req *user.FindUserRequest, rsp *user.FindUserResponse) error {
	utils.InfoLog(ActionFindUser, utils.MsgProcessStarted)

	if req.GetType() == 1 {
		// FindUserByEmail 通过email,查找单个用户记录
		res, err := model.FindUserByEmail(ctx, req.GetEmail())
		if err != nil {
			utils.ErrorLog(ActionFindUser, err.Error())
			return err
		}
		rsp.User = res.ToProto()

	} else {
		// FindUserByID 通过UserID,查找单个用户记录
		res, err := model.FindUserByID(ctx, req.GetDatabase(), req.GetUserId(), true)
		if err != nil {
			utils.ErrorLog(ActionFindUser, err.Error())
			return err
		}
		rsp.User = res.ToProto()

	}

	utils.InfoLog(ActionFindUser, utils.MsgProcessEnded)
	return nil
}

// FindDefaultUser 通过用户domain&用户FLG,查找默认用户记录
func (u *User) FindDefaultUser(ctx context.Context, req *user.FindDefaultUserRequest, rsp *user.FindDefaultUserResponse) error {
	utils.InfoLog(ActionFindDefaultUser, utils.MsgProcessStarted)

	res, err := model.FindDefaultUser(ctx, req.GetDatabase(), req.GetUserType())
	if err != nil {
		utils.ErrorLog(ActionFindDefaultUser, err.Error())
		return err
	}
	rsp.User = res.ToProto()

	utils.InfoLog(ActionFindDefaultUser, utils.MsgProcessEnded)
	return nil
}

// AddUser 添加单个用户记录
func (u *User) AddUser(ctx context.Context, req *user.AddUserRequest, rsp *user.AddUserResponse) error {
	utils.InfoLog(ActionAddUser, utils.MsgProcessStarted)

	params := model.User{
		UserName:          req.GetUserName(),
		Email:             req.GetEmail(),
		NoticeEmail:       req.GetNoticeEmail(),
		NoticeEmailStatus: req.GetNoticeEmailStatus(),
		Password:          req.GetPassword(),
		Avatar:            req.GetAvatar(),
		CurrentApp:        req.GetCurrentApp(),
		Group:             req.GetGroup(),
		Signature:         req.GetSignature(),
		Language:          req.GetLanguage(),
		Theme:             req.GetTheme(),
		Roles:             req.GetRoles(),
		Apps:              req.GetApps(),
		Domain:            req.GetDomain(),
		CustomerID:        req.GetCustomerId(),
		UserType:          req.GetUserType(),
		TimeZone:          req.GetTimezone(),
		CreatedAt:         time.Now(),
		CreatedBy:         req.GetWriter(),
		UpdatedAt:         time.Now(),
		UpdatedBy:         req.GetWriter(),
	}

	id, err := model.AddUser(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddUser, err.Error())
		return err
	}

	rsp.UserId = id

	utils.InfoLog(ActionAddUser, utils.MsgProcessEnded)

	return nil
}

// ModifyUser 更新用户的信息
func (u *User) ModifyUser(ctx context.Context, req *user.ModifyUserRequest, rsp *user.ModifyUserResponse) error {
	utils.InfoLog(ActionModifyUser, utils.MsgProcessStarted)

	params := model.User{
		UserID:            req.GetUserId(),
		UserName:          req.GetUserName(),
		Email:             req.GetEmail(),
		NoticeEmail:       req.GetNoticeEmail(),
		NoticeEmailStatus: req.GetNoticeEmailStatus(),
		Password:          req.GetPassword(),
		Avatar:            req.GetAvatar(),
		CurrentApp:        req.GetCurrentApp(),
		Group:             req.GetGroup(),
		Signature:         req.GetSignature(),
		Language:          req.GetLanguage(),
		Theme:             req.GetTheme(),
		Roles:             req.GetRoles(),
		Apps:              req.GetApps(),
		TimeZone:          req.GetTimezone(),
		UpdatedAt:         time.Now(),
		UpdatedBy:         req.GetWriter(),
	}

	err := model.ModifyUser(ctx, req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyUser, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyUser, utils.MsgProcessEnded)
	return nil
}

// DeleteUser 删除单个用户
func (u *User) DeleteUser(ctx context.Context, req *user.DeleteUserRequest, rsp *user.DeleteUserResponse) error {
	utils.InfoLog(ActionDeleteUser, utils.MsgProcessStarted)

	err := model.DeleteUser(ctx, req.GetDatabase(), req.GetUserId(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteUser, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteUser, utils.MsgProcessEnded)
	return nil
}

// DeleteSelectUsers 删除选中用户
func (u *User) DeleteSelectUsers(ctx context.Context, req *user.DeleteSelectUsersRequest, rsp *user.DeleteSelectUsersResponse) error {
	utils.InfoLog(ActionDeleteSelectUsers, utils.MsgProcessStarted)

	err := model.DeleteSelectUsers(ctx, req.GetDatabase(), req.GetUserIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionDeleteSelectUsers, err.Error())
		return err
	}

	utils.InfoLog(ActionDeleteSelectUsers, utils.MsgProcessEnded)
	return nil
}

// RecoverSelectUsers 恢复选中用户
func (u *User) RecoverSelectUsers(ctx context.Context, req *user.RecoverSelectUsersRequest, rsp *user.RecoverSelectUsersResponse) error {
	utils.InfoLog(ActionRecoverSelectUsers, utils.MsgProcessStarted)

	err := model.RecoverSelectUsers(ctx, req.GetDatabase(), req.GetUserIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionRecoverSelectUsers, err.Error())
		return err
	}

	utils.InfoLog(ActionRecoverSelectUsers, utils.MsgProcessEnded)
	return nil
}

// UnlockSelectUsers 恢复被锁用户
func (u *User) UnlockSelectUsers(ctx context.Context, req *user.UnlockSelectUsersRequest, rsp *user.UnlockSelectUsersResponse) error {
	utils.InfoLog(ActionUnlockSelectUsers, utils.MsgProcessStarted)

	err := model.UnlockSelectUsers(ctx, req.GetDatabase(), req.GetUserIdList(), req.GetWriter())
	if err != nil {
		utils.ErrorLog(ActionUnlockSelectUsers, err.Error())
		return err
	}

	utils.InfoLog(ActionUnlockSelectUsers, utils.MsgProcessEnded)
	return nil
}

// AddUserCollectionIndex 去掉重复数据，并所有用户集合的唯一索引
func (u *User) AddUserCollectionIndex(ctx context.Context, req *user.AddUserIndexRequest, rsp *user.AddUserIndexResponse) error {
	utils.InfoLog(ActionAddUserCollectionIndex, utils.MsgProcessStarted)
	err := model.AddUserCollectionIndex(ctx, req.GetDb())
	if err != nil {
		utils.ErrorLog(ActionAddUserCollectionIndex, err.Error())
		return err
	}
	utils.InfoLog(ActionAddUserCollectionIndex, utils.MsgProcessEnded)
	return nil
}

// Upload 用户上传
func (u *User) Upload(ctx context.Context, stream user.UserService_UploadStream) error {
	utils.InfoLog(ActionUpload, utils.MsgProcessStarted)

	err := model.UploadUsers(ctx, stream)
	if err != nil {
		utils.ErrorLog(ActionUpload, err.Error())
		return err
	}

	utils.InfoLog(ActionUpload, utils.MsgProcessEnded)
	return nil
}

// Download 用户批量导出
func (u *User) Download(ctx context.Context, req *user.DownloadRequest, stream user.UserService_DownloadStream) error {
	utils.InfoLog(ActionDownload, utils.MsgProcessStarted)

	param := model.DownloadParam{
		UserName:      req.GetUserName(),
		Email:         req.GetEmail(),
		Group:         req.GetGroup(),
		App:           req.GetApp(),
		Role:          req.GetRole(),
		InvalidatedIn: req.GetInvalidatedIn(),
		HasLock:       req.GetErrorCount(),
	}

	err := model.DownloadUsers(ctx, req.GetDatabase(), param, stream)
	if err != nil {
		utils.ErrorLog(ActionDownload, err.Error())
		return err
	}

	utils.InfoLog(ActionDownload, utils.MsgProcessEnded)
	return nil
}
