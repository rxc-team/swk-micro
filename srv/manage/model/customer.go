package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	CustomersCollection = "customers"
	UserCollection      = "users"
)

// Customer 顾客
type Customer struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	CustomerID       string             `json:"customer_id" bson:"customer_id"`
	CustomerName     string             `json:"customer_name" bson:"customer_name"`
	SecondCheck      bool               `json:"second_check" bson:"second_check"`
	CustomerLogo     string             `json:"customer_logo" bson:"customer_logo"`
	Domain           string             `json:"domain" bson:"domain"`
	Database         string             `json:"database" bson:"database"`
	DefaultUser      string             `json:"default_user" bson:"default_user"`
	DefaultUserEmail string             `json:"default_user_email" bson:"default_user_email"`
	DefaultTimezone  string             `json:"default_timezone" bson:"default_timezone"`
	DefaultLanguage  string             `json:"default_language" bson:"default_language"`
	MaxUsers         int32              `json:"max_users" bson:"max_users"`
	UsedUsers        int32              `json:"used_users" bson:"used_users"`
	MaxSize          float64            `json:"max_size" bson:"max_size"`
	UsedSize         float64            `json:"used_size" bson:"used_size"`
	MaxDataSize      float64            `json:"max_data_size" bson:"max_data_size"`
	UsedDataSize     float64            `json:"used_data_size" bson:"used_data_size"`
	Level            string             `json:"level" bson:"level"`
	UploadFileSize   int64              `json:"upload_file_size" bson:"upload_file_size"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy        string             `json:"created_by" bson:"created_by"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy        string             `json:"updated_by" bson:"updated_by"`
	DeletedAt        time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy        string             `json:"deleted_by" bson:"deleted_by"`
}

// UpdateParam 更新参数
type UpdateParam struct {
	CustomerID       string
	CustomerName     string
	SecondCheck      string
	CustomerLogo     string
	DefaultUserEmail string
	MaxUsers         int32
	MaxSize          float64
	MaxDataSize      float64
	Level            string
	UploadFileSize   int64
}

// ToProto 转换为proto数据
func (c *Customer) ToProto() *customer.Customer {
	return &customer.Customer{
		CustomerId:       c.CustomerID,
		CustomerName:     c.CustomerName,
		SecondCheck:      c.SecondCheck,
		CustomerLogo:     c.CustomerLogo,
		Domain:           c.Domain,
		Database:         c.Database,
		DefaultUser:      c.DefaultUser,
		DefaultUserEmail: c.DefaultUserEmail,
		DefaultTimezone:  c.DefaultTimezone,
		DefaultLanguage:  c.DefaultLanguage,
		MaxUsers:         c.MaxUsers,
		UsedUsers:        c.UsedUsers,
		MaxSize:          c.MaxSize,
		UsedSize:         c.UsedSize,
		MaxDataSize:      c.MaxDataSize,
		UsedDataSize:     c.UsedDataSize,
		Level:            c.Level,
		UploadFileSize:   c.UploadFileSize,
		CreatedAt:        c.CreatedAt.String(),
		CreatedBy:        c.CreatedBy,
		UpdatedAt:        c.UpdatedAt.String(),
		UpdatedBy:        c.UpdatedBy,
		DeletedAt:        c.DeletedAt.String(),
		DeletedBy:        c.DeletedBy,
	}
}

// FindCustomers 查找多个顾客记录
func FindCustomers(ctx context.Context, customerName, domain, invalidatedIn string) (cus []Customer, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	query := bson.M{
		"deleted_by": "",
	}

	if customerName != "" {
		query["customer_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(customerName), Options: "m"}}
	}
	if domain != "" {
		query["domain"] = domain
	}
	if invalidatedIn != "" {
		delete(query, "deleted_by")
	}

	var result []Customer
	sortItem := bson.D{
		{Key: "created_at", Value: 1},
	}
	opts := options.Find().SetSort(sortItem)

	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindCustomers", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var cus Customer
		err := cur.Decode(&cus)
		if err != nil {
			utils.ErrorLog("error FindCustomers", err.Error())
			return nil, err
		}
		result = append(result, cus)
	}

	return result, nil

}

// FindCustomer 查找单个顾客记录
func FindCustomer(ctx context.Context, customerID string) (cus Customer, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	var result Customer

	objectID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		utils.ErrorLog("error FindCustomer", err.Error())
		return result, err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindCustomer", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindCustomer", err.Error())
		return result, err
	}

	type DbStats struct {
		StorageSize float64 `bson:"storageSize"`
	}

	var db DbStats
	if err := client.Database(database.GetDBName(customerID)).RunCommand(ctx, bson.M{"dbstats": 1}).Decode(&db); err != nil {
		utils.ErrorLog("error FindCustomer", err.Error())
		return result, err
	}

	result.UsedDataSize = db.StorageSize

	return result, nil
}

