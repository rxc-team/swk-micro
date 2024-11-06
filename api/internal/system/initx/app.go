package initx

import (
	"bytes"
	"context"
	"html/template"
	"net/url"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/langx"
	"rxcsoft.cn/pit3/api/internal/common/logic/mailx"
	"rxcsoft.cn/pit3/api/internal/common/originx"
	"rxcsoft.cn/pit3/api/internal/common/storex"
	"rxcsoft.cn/pit3/api/internal/system/aclx"
	"rxcsoft.cn/pit3/srv/global/proto/language"
	"rxcsoft.cn/pit3/srv/journal/proto/journal"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/group"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

func AddDefaultJournals(db, userID, appID, filename string) error {
	var journals []*journal.Journal

	// 读取journals数据
	err := filex.ReadFile(filename, &journals)
	if err != nil {
		loggerx.ErrorLog("addDefaultJournals", err.Error())
		return err
	}

	for _, j := range journals {
		j.AppId = appID
	}

	journalService := journal.NewJournalService("journal", client.DefaultClient)

	var req journal.ImportRequest
	req.Journals = journals
	req.Writer = userID
	req.Database = db

	_, err = journalService.ImportJournal(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("addDefaultJournals", err.Error())
		return err
	}

	return nil
}

// GetUserAccessKeys 获取用户的许可
func GetUserAccessKeys(userID, db string) (accessKeys []string, err error) {

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUserRequest
	req.Type = 0
	req.UserId = userID
	req.Database = db
	userInfo, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		return nil, err
	}

	groupService := group.NewGroupService("manage", client.DefaultClient)

	var groupReq group.FindGroupAccessRequest
	groupReq.GroupId = userInfo.GetUser().GetGroup()
	groupReq.Database = db
	response, err := groupService.FindGroupAccess(context.TODO(), &groupReq)
	if err != nil {
		return nil, err
	}

	return []string{response.GetAccessKey()}, nil
}

// 添加APP对应用户组
func AddDefaultGroup(db, domain, userID string) (gid string, err error) {
	groupService := group.NewGroupService("manage", client.DefaultClient)

	req := group.FindGroupsRequest{
		Domain:   domain,
		Database: db,
	}

	groups, err := groupService.FindGroups(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("addDefaultGroup", err.Error())
		return "", err
	}

	// 默认组已经存在
	if len(groups.GetGroups()) > 0 {
		return groups.GetGroups()[0].GroupId, nil
	}

	appReq := group.AddGroupRequest{
		ParentGroupId: "root",
		GroupName:     "",
		DisplayOrder:  1,
		Domain:        domain,
		Writer:        userID,
		Database:      db,
	}

	gp, err := groupService.AddGroup(context.TODO(), &appReq)
	if err != nil {
		loggerx.ErrorLog("addDefaultGroup", err.Error())
		return "", err
	}

	// 获取默认用户组名称
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	customerReq := customer.FindCustomerByDomainRequest{
		Domain: domain,
	}
	customer, err := customerService.FindCustomerByDomain(context.TODO(), &customerReq)
	if err != nil {
		loggerx.ErrorLog("addDefaultGroup", err.Error())
		return gp.GetGroupId(), err
	}

	// 添加默认用户组对应的语言
	languageService := language.NewLanguageService("global", client.DefaultClient)
	langParams := language.AddCommonDataRequest{
		Domain:   req.GetDomain(),
		LangCd:   customer.Customer.GetDefaultLanguage(),
		Type:     "groups",
		Key:      gp.GetGroupId(),
		Value:    customer.Customer.GetCustomerName(),
		Writer:   userID,
		Database: db,
	}
	_, err = languageService.AddCommonData(context.TODO(), &langParams)
	if err != nil {
		loggerx.ErrorLog("addDefaultGroup", err.Error())
		return gp.GetGroupId(), err
	}

	// 通知刷新多语言数据
	langx.RefreshLanguage(userID, domain)

	return gp.GetGroupId(), nil
}

// 添加APP对应的语言
func AddAppLangItem(db, domain, langCd, appID, appName, userID string) error {
	langService := language.NewLanguageService("global", client.DefaultClient)

	appReq := language.AddLanguageDataRequest{
		Domain:   domain,
		LangCd:   langCd,
		AppId:    appID,
		AppName:  appName,
		Writer:   userID,
		Database: db,
	}

	_, err := langService.AddLanguageData(context.TODO(), &appReq)
	if err != nil {
		loggerx.ErrorLog("addAppLangItem", err.Error())
		return err
	}

	return nil
}

