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

	"rxcsoft.cn/pit3/srv/storage/proto/file"
	"rxcsoft.cn/pit3/srv/storage/utils"
	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

const (
	// FilesCollection file collection
	FilesCollection = "files"
)

// File 文件信息
type File struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	FileID      string             `json:"file_id" bson:"file_id"`
	FolderID    string             `json:"folder_id" bson:"folder_id"`
	FileName    string             `json:"file_name" bson:"file_name"`
	ObjectName  string             `json:"object_name" bson:"object_name"`
	FilePath    string             `json:"file_path" bson:"file_path"`
	FileSize    int64              `json:"file_size" bson:"file_size"`
	ContentType string             `json:"content_type" bson:"content_type"`
	Domain      string             `json:"domain" bson:"domain"`
	Owners      []string           `json:"owners" bson:"owners"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy   string             `json:"created_by" bson:"created_by"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy   string             `json:"updated_by" bson:"updated_by"`
	DeletedAt   time.Time          `json:"deleted_at" bson:"deleted_at"`
	DeletedBy   string             `json:"deleted_by" bson:"deleted_by"`
}

const (
	// TimeFormat 日期格式化format
	TimeFormat = "2006-01-02 03:04:05"
)

// ToProto 转换为proto数据
func (f *File) ToProto() *file.File {
	return &file.File{
		FileId:      f.FileID,
		FolderId:    f.FolderID,
		ObjectName:  f.ObjectName,
		FileName:    f.FileName,
		FilePath:    f.FilePath,
		FileSize:    f.FileSize,
		ContentType: f.ContentType,
		Owners:      f.Owners,
		Domain:      f.Domain,
		CreatedAt:   f.CreatedAt.String(),
		CreatedBy:   f.CreatedBy,
		UpdatedAt:   f.UpdatedAt.String(),
		UpdatedBy:   f.UpdatedBy,
		DeletedAt:   f.DeletedAt.String(),
		DeletedBy:   f.DeletedBy,
	}
}

// FindPublicFiles 获取所有公开的文件
func FindPublicFiles(db, fileName, contentType string) (f []File, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"folder_id":  "public",
	}

	// 文件名不为空的场合，添加到查询条件中
	if fileName != "" {
		query["file_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(fileName), Options: "m"}}
	}

	// 文件类型不为空的场合，添加到查询条件中
	if contentType != "" {
		query["content_type"] = contentType
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindPublicFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []File
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	files, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindPublicFiles", err.Error())
		return nil, err
	}
	defer files.Close(ctx)
	for files.Next(ctx) {
		var file File
		err := files.Decode(&file)
		if err != nil {
			utils.ErrorLog("error FindPublicFiles", err.Error())
			return nil, err
		}
		result = append(result, file)
	}

	return result, nil
}

// FindCompanyFiles 获取公司所有公开的文件
func FindCompanyFiles(db, domain, fileName, contentType string) (f []File, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"folder_id":  "company",
		"domain":     domain,
		"owners":     bson.M{"$in": []string{domain}},
	}

	// 文件名不为空的场合，添加到查询条件中
	if fileName != "" {
		query["file_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(fileName), Options: "m"}}
	}

	// 文件类型不为空的场合，添加到查询条件中
	if contentType != "" {
		query["content_type"] = contentType
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindCompanyFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []File
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	files, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindCompanyFiles", err.Error())
		return nil, err
	}
	defer files.Close(ctx)
	for files.Next(ctx) {
		var file File
		err := files.Decode(&file)
		if err != nil {
			utils.ErrorLog("error FindCompanyFiles", err.Error())
			return nil, err
		}
		result = append(result, file)
	}

	return result, nil
}

