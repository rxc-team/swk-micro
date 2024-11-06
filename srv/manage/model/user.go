package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/manage/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// UsersCollection user collection
	UsersCollection    = "users"
	RoleCollection     = "roles"
	CustomerCollection = "customers"
)

// User 用户信息
type User struct {
	ID                primitive.ObjectID `json:"id" bson:"_id"`
	UserID            string             `json:"user_id" bson:"user_id"`
	UserName          string             `json:"user_name" bson:"user_name"`
	Email             string             `json:"email" bson:"email"`
	NoticeEmail       string             `json:"notice_email" bson:"notice_email"`
	Password          string             `json:"password" bson:"password"`
	Avatar            string             `json:"avatar" bson:"avatar"`
	CurrentApp        string             `json:"current_app" bson:"current_app"`
	Group             string             `json:"group" bson:"group"`
	Signature         string             `json:"signature" bson:"signature"`
	Language          string             `json:"language" bson:"language"`
	Theme             string             `json:"theme" bson:"theme"`
	Roles             []string           `json:"roles" bson:"roles"`
	Apps              []string           `json:"apps" bson:"apps"`
	Domain            string             `json:"domain" bson:"domain"`
	CustomerID        string             `json:"customer_id" bson:"customer_id"`
	TimeZone          string             `json:"timezone" bson:"timezone"`
	UserType          int32              `json:"user_type" bson:"user_type"`
	ErrorCount        int32              `json:"error_count" bson:"error_count"`
	NoticeEmailStatus string             `json:"notice_email_status" bson:"notice_email_status"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy         string             `json:"created_by" bson:"created_by"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy         string             `json:"updated_by" bson:"updated_by"`
	DeletedAt         time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy         string             `json:"deleted_by" bson:"deleted_by"`
}