// FindCustomerByDomain 通过域名查找单个顾客记录
func FindCustomerByDomain(ctx context.Context, domain string) (cus Customer, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	var result Customer

	query := bson.M{
		"domain": domain,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindCustomerByDomain", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindCustomerByDomain", err.Error())
		return result, err
	}

	return result, nil
}

// AddCustomer 添加单个顾客记录
func AddCustomer(ctx context.Context, g *Customer) (id string, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	g.ID = primitive.NewObjectID()
	g.CustomerID = g.ID.Hex()
	g.Database = database.GetDBName(g.CustomerID)

	queryJSON, _ := json.Marshal(g)
	utils.DebugLog("FindDeleteCustomer", fmt.Sprintf("Customer: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, g)
	if err != nil {
		utils.ErrorLog("error AddCustomer", err.Error())
		return "", err
	}

	// 添加app排序的seq
	seq := GetAppDisplayOrderSequenceName(g.CustomerID)
	if err := CreateSequence(ctx, g.CustomerID, seq, 0); err != nil {
		utils.ErrorLog("error AddCustomer", err.Error())
		return "", err
	}
	// 创建新集合
	err = client.Database(database.GetDBName(g.CustomerID)).CreateCollection(ctx, UsersCollection)
	if err != nil {
		utils.ErrorLog("AddCustomer", err.Error())
		return "", err
	}
	// 添加 user_name、email唯一索引
	cUser := client.Database(database.GetDBName(g.CustomerID)).Collection(UserCollection)
	indexName := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := cUser.Indexes().CreateOne(ctx, indexName); err != nil {
		utils.ErrorLog("AddCustomer", err.Error())
		return "", err
	}
	indexEmailName := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	if _, err := cUser.Indexes().CreateOne(ctx, indexEmailName); err != nil {
		utils.ErrorLog("AddCustomer", err.Error())
		return "", err
	}

	return g.CustomerID, nil
}

// ModifyCustomer 修改单个顾客记录
func ModifyCustomer(ctx context.Context, g *UpdateParam, userID string) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	objectID, err := primitive.ObjectIDFromHex(g.CustomerID)
	if err != nil {
		utils.ErrorLog("error ModifyCustomer", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"second_check": g.SecondCheck,
		"updated_at":   time.Now(),
		"updated_by":   userID,
	}

	if len(g.CustomerName) > 0 {
		change["customer_name"] = g.CustomerName
	}
	if len(g.CustomerLogo) > 0 {
		change["customer_logo"] = g.CustomerLogo
	}
	if len(g.DefaultUserEmail) > 0 {
		change["default_user_email"] = g.DefaultUserEmail
	}
	if len(g.SecondCheck) > 0 {
		sk, err := strconv.ParseBool(g.SecondCheck)
		if err == nil {
			change["second_check"] = sk
		}
	}

	if g.MaxUsers > 0 {
		change["max_users"] = g.MaxUsers
	}
	if g.MaxSize > 0 {
		change["max_size"] = g.MaxSize
	}
	if g.MaxDataSize > 0 {
		change["max_data_size"] = g.MaxDataSize
	}
	if len(g.Level) > 0 {
		change["level"] = g.Level
	}
	if g.UploadFileSize > 0 {
		change["upload_file_size"] = g.UploadFileSize
	}

	update := bson.M{
		"$set": change,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyCustomer", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyCustomer", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyCustomer", err.Error())
		return err
	}

	return nil
}

// ModifyUsedUsers 修改顾客的已使用的用户数量
func ModifyUsedUsers(ctx context.Context, customerID string, usedUsers int32) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	objectID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		utils.ErrorLog("error ModifyUsedUsers", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{
		"$inc": bson.M{
			"used_users": usedUsers,
		},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyUsedUsers", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyUsedUsers", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyUsedUsers", err.Error())
		return err
	}

	return nil
}

// ModifyUsedSizeOfCustomer 修改顾客的已使用的存储空间大小
func ModifyUsedSizeOfCustomer(ctx context.Context, domain string, usedSize float64) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	query := bson.M{
		"domain": domain,
	}

	update := bson.M{
		"$inc": bson.M{
			"used_size": usedSize,
		}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyUsedSizeOfCustomer", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyUsedSizeOfCustomer", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyUsedSizeOfCustomer", err.Error())
		return err
	}

	return nil
}

// DeleteCustomer 删除单个顾客记录
func DeleteCustomer(ctx context.Context, customerID, userID string) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	objectID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		utils.ErrorLog("error DeleteCustomer", err.Error())
		return err
	}

	query := bson.M{
		"_id": objectID,
	}

	update := bson.M{"$set": bson.M{
		"deleted_at": time.Now(),
		"deleted_by": userID,
	}}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteCustomer", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteCustomer", fmt.Sprintf("update: [ %s ]", updateSON))
	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteCustomer", err.Error())
		return err
	}

	return nil
}

