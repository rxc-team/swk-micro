package handler

import (
	"context"
	"time"

	"rxcsoft.cn/pit3/srv/global/model"
	config "rxcsoft.cn/pit3/srv/global/proto/mail-config"
	"rxcsoft.cn/pit3/srv/global/utils"
)

// Config 邮件配置
type Config struct{}

// log出力使用
const (
	ActionFindConfigs  = "FindConfigs"
	ActionFindConfig   = "FindConfig"
	ActionAddConfig    = "AddConfig"
	ActionModifyConfig = "ModifyConfig"
)

// FindConfigs 获取邮件配置集合
func (mc *Config) FindConfigs(ctx context.Context, req *config.FindConfigsRequest, rsp *config.FindConfigsResponse) error {
	utils.InfoLog(ActionFindConfigs, utils.MsgProcessStarted)

	configs, err := model.FindConfigs(req.GetDatabase())
	if err != nil {
		utils.ErrorLog(ActionFindConfigs, err.Error())
		return err
	}

	res := &config.FindConfigsResponse{}
	for _, config := range configs {
		res.Configs = append(res.Configs, config.ToProto())
	}

	*rsp = *res

	utils.InfoLog(ActionFindConfigs, utils.MsgProcessEnded)
	return nil
}

// FindConfig 获取邮件配置
func (mc *Config) FindConfig(ctx context.Context, req *config.FindConfigRequest, rsp *config.FindConfigResponse) error {
	utils.InfoLog(ActionFindConfig, utils.MsgProcessStarted)

	config, err := model.FindConfig(req.GetDatabase())
	if err != nil {
		utils.ErrorLog(ActionFindConfig, err.Error())
		return err
	}

	rsp.Config = config.ToProto()

	utils.InfoLog(ActionFindConfig, utils.MsgProcessEnded)
	return nil
}

// AddConfig 添加邮件配置
func (mc *Config) AddConfig(ctx context.Context, req *config.AddConfigRequest, rsp *config.AddConfigResponse) error {
	utils.InfoLog(ActionAddConfig, utils.MsgProcessStarted)

	params := model.Config{
		Mail:      req.GetMail(),
		Password:  req.GetPassword(),
		Host:      req.GetHost(),
		Port:      req.GetPort(),
		Ssl:       req.GetSsl(),
		CreatedAt: time.Now(),
		CreatedBy: req.GetWriter(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	id, err := model.AddConfig(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionAddConfig, err.Error())
		return err
	}

	rsp.ConfigId = id

	utils.InfoLog(ActionAddConfig, utils.MsgProcessEnded)
	return nil
}

// ModifyConfig 更新邮件配置
func (mc *Config) ModifyConfig(ctx context.Context, req *config.ModifyConfigRequest, rsp *config.ModifyConfigResponse) error {
	utils.InfoLog(ActionModifyConfig, utils.MsgProcessStarted)

	params := model.Config{
		ConfigID:  req.GetConfigId(),
		Mail:      req.GetMail(),
		Password:  req.GetPassword(),
		Host:      req.GetHost(),
		Port:      req.GetPort(),
		Ssl:       req.GetSsl(),
		UpdatedAt: time.Now(),
		UpdatedBy: req.GetWriter(),
	}

	err := model.ModifyConfig(req.GetDatabase(), &params)
	if err != nil {
		utils.ErrorLog(ActionModifyConfig, err.Error())
		return err
	}

	utils.InfoLog(ActionModifyConfig, utils.MsgProcessEnded)
	return nil
}