// CommonUser 共通用户表信息
type CommonUser struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	UserID     string             `json:"user_id" bson:"user_id"`
	Email      string             `json:"email" bson:"email"`
	UserType   int32              `json:"user_type" bson:"user_type"`
	CustomerID string             `json:"customer_id" bson:"customer_id"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy  string             `json:"created_by" bson:"created_by"`
}

// ToProto 转换为proto数据
func (u *User) ToProto() *user.User {
	return &user.User{
		UserId:            u.UserID,
		UserName:          u.UserName,
		Email:             u.Email,
		NoticeEmail:       u.NoticeEmail,
		NoticeEmailStatus: u.NoticeEmailStatus,
		Avatar:            u.Avatar,
		CurrentApp:        u.CurrentApp,
		Group:             u.Group,
		Signature:         u.Signature,
		Roles:             u.Roles,
		Apps:              u.Apps,
		Language:          u.Language,
		Theme:             u.Theme,
		Domain:            u.Domain,
		CustomerId:        u.CustomerID,
		UserType:          u.UserType,
		ErrorCount:        u.ErrorCount,
		Timezone:          u.TimeZone,
		CreatedAt:         u.CreatedAt.String(),
		CreatedBy:         u.CreatedBy,
		UpdatedAt:         u.UpdatedAt.String(),
		UpdatedBy:         u.UpdatedBy,
		DeletedAt:         u.DeletedAt.String(),
		DeletedBy:         u.DeletedBy,
	}
}

// Login 登录，返回用户信息
func Login(ctx context.Context, email, password string) (u *User, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(UsersCollection)

	query := bson.M{
		"email": email,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("Login", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result CommonUser
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error Login", err.Error())
		return nil, mongo.ErrNoDocuments
	}

	// 从用户表中查询该用户的详细信息返回给接口
	us, e := FindUserByID(ctx, result.CustomerID, result.UserID, false)
	if e != nil {
		return nil, e
	}

	// 判断登录密码输入错误次数
	if us.ErrorCount >= MaxPasswordInputErrorTimes {
		errMsg := "user has been locked"
		return nil, errors.New(errMsg)
	}

	// 判断登录密码輸入正確否
	if us.Password != password {
		// 登录密码输入错误次数累加
		err := errorCountPlus(result.CustomerID, result.UserID)
		if err != nil {
			return nil, err
		}
		// 登录失败-返回登录密码不匹配错误信息
		return nil, errors.New("password is invalid")
	}

	// 登录成功-密码输入错误次数重置
	errReset := unlockUser(result.CustomerID, result.UserID)
	if errReset != nil {
		return nil, errReset
	}

	return &us, nil
}

// FindUserByEmail 通过用户通知邮件查询返回用户信息
func FindUserByEmail(ctx context.Context, email string) (u *User, err error) {
	client := database.New()
	opt := options.Collection()
	opt.SetReadPreference(readpref.PrimaryPreferred())
	c := client.Database(database.Db).Collection(UsersCollection, opt)

	query := bson.M{
		"email": email,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindUserByEmail", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result CommonUser
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindUserByEmail", err.Error())
		return nil, mongo.ErrNoDocuments
	}

	// 从用户表中查询该用户的详细信息返回给接口
	us, e := FindUserByID(ctx, result.CustomerID, result.UserID, false)
	if e != nil {
		return nil, e
	}

	return &us, nil
}

// FindRelatedUsers 查找用户组&关联用户组的多个用户记录
func FindRelatedUsers(ctx context.Context, db, domain, invalidatedIn string, groupIDs []string) (u []User, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)

	var result []User
	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
	}

	// 是否包含无效数据
	if invalidatedIn != "" {
		delete(query, "deleted_by")
	}

	// 组的ID不为空
	if len(groupIDs) > 0 {
		query["group"] = bson.M{"$in": groupIDs}
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindRelatedUsers", fmt.Sprintf("query: [ %s ]", queryJSON))

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindUsers", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			utils.ErrorLog("error FindUsers", err.Error())
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}

// FindUsers 查找多个用户记录
func FindUsers(ctx context.Context, db, userName, email, group, app, role, domain, invalidatedIn, errorCount string) (u []User, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)

	query := bson.M{
		"deleted_by": "",
	}

	// 用户被锁计不为空
	if errorCount != "" {
		query["error_count"] = bson.M{"$gte": MaxPasswordInputErrorTimes}
	}

	// 是否包含无效数据
	if invalidatedIn != "" {
		delete(query, "deleted_by")
	}

	// 用户名不为空
	if userName != "" {
		query["user_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(userName), Options: "im"}}
	}

	// 用户邮箱不为空
	if email != "" {
		query["email"] = email
	}

	// 组的ID不为空
	if group != "" {
		query["group"] = group
	}

	// appID不为空
	if app != "" {
		query["apps"] = bson.M{"$in": []string{app}}
	}

	// 角色ID不为空
	if role != "" {
		query["roles"] = bson.M{"$in": []string{role}}
	}

	queryJSON, err := json.Marshal(query)
	utils.DebugLog("FindUsers", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []User
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindUsers", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			utils.ErrorLog("error FindUsers", err.Error())
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}

// FindUserByID 通过UserID,查找单个用户记录
func FindUserByID(ctx context.Context, db, userID string, showDeleted bool) (u User, err error) {
	client := database.New()
	opt := options.Collection()
	opt.SetReadPreference(readpref.PrimaryPreferred())
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection, opt)

	query := bson.M{
		"user_id": userID,
	}

	if !showDeleted {
		query["deleted_by"] = ""
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindUserByID", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result User
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindUserByID", err.Error())
		return result, err
	}

	return result, nil
}

// FindDefaultUser 通过用户domain&用户FLG,查找默认用户记录
func FindDefaultUser(ctx context.Context, db string, userType int32) (u *User, err error) {
	client := database.New()
	if len(db) > 0 {
		c := client.Database(database.GetDBName(db)).Collection(UsersCollection)

		query := bson.M{
			"user_type": userType,
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("FindDefaultUser", fmt.Sprintf("query: [ %s ]", queryJSON))

		var result User
		if err := c.FindOne(ctx, query).Decode(&result); err != nil {
			utils.ErrorLog("error FindDefaultUser", err.Error())
			return &result, err
		}

		return &result, nil
	}

	c := client.Database(database.Db).Collection(UsersCollection)

	query := bson.M{
		"user_type": userType,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindDefaultUser", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result CommonUser
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindDefaultUser", err.Error())
		return nil, err
	}

	// 从用户表中查询该用户的详细信息返回给接口
	us, e := FindUserByID(ctx, result.CustomerID, result.UserID, true)
	if e != nil {
		return nil, e
	}

	return &us, nil
}

// AddUser 添加单个用户记录
func AddUser(ctx context.Context, db string, u *User) (id string, err error) {
	// 开始处理
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	comm := client.Database(database.Db).Collection(UsersCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error AddUser", err.Error())
		return "", err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error AddUser", err.Error())
		return "", err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		u.ID = primitive.NewObjectID()
		u.UserID = u.ID.Hex()

		queryJSON, _ := json.Marshal(u)
		utils.DebugLog("AddUser", fmt.Sprintf("User: [ %s ]", queryJSON))

		_, err = c.InsertOne(sc, u)
		if err != nil {
			utils.ErrorLog("error AddUser", err.Error())
			return err
		}

		// 插入到共通用户表
		us := CommonUser{
			ID:         u.ID,
			UserID:     u.UserID,
			Email:      u.Email,
			CustomerID: u.CustomerID,
			UserType:   u.UserType,
			CreatedAt:  u.CreatedAt,
			CreatedBy:  u.CreatedBy,
		}

		_, err = comm.InsertOne(sc, us)
		if err != nil {
			utils.ErrorLog("error AddUserCommon", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error AddUser", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error AddUser", err.Error())
		return "", err
	}
	session.EndSession(ctx)

	return u.UserID, nil
}

// errorCountPlus 登录密码输入错误次数累加
func errorCountPlus(db, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorLog("error PasswordInputErrorCountPlus", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$inc": bson.M{"error_count": 1}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("PasswordInputErrorCountPlus", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateJSON, _ := json.Marshal(update)
	utils.DebugLog("PasswordInputErrorCountPlus", fmt.Sprintf("update: [ %s ]", updateJSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error errorCountPlus", err.Error())
		return err
	}

	return nil
}

// unlockUser 恢复被锁用户(登录成功-密码输入错误次数重置)
func unlockUser(db, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorLog("error UnlockUser", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{"error_count": 0}}

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error UnlockUser", err.Error())
		return err
	}

	return nil
}

// ModifyUser 更新用户的信息
func ModifyUser(ctx context.Context, db string, u *User) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	comm := client.Database(database.Db).Collection(UsersCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	objectID, err := primitive.ObjectIDFromHex(u.UserID)
	if err != nil {
		utils.ErrorLog("error ModifyUser", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{}

	// 用户管理更改用户或个人中心更改用户的时候才保存更改时间
	change["updated_at"] = u.UpdatedAt
	change["updated_by"] = u.UpdatedBy

	// 用户名称不为空的场合
	if u.UserName != "" {
		change["user_name"] = u.UserName
	}
	// 用户登录ID不为空的场合
	if u.Email != "" {
		change["email"] = u.Email
	}
	// 用户通知邮箱不为空的场合
	if u.NoticeEmail != "" {
		change["notice_email"] = u.NoticeEmail
	}
	// 用户通知邮箱状态不为空的场合
	if u.NoticeEmailStatus != "" {
		change["notice_email_status"] = u.NoticeEmailStatus
	}
	// 用户密码不为空的场合
	if u.Password != "" {
		change["password"] = u.Password
	}
	// 用户头像不为空的场合
	if u.Avatar != "" {
		change["avatar"] = u.Avatar
	}
	// 用户当前app不为空的场合
	if u.CurrentApp != "" {
		change["current_app"] = u.CurrentApp
	}
	// 用户组不为空的场合
	if u.Group != "" {
		change["group"] = u.Group
	}
	// 用户个性签名不为空的场合
	if u.Signature != "" {
		change["signature"] = u.Signature
	}
	// 用户的当前语言不为空的场合
	if u.Language != "" {
		change["language"] = u.Language
	}
	// 用户的当前时区不为空的场合
	if u.TimeZone != "" {
		change["timezone"] = u.TimeZone
	}
	// 用户的当前主题不为空的场合
	if u.Theme != "" {
		change["theme"] = u.Theme
	}
	// 用户拥有的角色不为空的场合
	if len(u.Roles) > 0 {
		change["roles"] = u.Roles
	}
	// 用户拥有的app不为空的场合
	if len(u.Apps) > 0 {
		change["apps"] = u.Apps
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyUser", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyUser", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyUser", err.Error())
		return err
	}

	changeEmail := bson.M{}

	// 用户管理更改用户或个人中心更改用户的时候才保存更改时间
	changeEmail["updated_at"] = u.UpdatedAt
	changeEmail["updated_by"] = u.UpdatedBy

	if u.Email != "" {
		changeEmail["email"] = u.Email
	}

	updateEmail := bson.M{"$set": changeEmail}

	queryJSONEmail, _ := json.Marshal(query)
	utils.DebugLog("ModifyUser", fmt.Sprintf("query: [ %s ]", queryJSONEmail))

	updateSONEmail, _ := json.Marshal(updateEmail)
	utils.DebugLog("ModifyUser", fmt.Sprintf("update: [ %s ]", updateSONEmail))

	_, err = comm.UpdateOne(ctx, query, updateEmail)
	if err != nil {
		utils.ErrorLog("error ModifyUser", err.Error())
		return err
	}

	return nil
}

// DeleteUser 删除单个用户
func DeleteUser(ctx context.Context, db, userID, writer string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorLog("error DeleteUser", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": writer,
	}}

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteUser", err.Error())
		return err
	}

	return nil
}

// DeleteSelectUsers 删除选中用户
func DeleteSelectUsers(ctx context.Context, db string, userIDlist []string, writer string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectUsers", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectUsers", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, userID := range userIDlist {
			objectID, err := primitive.ObjectIDFromHex(userID)
			if err != nil {
				utils.ErrorLog("error DeleteSelectUsers", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": writer,
			}}

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error DeleteSelectUsers", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectUsers", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectUsers", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// RecoverSelectUsers 恢复选中用户
func RecoverSelectUsers(ctx context.Context, db string, userIDlist []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	comm := client.Database(database.Db).Collection(UsersCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectUsers", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectUsers", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, id := range userIDlist {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				utils.ErrorLog("error RecoverSelectUsers", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"updated_at": time.Now(),
				"updated_by": userID,
				"deleted_by": "",
			}}

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectUsers", err.Error())
				return err
			}

			_, err = comm.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectUsersCommon", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectUsers", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectUsers", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// UnlockSelectUsers 恢复被锁用户
func UnlockSelectUsers(ctx context.Context, db string, userIDlist []string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	comm := client.Database(database.Db).Collection(UsersCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error UnlockSelectUsers", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error UnlockSelectUsers", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, id := range userIDlist {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				utils.ErrorLog("error UnlockSelectUsers", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"updated_at":  time.Now(),
				"updated_by":  userID,
				"error_count": 0,
			}}

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error UnlockSelectUsers", err.Error())
				return err
			}

			_, err = comm.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error UnlockSelectUsersCommon", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error UnlockSelectUsers", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error UnlockSelectUsers", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// AddUserCollectionIndex 去掉重复数据，并所有用户集合的唯一索引
func AddUserCollectionIndex(ctx context.Context, db string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	rolec := client.Database(database.GetDBName(db)).Collection(RoleCollection)
	comm := client.Database(database.Db).Collection(UsersCollection)

	pituniqueMap := make(map[string]int64)
	uniqueMap := make(map[string]int64)

	// 查询pit_system所有用户
	cur, err := c.Find(ctx, bson.M{})
	if err != nil {
		utils.ErrorLog("error AddUserCollectionIndex", err.Error())
		return err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			utils.ErrorLog("error AddUserCollectionIndex", err.Error())
			return err
		}
		if u.UserName == "SYSTEM" {
			role, err := FindRole(ctx, db, u.Roles[0])
			if err != nil {
				utils.ErrorLog("error AddUserCollectionIndex", err.Error())
				return err
			}

			// 删除重复的系统管理员角色
			roleQuery := bson.M{
				"$and": []bson.M{
					{"role_name": "SYSTEM"},
					{"role_id": bson.M{"$ne": role.RoleID}},
				},
			}
			_, err = rolec.DeleteMany(ctx, roleQuery)
			if err != nil {
				utils.ErrorLog("error AddUserCollectionIndex", err.Error())
				return err
			}

		}

		// 查询两表的相同用户，并保存
		query := bson.M{
			"user_id": u.UserID,
			"email":   u.Email,
		}
		var user User
		err = comm.FindOne(ctx, query).Decode(&user)
		if err != nil {
			if err.Error() == mongo.ErrNoDocuments.Error() {
				_, err = c.DeleteOne(ctx, query)
				if err != nil {
					utils.ErrorLog("error AddUserCollectionIndex", err.Error())
					return err
				}
			} else {
				utils.ErrorLog("error AddUserCollectionIndex", err.Error())
				return err
			}
		}
		// 记录pit表的唯一值
		_, ok := pituniqueMap[user.Email]
		if ok {
			// 两表都出现重复的用户，删除
			deleteQuery := bson.M{
				"user_id": u.UserID,
			}
			queryJSON, _ := json.Marshal(deleteQuery)
			utils.DebugLog("AddUserCollectionIndex", fmt.Sprintf("query: [ %s ]", queryJSON))
			_, err = c.DeleteOne(ctx, deleteQuery)
			if err != nil {
				utils.ErrorLog("error AddUserCollectionIndex", err.Error())
				return err
			}
			_, err = comm.DeleteOne(ctx, deleteQuery)
			if err != nil {
				utils.ErrorLog("error AddUserCollectionIndex", err.Error())
				return err
			}
		} else {
			// 记录两表重复的email
			pituniqueMap[user.Email] = 1
			uniqueMap[user.Email] = 1
		}

		// 查询所有相同email的用户
		querypit := bson.M{
			"email": user.Email,
		}
		curpit, err := comm.Find(ctx, querypit)
		if err != nil {
			utils.ErrorLog("error AddUserCollectionIndex", err.Error())
			return err
		}
		defer curpit.Close(ctx)
		for curpit.Next(ctx) {
			var u User
			err := curpit.Decode(&u)
			if err != nil {
				utils.ErrorLog("error AddUserCollectionIndex", err.Error())
				return err
			}
			// 过滤掉匹配的用户
			if u.UserID == user.UserID {
				continue
			}
			_, ok := pituniqueMap[u.Email]
			if ok {
				deleteQuery := bson.M{
					"user_id": u.UserID,
				}
				queryJSON, _ := json.Marshal(deleteQuery)
				utils.DebugLog("AddUserCollectionIndex", fmt.Sprintf("query: [ %s ]", queryJSON))
				_, err = comm.DeleteOne(ctx, deleteQuery)
				if err != nil {
					utils.ErrorLog("error AddUserCollectionIndex", err.Error())
					return err
				}
			} else {
				pituniqueMap[u.Email] = 1
			}
		}
	}

	// 查询pit中所有非system的用户数据
	customerQuery := bson.M{
		"customer_id": bson.M{
			"$ne": "system",
		},
	}
	pitUserCus, err := comm.Find(ctx, customerQuery)
	if err != nil {
		utils.ErrorLog("error AddUserCollectionIndex", err.Error())
		return err
	}
	defer pitUserCus.Close(ctx)
	for pitUserCus.Next(ctx) {
		var pitUser User
		err = pitUserCus.Decode(&pitUser)
		if err != nil {
			utils.ErrorLog("error AddUserCollectionIndex", err.Error())
			return err
		}
		query := bson.M{
			"user_id": pitUser.UserID,
			"email":   pitUser.Email,
		}
		var user User
		// 查询顾客的user表是否有相同用户
		ccustomer := client.Database(database.GetDBName(pitUser.CustomerID)).Collection(UsersCollection)
		err = ccustomer.FindOne(ctx, query).Decode(&user)
		if err != nil {
			if err.Error() == mongo.ErrNoDocuments.Error() {
				_, err = comm.DeleteOne(ctx, query)
				if err != nil {
					utils.ErrorLog("error AddUserCollectionIndex", err.Error())
					return err
				}
			} else {
				utils.ErrorLog("error AddUserCollectionIndex", err.Error())
				return err
			}
		}
	}
	// email唯一索引
	indexEmail := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// user_name唯一索引
	indexName := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	// role_name唯一索引
	roleName := mongo.IndexModel{
		Keys:    bson.D{{Key: "role_name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// 判断pit_system中是否存在 role_name 索引
	systemRoleName, err := IndexExits(ctx, rolec, "role_name")
	if err != nil {
		utils.ErrorLog("AddUserCollectionIndex", err.Error())
		return err
	}
	if !systemRoleName {
		// 添加pit表 email 唯一索引
		if _, err := rolec.Indexes().CreateOne(ctx, roleName); err != nil {
			utils.ErrorLog("AddUserCollectionIndex", err.Error())
			return err
		}
	}
	// 判断pit中是否存在 email 索引
	pitEmail, err := IndexExits(ctx, comm, "email")
	if err != nil {
		utils.ErrorLog("AddUserCollectionIndex", err.Error())
		return err
	}
	if !pitEmail {
		// 添加pit表 email 唯一索引
		if _, err := comm.Indexes().CreateOne(ctx, indexEmail); err != nil {
			utils.ErrorLog("AddUserCollectionIndex", err.Error())
			return err
		}
	}
	// 判断pit_system中是否存在 email 和 user_name 索引
	systemEmail, err := IndexExits(ctx, c, "email")
	if err != nil {
		utils.ErrorLog("AddUserCollectionIndex", err.Error())
		return err
	}
	systemUserName, err := IndexExits(ctx, c, "user_name")
	if err != nil {
		utils.ErrorLog("AddUserCollectionIndex", err.Error())
		return err
	}
	if !systemEmail {
		// 添加pit_system表email唯一索引
		if _, err := c.Indexes().CreateOne(ctx, indexEmail); err != nil {
			utils.ErrorLog("AddUserCollectionIndex", err.Error())
			return err
		}
	}
	if !systemUserName {
		// 添加pit_system表user_name唯一索引
		if _, err := c.Indexes().CreateOne(ctx, indexName); err != nil {
			utils.ErrorLog("AddUserCollectionIndex", err.Error())
			return err
		}
	}

	// 添加每个顾客user表的user_name唯一索引
	ccustom := client.Database(database.Db).Collection(CustomerCollection)
	customerCur, err := ccustom.Find(ctx, bson.M{})
	if err != nil {
		utils.ErrorLog("AddUserCollectionIndex", err.Error())
		return err
	}
	defer customerCur.Close(ctx)
	for customerCur.Next(ctx) {
		var customer Customer
		err = customerCur.Decode(&customer)
		if err != nil {
			utils.ErrorLog("AddUserCollectionIndex", err.Error())
			return err
		}
		cuser := client.Database(database.GetDBName(customer.CustomerID)).Collection(UsersCollection)
		CustomerUserName, err := IndexExits(ctx, c, "user_name")
		if err != nil {
			utils.ErrorLog("AddUserCollectionIndex", err.Error())
			return err
		}
		if !CustomerUserName {
			// 添加顾客user表的user_name唯一索引
			if _, err := cuser.Indexes().CreateOne(ctx, indexName); err != nil {
				utils.ErrorLog("AddUserCollectionIndex", err.Error())
				return err
			}
		}
		CustomerEmail, err := IndexExits(ctx, c, "email")
		if err != nil {
			utils.ErrorLog("AddUserCollectionIndex", err.Error())
			return err
		}
		if !CustomerEmail {
			// 添加顾客user表的email唯一索引
			if _, err := cuser.Indexes().CreateOne(ctx, indexEmail); err != nil {
				utils.ErrorLog("AddUserCollectionIndex", err.Error())
				return err
			}
		}
	}
	return nil
}

// UploadUsers 批量导入用户
func UploadUsers(ctx context.Context, stream user.UserService_UploadStream) error {
	// 开始处理
	client := database.New()
	db := database.Db

	defer stream.Close()

	// 执行任务
	var userModels []mongo.WriteModel
	var commModels []mongo.WriteModel

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		// 当发送结束
		if req.Status == user.SendStatus_COMPLETE {
			break
		}

		id, err := primitive.ObjectIDFromHex(req.GetUserId())
		if err != nil {
			return err
		}

		user := User{
			ID:                id,
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
			Domain:            req.GetDomain(),
			CustomerID:        req.GetCustomerId(),
			UserType:          req.GetUserType(),
			TimeZone:          req.GetTimezone(),
			CreatedAt:         time.Now(),
			CreatedBy:         req.GetWriter(),
			UpdatedAt:         time.Now(),
			UpdatedBy:         req.GetWriter(),
		}

		db = req.GetDatabase()

		om := mongo.NewInsertOneModel()
		om.SetDocument(user)
		userModels = append(userModels, om)

		// 插入到共通用户表
		us := CommonUser{
			ID:         user.ID,
			UserID:     user.UserID,
			Email:      user.Email,
			CustomerID: user.CustomerID,
			UserType:   user.UserType,
			CreatedAt:  user.CreatedAt,
			CreatedBy:  user.CreatedBy,
		}

		cm := mongo.NewInsertOneModel()
		cm.SetDocument(us)

		commModels = append(commModels, cm)
	}

	var inserted int64 = 0
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)
	comm := client.Database(database.Db).Collection(UsersCollection)

	if len(userModels) > 0 {
		res, err := c.BulkWrite(ctx, userModels)
		if err != nil {
			isDuplicate := mongo.IsDuplicateKeyError(err)
			if isDuplicate {
				utils.ErrorLog("UploadUsers", err.Error())
				return errors.New("duplicate user")
			}

			utils.ErrorLog("UploadUsers", err.Error())
			return err
		}

		inserted = res.InsertedCount
	}

	if len(commModels) > 0 {
		_, err := comm.BulkWrite(ctx, commModels)
		if err != nil {
			isDuplicate := mongo.IsDuplicateKeyError(err)
			if isDuplicate {
				utils.ErrorLog("UploadUsers", err.Error())
				return errors.New("duplicate user")
			}

			utils.ErrorLog("UploadUsers", err.Error())
			return err
		}
	}

	err := stream.Send(&user.UploadResponse{
		Status:   user.Status_SUCCESS,
		Inserted: inserted,
	})

	if err != nil {
		return err
	}

	return nil
}

type DownloadParam struct {
	UserName      string
	Email         string
	Group         string
	App           string
	Role          string
	InvalidatedIn string
	HasLock       string
}

// DownloadUsers 批量导出用户
func DownloadUsers(ctx context.Context, db string, p DownloadParam, stream user.UserService_DownloadStream) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(UsersCollection)

	query := bson.M{
		"deleted_by": "",
	}

	// 用户被锁计不为空
	if p.HasLock != "" {
		query["error_count"] = bson.M{"$gte": MaxPasswordInputErrorTimes}
	}

	// 是否包含无效数据
	if p.InvalidatedIn != "" {
		delete(query, "deleted_by")
	}

	// 用户名不为空
	if p.UserName != "" {
		query["user_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(p.UserName), Options: "im"}}
	}

	// 用户邮箱不为空
	if p.Email != "" {
		query["email"] = p.Email
	}

	// 组的ID不为空
	if p.Group != "" {
		query["group"] = p.Group
	}

	// appID不为空
	if p.App != "" {
		query["apps"] = bson.M{"$in": []string{p.App}}
	}

	// 角色ID不为空
	if p.Role != "" {
		query["roles"] = bson.M{"$in": []string{p.Role}}
	}

	queryJSON, err := json.Marshal(query)
	utils.DebugLog("DownloadUsers", fmt.Sprintf("query: [ %s ]", queryJSON))

	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("DownloadUsers", err.Error())
		return err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var u User
		err := cur.Decode(&u)
		if err != nil {
			utils.ErrorLog("DownloadUsers", err.Error())
			return err
		}

		if err := stream.Send(&user.DownloadResponse{User: u.ToProto()}); err != nil {
			utils.ErrorLog("DownloadUsers", err.Error())
			return err
		}
	}

	return nil
}
