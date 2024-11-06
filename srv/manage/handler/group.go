package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/manage/model"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/utils"
)

// Group 角色
type Group struct{}

// log出力使用
const (
	GroupProcessName       = "Group"
	ActionFindGroups       = "FindGroups"
	ActionFindGroup        = "FindGroup"
	ActionFindGroupAccess  = "FindGroupAccess"
	ActionAddGroup         = "AddGroup"
	ActionModifyGroup      = "ModifyGroup"
	ActionModifyGroupSort  = "ModifyGroupSort"
	ActionHardDeleteGroups = "HardDeleteGroups"
)

// FindGroups 查找多个Group记录
func (g *Group) FindGroups(ctx context.Context, req *group.FindGroupsRequest, rsp *group.FindGroupsResponse) error {
	utils.InfoLog(ActionFindGroups, utils.MsgProcessStarted)

	groups, err := model.FindGroups(ctx, req.GetDatabase(), req.GetDomain(), req.GetGroupName())
	if err != nil {
		utils.ErrorLog(ActionFindGroups, err.Error())
		return err
	}

	res := &group.FindGroupsResponse{}
	for _, r := range groups {
		res.Groups = append(res.Groups, r.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindGroups, utils.MsgProcessEnded)
	return nil
}

// FindGroup 查找单个Group记录
func (g *Group) FindGroup(ctx context.Context, req *group.FindGroupRequest, rsp *group.FindGroupResponse) error {
	utils.InfoLog(ActionFindGroup, utils.MsgProcessStarted)

	res, err := model.FindGroup(ctx, req.Database, req.GetGroupId())
	if err != nil {
		utils.ErrorLog(ActionFindGroup, err.Error())
		return err
	}

	rsp.Group = res.ToProto()

	utils.InfoLog(ActionFindGroup, utils.MsgProcessEnded)
	return nil
}

// FindGroupAccess 查找Group的权限记录
func (g *Group) FindGroupAccess(ctx context.Context, req *group.FindGroupAccessRequest, rsp *group.FindGroupAccessResponse) error {
	utils.InfoLog(ActionFindGroupAccess, utils.MsgProcessStarted)

	accessKey, err := model.FindGroupAccess(ctx, req.Database, req.GetGroupId())
	if err != nil {
		utils.ErrorLog(ActionFindGroupAccess, err.Error())
		return err
	}

	rsp.AccessKey = accessKey

	utils.InfoLog(ActionFindGroupAccess, utils.MsgProcessEnded)
	return nil
}

// AddGroup 添加单个Group记录
func (g *Group) AddGroup(ctx context.Context, req *group.AddGroupRequest, rsp *group.AddGroupResponse) error {
	utils.InfoLog(ActionAddGroup, utils.MsgProcessStarted)

	params := model.Group{
		ParentGroupID: req.GetParentGroupId(),
		GroupName:     req.GetGroupName(),
		DisplayOrder:  req.GetDisplayOrder(),
		Domain:        req.GetDomain(),
		CreatedAt:     time.Now(),
		CreatedBy:     req.GetWriter(),
		UpdatedAt:     time.Now(),
		UpdatedBy:     req.GetWriter(),
	}

	id, err := model.AddGroup(ctx, req.Database, &params)
	if err != nil {
		utils.ErrorLog(ActionAddGroup, err.Error())
		return err
	}

	rsp.GroupId = id

	utils.InfoLog(ActionAddGroup, utils.MsgProcessEnded)
	return nil
}

// ModifyGroup 更新组的信息
func (g *Group) ModifyGroup(ctx context.Context, req *group.ModifyGroupRequest, rsp *group.ModifyGroupResponse) error {
	utils.InfoLog(ActionModifyGroup, utils.MsgProcessStarted)

	params := model.Group{
		GroupID:       req.GetGroupId(),
		ParentGroupID: req.GetParentGroupId(),
		UpdatedAt:     time.Now(),
		UpdatedBy:     req.GetWriter(),
	}

	err := model.ModifyGroup(ctx, req.Database, &params)
	if err != nil {
		utils.ErrorLog(ActionModifyGroup, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyGroup, utils.MsgProcessEnded)
	return nil
}

// HardDeleteGroups 物理删除选中Group
func (g *Group) HardDeleteGroups(ctx context.Context, req *group.HardDeleteGroupsRequest, rsp *group.HardDeleteGroupsResponse) error {
	utils.InfoLog(ActionHardDeleteGroups, utils.MsgProcessStarted)

	err := model.HardDeleteGroups(ctx, req.Database, req.GetGroupIdList())
	if err != nil {
		utils.ErrorLog(ActionHardDeleteGroups, err.Error())
		return err
	}

	utils.InfoLog(ActionHardDeleteGroups, utils.MsgProcessEnded)
	return nil
}