// FindUserFiles 获取用户的所有文件
func FindUserFiles(db, userID, domain, fileName, contentType string) (u []File, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"folder_id":  "user",
		"domain":     domain,
		"owners":     bson.M{"$in": []string{userID}},
	}

	// 文件名不为空的场合，添加到查询条件中
	if fileName != "" {
		query["file_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(fileName), Options: "m"}}
	}

	// 文件类型不为空的场合，添加到查询条件中
	if contentType != "" {
		query["content_type"] = contentType
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindUserFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []File
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	files, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindUserFiles", err.Error())
		return nil, err
	}
	defer files.Close(ctx)
	for files.Next(ctx) {
		var file File
		err := files.Decode(&file)
		if err != nil {
			utils.ErrorLog("error FindUserFiles", err.Error())
			return nil, err
		}
		result = append(result, file)
	}

	return result, nil
}

// FindFiles 获取当前文件夹的所有的文件
func FindFiles(db, folderID, domain, fileName, contentType string) (u []File, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"domain":     domain,
		"folder_id":  folderID,
	}

	// 文件名不为空的场合，添加到查询条件中
	if fileName != "" {
		query["file_name"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(fileName), Options: "m"}}
	}

	// 文件类型不为空的场合，添加到查询条件中
	if contentType != "" {
		query["content_type"] = contentType
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []File
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	files, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindUserFiles", err.Error())
		return nil, err
	}
	defer files.Close(ctx)
	for files.Next(ctx) {
		var file File
		err := files.Decode(&file)
		if err != nil {
			utils.ErrorLog("error FindUserFiles", err.Error())
			return nil, err
		}
		result = append(result, file)
	}

	return result, nil
}

// FindFile 通过ID获取文件信息
func FindFile(db, fileID string) (f File, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result File
	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		utils.ErrorLog("error FindFile", err.Error())
		return result, err
	}
	// 默认过滤掉被软删除的数据
	query := bson.M{
		"deleted_by": "",
		"_id":        objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindFile", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindFile", err.Error())
		return result, err
	}

	return result, nil
}