// DeleteSelectCustomers 删除多个顾客记录
func DeleteSelectCustomers(ctx context.Context, customerIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectCustomers", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectCustomers", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, customerID := range customerIDList {
			// 无效化顾客信息
			objectID, err := primitive.ObjectIDFromHex(customerID)
			if err != nil {
				utils.ErrorLog("error DeleteSelectCustomers", err.Error())
				return err
			}

			query := bson.M{
				"_id": objectID,
			}

			update := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteSelectCustomers", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectCustomers", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error DeleteSelectCustomers", err.Error())
				return err
			}
			// 无效化顾客的所有App
			ac := client.Database(database.GetDBName(customerID)).Collection(AppsCollection)
			appQuery := bson.M{}

			appUpdate := bson.M{"$set": bson.M{
				"deleted_at": time.Now(),
				"deleted_by": userID,
			}}

			appQueryJSON, _ := json.Marshal(appQuery)
			utils.DebugLog("DeleteSelectCustomers", fmt.Sprintf("query: [ %s ]", appQueryJSON))

			appUpdateSON, _ := json.Marshal(appUpdate)
			utils.DebugLog("DeleteSelectCustomers", fmt.Sprintf("update: [ %s ]", appUpdateSON))

			_, err = ac.UpdateMany(ctx, appQuery, appUpdate)
			if err != nil {
				utils.ErrorLog("error DeleteSelectCustomers", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectCustomers", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectCustomers", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}

// HardDeleteCustomers 物理删除选中客户
func HardDeleteCustomers(ctx context.Context, customerIDList []string) error {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)
	uc := client.Database(database.Db).Collection(UsersCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteCustomers", err.Error())
		return err
	}

	opt := options.Transaction()
	opt.ReadConcern = readconcern.Snapshot()

	if err = session.StartTransaction(opt); err != nil {
		utils.ErrorLog("error HardDeleteCustomers", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, customerID := range customerIDList {

			// 删除pit数据库中customers表中的对应顾客信息
			objectID, err := primitive.ObjectIDFromHex(customerID)
			if err != nil {
				utils.ErrorLog("error HardDeleteCustomers", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteCustomers", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteCustomers", err.Error())
				return err
			}
			// 删除pit数据库中users表中的对应用户信息
			delUserQuery := bson.M{
				"customer_id": customerID,
			}

			uQueryJSON, _ := json.Marshal(delUserQuery)
			utils.DebugLog("HardDeleteCustomers", fmt.Sprintf("query: [ %s ]", uQueryJSON))
			_, err = uc.DeleteMany(sc, delUserQuery)
			if err != nil {
				utils.ErrorLog("error HardDeleteCustomers", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			session.AbortTransaction(ctx)
			if err != nil {
				utils.ErrorLog("error HardDeleteCustomers", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteCustomers", err.Error())
		return err
	}
	session.EndSession(ctx)

	// 删除顾客的数据库
	for _, customerID := range customerIDList {
		d := client.Database(database.GetDBName(customerID))
		if err := d.Drop(ctx); err != nil {
			utils.ErrorLog("error HardDeleteCustomers", err.Error())
			return err
		}
	}

	return nil
}

// RecoverSelectCustomers 恢复选中顾客记录
func RecoverSelectCustomers(ctx context.Context, customerIDList []string, userID string) error {
	client := database.New()
	c := client.Database(database.Db).Collection(CustomersCollection)
	// ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	// defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectCustomers", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectCustomers", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		for _, customerID := range customerIDList {
			objectID, err := primitive.ObjectIDFromHex(customerID)
			if err != nil {
				utils.ErrorLog("error RecoverSelectCustomers", err.Error())
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
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("RecoverSelectCustomers", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectCustomers", fmt.Sprintf("update: [ %s ]", updateSON))

			_, err = c.UpdateOne(sc, query, update)
			if err != nil {
				utils.ErrorLog("error RecoverSelectCustomers", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectCustomers", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectCustomers", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}
