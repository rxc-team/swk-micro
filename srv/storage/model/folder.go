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

	"rxcsoft.cn/pit3/srv/storage/proto/folder"
	"rxcsoft.cn/pit3/srv/storage/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// FoldersCollection folder collection
	FoldersCollection = "folders"
)

// Folder 文件夹信息
type Folder struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	FolderID   string             `json:"folder_id" bson:"folder_id"`
	FolderName string             `json:"folder_name" bson:"folder_name"`
	FolderDir  string             `json:"folder_dir" bson:"folder_dir"`
	Domain     string             `json:"domain" bson:"domain"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy  string             `json:"created_by" bson:"created_by"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy  string             `json:"updated_by" bson:"updated_by"`
	DeletedAt  time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy  string             `json:"deleted_by" bson:"deleted_by"`
}

// ToProto 转换为proto数据
func (f *Folder) ToProto() *folder.Folder {
	return &folder.Folder{
		FolderId:   f.FolderID,
		FolderName: f.FolderName,
		FolderDir:  f.FolderDir,
		Domain:     f.Domain,
		CreatedAt:  f.CreatedAt.String(),
		CreatedBy:  f.CreatedBy,
		UpdatedAt:  f.UpdatedAt.String(),
		UpdatedBy:  f.UpdatedBy,
		DeletedAt:  f.DeletedAt.String(),
		DeletedBy:  f.DeletedBy,
	}
}

// FindFolders 获取所有的文件夹
func FindFolders(db, domain, fileName string) (u []Folder, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
	}

	// 文件夹名不为空的场合，添加到查询条件中
	if fileName != "" {
		query["folder_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(fileName), Options: "m"}}
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindFolders", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Folder
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	folders, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindFolders", err.Error())
		return nil, err
	}
	defer folders.Close(ctx)
	for folders.Next(ctx) {
		var folder Folder
		err := folders.Decode(&folder)
		if err != nil {
			utils.ErrorLog("error FindFolders", err.Error())
			return nil, err
		}
		result = append(result, folder)
	}

	return result, nil
}

// FindFolder 通过ID获取文件夹信息
func FindFolder(db, folderID string) (u Folder, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Folder
	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		utils.ErrorLog("error FindFolder", err.Error())
		return result, err
	}
	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"_id":        objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindFolder", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindFolder", err.Error())
		return result, err
	}

	return result, nil
}

// AddFolder 添加文件夹
func AddFolder(db string, f *Folder) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	f.ID = primitive.NewObjectID()
	f.FolderID = f.ID.Hex()

	queryJSON, _ := json.Marshal(f)
	utils.DebugLog("AddFolder", fmt.Sprintf("AddFolder: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, f)
	if err != nil {
		utils.ErrorLog("error AddFolder", err.Error())
		return "", err
	}

	return f.FolderID, nil
}

// ModifyFolder 修改文件夹
func ModifyFolder(db, folderID, floderName, userID string) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		utils.ErrorLog("error ModifyFolder", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": time.Now(),
		"updated_by": userID,
	}

	// 文件夹名不为空的场合
	if floderName != "" {
		change["folder_name"] = floderName
	}

	update := bson.M{"$set": change}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyFolder", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyFolder", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyFolder", err.Error())
		return err
	}

	return nil
}

// DeleteFolder 删除单个文件夹
func DeleteFolder(db, folderID, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		utils.ErrorLog("error DeleteFolder", err.Error())
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
	utils.DebugLog("DeleteFolder", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteFolder", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteFolder", err.Error())
		return err
	}

	return nil
}

// DeleteSelectFolders 删除多个文件夹
func DeleteSelectFolders(db string, folderIDList []string, userID string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var modifyCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectFolders", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectFolders", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, folderID := range folderIDList {
			objectID, err := primitive.ObjectIDFromHex(folderID)
			if err != nil {
				utils.ErrorLog("error DeleteSelectFolders", err.Error())
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
			utils.DebugLog("DeleteSelectFolders", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectFolders", fmt.Sprintf("update: [ %s ]", updateSON))

			result, err := c.UpdateOne(sc, query, update)
			modifyCount = result.ModifiedCount
			if err != nil {
				utils.ErrorLog("error DeleteSelectFolders", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectFolders", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectFolders", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//result.ModifiedCount是int64，转换为int
	strInt64 := strconv.FormatInt(modifyCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}

// HardDeleteFolders 物理删除多个文件夹
func HardDeleteFolders(db string, folderIDList []string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var deletedCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteFolders", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteFolders", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, folderID := range folderIDList {
			objectID, err := primitive.ObjectIDFromHex(folderID)
			if err != nil {
				utils.ErrorLog("error HardDeleteFolders", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteFolders", fmt.Sprintf("query: [ %s ]", queryJSON))

			result, err := c.DeleteOne(sc, query)
			deletedCount = result.DeletedCount
			if err != nil {
				utils.ErrorLog("error HardDeleteFolders", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteFolders", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteFolders", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//result.DeletedCount是int64，转换为int
	strInt64 := strconv.FormatInt(deletedCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}

// RecoverSelectFolders 恢复选中文件夹
func RecoverSelectFolders(db string, folderIDList []string, userID string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FoldersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var modifiedCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectFolders", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectFolders", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, folderID := range folderIDList {
			objectID, err := primitive.ObjectIDFromHex(folderID)
			if err != nil {
				utils.ErrorLog("error RecoverSelectFolders", err.Error())
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
			utils.DebugLog("RecoverSelectFolders", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectFolders", fmt.Sprintf("update: [ %s ]", updateSON))

			result, err := c.UpdateOne(sc, query, update)
			modifiedCount = result.ModifiedCount
			if err != nil {
				utils.ErrorLog("error RecoverSelectFolders", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverSelectFolders", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectFolders", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//result.ModifiedCount是int64，转换为int
	strInt64 := strconv.FormatInt(modifiedCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}