// AddFile 添加文件
func AddFile(db string, f *File) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	f.ID = primitive.NewObjectID()
	f.FileID = f.ID.Hex()

	queryJSON, _ := json.Marshal(f)
	utils.DebugLog("AddFile", fmt.Sprintf("AddFile: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, f)
	if err != nil {
		utils.ErrorLog("error AddFile", err.Error())
		return "", err
	}

	return f.FileID, nil
}

// DeleteFile 删除单个文件
func DeleteFile(db, fileID, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		utils.ErrorLog("error DeleteFile", err.Error())
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
	utils.DebugLog("DeleteFile", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("DeleteFile", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error DeleteFile", err.Error())
		return err
	}

	return nil
}

// DeleteSelectFiles 删除多个文件
func DeleteSelectFiles(db string, fileIDList []string, userID string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var modifiedCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteSelectFiles", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteSelectFiles", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, fileID := range fileIDList {
			objectID, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				utils.ErrorLog("error DeleteSelectFiles", err.Error())
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
			utils.DebugLog("DeleteSelectFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("DeleteSelectFiles", fmt.Sprintf("update: [ %s ]", updateSON))

			result, err := c.UpdateOne(sc, query, update)
			modifiedCount = result.ModifiedCount
			if err != nil {
				utils.ErrorLog("error DeleteSelectFiles", err.Error())
				return err
			}

		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteSelectFiles", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteSelectFiles", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//result.ModifiedCount是int64，转换为int
	strInt64 := strconv.FormatInt(modifiedCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}

// HardDeleteFile 硬删除单个文件
func HardDeleteFile(db, fileID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		utils.ErrorLog("error HardDeleteFile", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("HardDeleteFile", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err = c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error HardDeleteFile", err.Error())
		return err
	}

	return nil
}

// HardDeleteFiles 物理删除多个文件
func HardDeleteFiles(db string, fileIDList []string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var deletedCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error HardDeleteFiles", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error HardDeleteFiles", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, fileID := range fileIDList {
			objectID, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				utils.ErrorLog("error HardDeleteFiles", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}

			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("HardDeleteFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

			result, err := c.DeleteOne(sc, query)
			deletedCount = result.DeletedCount
			if err != nil {
				utils.ErrorLog("error HardDeleteFiles", err.Error())
				return err
			}
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error HardDeleteFiles", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error HardDeleteFiles", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//result.DeletedCount是int64，转换为int
	strInt64 := strconv.FormatInt(deletedCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}

// DeleteFolderFile 删除文件夹文件
func DeleteFolderFile(db, folderID, userID string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// query := bson.M{
	// 	"folder_id": folderID,
	// }

	// update := bson.M{"$set": bson.M{
	// 	"deleted_at": time.Now(),
	// 	"deleted_by": userID,
	// }}

	// queryJSON, _ := json.Marshal(query)
	// utils.InfoLog("DeleteFolderFile", "DeleteFolderFile", fmt.Sprintf("query: [ %s ]", queryJSON))

	// updateSON, _ := json.Marshal(update)
	// utils.InfoLog("DeleteFolderFile", "DeleteFolderFile", fmt.Sprintf("update: [ %s ]", updateSON))

	// change, err := c.UpdateMany(ctx, query, update)
	// if err != nil {
	// 	utils.ErrorLog("error DeleteFolderFile", err.Error())
	// 	return 0, err
	// }
	var modifiedCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteFolderFile", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteFolderFile", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		query := bson.M{
			"folder_id": folderID,
		}

		update := bson.M{"$set": bson.M{
			"deleted_at": time.Now(),
			"deleted_by": userID,
		}}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("DeleteFolderFile", fmt.Sprintf("query: [ %s ]", queryJSON))

		updateSON, _ := json.Marshal(update)
		utils.DebugLog("DeleteFolderFile", fmt.Sprintf("update: [ %s ]", updateSON))

		result, err := c.UpdateMany(sc, query, update)
		modifiedCount = result.ModifiedCount
		if err != nil {
			utils.ErrorLog("error DeleteFolderFile", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteFolderFile", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteFolderFile", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//change.ModifiedCount是int64，转换为int
	strInt64 := strconv.FormatInt(modifiedCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}

// RecoverSelectFiles 恢复选中文件
func RecoverSelectFiles(db string, fileIDList []string, userID string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var modifiedCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverSelectFiles", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverSelectFiles", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, fileID := range fileIDList {
			objectID, err := primitive.ObjectIDFromHex(fileID)
			if err != nil {
				utils.ErrorLog("error RecoverSelectFiles", err.Error())
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
			utils.DebugLog("RecoverSelectFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

			updateSON, _ := json.Marshal(update)
			utils.DebugLog("RecoverSelectFiles", fmt.Sprintf("update: [ %s ]", updateSON))
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
				utils.ErrorLog("error RecoverSelectFiles", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverSelectFiles", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//result.ModifiedCount是int64，转换为int
	strInt64 := strconv.FormatInt(modifiedCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}

// RecoverFolderFiles 恢复文件夹文件
func RecoverFolderFiles(db, folderID, userID string) (count int, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(FilesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var modifiedCount int64
	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error RecoverFolderFiles", err.Error())
		return 0, err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error RecoverFolderFiles", err.Error())
		return 0, err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		query := bson.M{
			"folder_id": folderID,
		}

		update := bson.M{"$set": bson.M{
			"updated_at": time.Now(),
			"updated_by": userID,
			"deleted_by": "",
		}}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("RecoverFolderFiles", fmt.Sprintf("query: [ %s ]", queryJSON))

		updateSON, _ := json.Marshal(update)
		utils.DebugLog("RecoverFolderFiles", fmt.Sprintf("update: [ %s ]", updateSON))

		result, err := c.UpdateMany(sc, query, update)
		modifiedCount = result.ModifiedCount
		if err != nil {
			utils.ErrorLog("error RecoverFolderFiles", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error RecoverFolderFiles", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error RecoverFolderFiles", err.Error())
		return 0, err
	}
	session.EndSession(ctx)

	//change.ModifiedCount是int64，转换为int
	strInt64 := strconv.FormatInt(modifiedCount, 10)
	id16, _ := strconv.Atoi(strInt64)

	return id16, nil
}