// 管理员用户若尚不存在的场合：添加默认的管理员用户，管理员已经存在的场合：为管理员用户更新添加APP
func AddDefaultAdminUser(db, appID, domain, userID, gid string) (u string, e error) {
	userService := user.NewUserService("manage", client.DefaultClient)

	req := user.FindDefaultUserRequest{
		UserType: 1,
		Database: db,
	}

	adminUser, err := userService.FindDefaultUser(context.TODO(), &req)
	if err != nil {
		er := errors.Parse(err.Error())
		if er.GetDetail() == mongo.ErrNoDocuments.Error() {
			// 添加默认的管理员角色
			roleID, e := addDefaultAdminRole(db, domain, userID)
			if e != nil {
				loggerx.ErrorLog("addDefaultAdminUser", err.Error())
				return "", e
			}

			// 获取默认管理员用户的名称、时区、语言
			customerService := customer.NewCustomerService("manage", client.DefaultClient)
			customerReq := customer.FindCustomerByDomainRequest{
				Domain: domain,
			}
			clientele, er := customerService.FindCustomerByDomain(context.TODO(), &customerReq)
			if er != nil {
				loggerx.ErrorLog("addDefaultAdminUser", err.Error())
				return "", er
			}
			userName := clientele.Customer.GetDefaultUser()
			defaultUserEmail := clientele.Customer.GetDefaultUserEmail()
			email := cryptox.GenerateMailAddress(userName, domain)
			//随机生成密码
			password := cryptox.GenerateRandPassword()
			// 添加默认管理员用户
			admin := user.AddUserRequest{
				UserName:          userName,
				Email:             email,
				NoticeEmail:       defaultUserEmail,
				Password:          cryptox.GenerateMd5Password(password, email),
				Group:             gid,
				Timezone:          clientele.Customer.GetDefaultTimezone(),
				Roles:             []string{roleID},
				Apps:              []string{appID},
				CustomerId:        db,
				Language:          clientele.Customer.GetDefaultLanguage(),
				Domain:            domain,
				UserType:          1,
				Writer:            userID,
				Database:          db,
				NoticeEmailStatus: "Verifying",
			}

			response, err := userService.AddUser(context.TODO(), &admin)
			if err != nil {
				loggerx.ErrorLog("addDefaultAdminUser", err.Error())
				return "", err
			}
			// 更新顾客已用用户数
			customerUpReq := customer.ModifyUsedUsersRequest{
				CustomerId: clientele.Customer.GetCustomerId(),
				UsedUsers:  1,
			}
			_, err = customerService.ModifyUsedUsers(context.TODO(), &customerUpReq)
			if err != nil {
				loggerx.ErrorLog("addDefaultAdminUser", err.Error())
				return "", err
			}

			aclx.SetUserCasbin(response.GetUserId(), []string{roleID}, []string{appID})

			// 使用用户ID,生成临时令牌发送给用户通知邮箱
			token, err := cryptox.Password(response.GetUserId())
			if err != nil {
				loggerx.ErrorLog("addDefaultAdminUser", err.Error())
				return "", err
			}
			// 临时令牌暂存redis
			store := storex.NewRedisStore(86400)
			// 存入用户邮箱
			store.Set(token, email)

			// 发送临时令牌给用户通知邮箱
			if len(defaultUserEmail) > 0 {
				// 发送密码重置邮件
				// 定义收件人
				mailTo := []string{
					defaultUserEmail,
				}
				// 定义抄送人
				mailCcTo := []string{}
				// 邮件主题
				subject := "Please set user login password"
				// 邮件正文
				origin := originx.GetOrigin(true)
				linkUrl := origin + "/password_reset/" + url.QueryEscape(token)
				tpl := template.Must(template.ParseFiles("assets/html/token.html"))
				params := map[string]string{
					"url": linkUrl,
				}

				var out bytes.Buffer
				err := tpl.Execute(&out, params)
				if err != nil {
					loggerx.ErrorLog("addDefaultAdminUser", err.Error())
					return "", err
				}

				er := mailx.SendMail(db, mailTo, mailCcTo, subject, out.String())
				if er != nil {
					loggerx.ErrorLog("addDefaultAdminUser", err.Error())
					return "", err
				}
			}
			return response.GetUserId(), e
		}

		loggerx.ErrorLog("addDefaultAdminUser", err.Error())
		return "", err
	}

	// 为管理员用户添加APP
	update := user.ModifyUserRequest{
		UserId:   adminUser.GetUser().GetUserId(),
		Apps:     append(adminUser.GetUser().GetApps(), appID),
		Writer:   userID,
		Database: db,
	}

	_, e = userService.ModifyUser(context.TODO(), &update)
	if e != nil {
		loggerx.ErrorLog("addDefaultAdminUser", err.Error())
		return "", e
	}

	aclx.SetUserCasbin(adminUser.GetUser().GetUserId(), adminUser.GetUser().GetRoles(), append(adminUser.GetUser().GetApps(), appID))

	return adminUser.GetUser().GetUserId(), nil
}

func addDefaultAdminRole(db, domain, userID string) (roleID string, err error) {
	roleService := role.NewRoleService("manage", client.DefaultClient)

	req := role.AddRoleRequest{
		RoleName:    "SYSTEM",
		Description: "Administrator",
		Permissions: []*role.Permission{},
		IpSegments:  []*role.IPSegment{},
		Domain:      domain,
		RoleType:    1,
		Writer:      userID,
		Database:    db,
	}

	response, err := roleService.AddRole(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("addDefaultAdminRole", err.Error())
		return "", err
	}
	return response.GetRoleId(), nil
}
