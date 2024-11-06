package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	BackupsCollection = "backups"
)

// Backup 备份
type Backup struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	BackupID      string             `json:"backup_id" bson:"backup_id"`
	BackupName    string             `json:"backup_name" bson:"backup_name"`
	BackupType    string             `json:"backup_type" bson:"backup_type"`
	CustomerID    string             `json:"customer_id" bson:"customer_id"`
	AppID         string             `json:"app_id" bson:"app_id"`
	AppType       string             `json:"app_type" bson:"app_type"`
	HasData       bool               `json:"has_data" bson:"has_data"`
	Size          int64              `json:"size" bson:"size"`
	CopyInfoList  []*CopyInfo        `json:"copy_info_list" bson:"copy_info_list"`
	FileName      string             `json:"file_name" bson:"file_name"`
	FilePath      string             `json:"file_path" bson:"file_path"`
	CloudFileName string             `json:"cloud_file_name" bson:"cloud_file_name"`
	CloudFilePath string             `json:"cloud_file_path" bson:"cloud_file_path"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy     string             `json:"created_by" bson:"created_by"`
}

// CopyInfo 复制信息
type CopyInfo struct {
	CopyType string `json:"copy_type" bson:"copy_type"`
	Source   string `json:"source" bson:"source"`
	Count    int64  `json:"count" bson:"count"`
}

// ToProto 转换为proto数据
func (b *Backup) ToProto() *backup.Backup {
	var copys []*backup.CopyInfo
	for _, cp := range b.CopyInfoList {
		copys = append(copys, cp.ToProto())
	}

	return &backup.Backup{
		BackupId:      b.BackupID,
		BackupName:    b.BackupName,
		BackupType:    b.BackupType,
		CustomerId:    b.CustomerID,
		AppId:         b.AppID,
		AppType:       b.AppType,
		HasData:       b.HasData,
		Size:          b.Size,
		CopyInfoList:  copys,
		FileName:      b.FileName,
		FilePath:      b.FilePath,
		CloudFileName: b.CloudFileName,
		CloudFilePath: b.CloudFilePath,
		CreatedAt:     b.CreatedAt.String(),
		CreatedBy:     b.CreatedBy,
	}
}

// ToProto 转换为proto数据
func (c *CopyInfo) ToProto() *backup.CopyInfo {
	return &backup.CopyInfo{
		CopyType: c.CopyType,
		Source:   c.Source,
		Count:    c.Count,
	}
}

// FindBackups 默认查询
func FindBackups(ctx context.Context, db, customerID, backupName, BackupType string) (a []Backup, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(BackupsCollection)

	// 默认检索所有数据
	query := bson.M{}

	// 试用顾客ID不为空
	if customerID != "" {
		query["customer_id"] = customerID
	}
	// 备份类型
	if BackupType != "" {
		query["backup_type"] = BackupType
	}

	// 备份名称不为空
	if backupName != "" {
		query["backup_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(backupName), Options: "m"}}
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindBackups", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Backup
	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindBackups", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var bac Backup
		err := cur.Decode(&bac)
		if err != nil {
			utils.ErrorLog("error FindBackups", err.Error())
			return nil, err
		}
		result = append(result, bac)
	}
	return result, nil
}

// FindBackup 通过APPID查找单个APP记录
func FindBackup(ctx context.Context, db, backupID string) (a Backup, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(BackupsCollection)

	var result Backup
	objectID, err := primitive.ObjectIDFromHex(backupID)
	if err != nil {
		utils.ErrorLog("error FindBackup", err.Error())
		return result, err
	}

	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindBackup", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindBackup", err.Error())
		return result, err
	}
	return result, nil
}

// AddBackup 添加单个APP记录
func AddBackup(ctx context.Context, db string, a *Backup) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(BackupsCollection)

	// 编辑ID
	a.ID = primitive.NewObjectID()
	a.BackupID = a.ID.Hex()
	if a.CopyInfoList == nil {
		a.CopyInfoList = make([]*CopyInfo, 0)
	}
	queryJSON, _ := json.Marshal(a)
	utils.DebugLog("AddBackup", fmt.Sprintf("Backup: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, a)
	if err != nil {
		utils.ErrorLog("error AddBackup", err.Error())
		return "", err
	}

	return a.BackupID, nil
}

// HardDeleteBackups 物理删除选中的APP记录
func HardDeleteBackups(ctx context.Context, db string, backupIDList []string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(BackupsCollection)

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteBackups", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteBackups", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, appID := range backupIDList {
			objectID, err := primitive.ObjectIDFromHex(appID)
			if err != nil {
				utils.ErrorLog("error HardDeleteBackups", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteBackups", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(sc, query)
			if err != nil {
				utils.ErrorLog("error HardDeleteBackups", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteBackups", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteBackups", err.Error())
		return err
	}
	session.EndSession(ctx)

	return nil
}
