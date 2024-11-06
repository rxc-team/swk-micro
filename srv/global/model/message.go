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
	"rxcsoft.cn/pit3/srv/global/proto/message"
	"rxcsoft.cn/pit3/srv/global/utils"

	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	MessagesCollection = "messages"
)

// Message 通知
type Message struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	MessageID string             `json:"message_id" bson:"message_id"`
	Domain    string             `json:"domain" bson:"domain"`
	Sender    string             `json:"sender" bson:"sender"`
	Recipient string             `json:"recipient" bson:"recipient"`
	MsgType   string             `json:"msg_type" bson:"msg_type"`
	Code      string             `json:"code" bson:"code"`
	Link      string             `json:"link" bson:"link"`
	Content   string             `json:"content" bson:"content"`
	Status    string             `json:"status" bson:"status"`
	Object    string             `json:"object" bson:"object"`
	SendTime  time.Time          `json:"send_time" bson:"send_time"`
	EndTime   time.Time          `json:"end_time" bson:"end_time,omitempty"`
}

// ToProto 转换为proto数据
func (m *Message) ToProto() *message.Message {
	return &message.Message{
		MessageId: m.MessageID,
		Domain:    m.Domain,
		Sender:    m.Sender,
		Recipient: m.Recipient,
		MsgType:   m.MsgType,
		Code:      m.Code,
		Link:      m.Link,
		Content:   m.Content,
		Status:    m.Status,
		Object:    m.Object,
		SendTime:  m.SendTime.String(),
		EndTime:   m.EndTime.String(),
	}
}

// FindMessage 获取单个通知
func FindMessage(db, messageID string) (m Message, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MessagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result Message
	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		utils.ErrorLog("error FindMessage", err.Error())
		return result, err
	}
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindMessage", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindMessage", err.Error())
		return result, err
	}

	return result, nil
}

// FindMessages 获取多个通知
func FindMessages(db, recipient, domain, status, msgType string, limit, skip int64) (m []Message, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MessagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	// 接受者非空
	// if recipient != "" {
	query["recipient"] = recipient
	// }
	// 公司域名非空
	// if domain != "" {
	query["domain"] = domain
	// }
	// 通知状态
	if status != "" {
		query["status"] = status
	}
	// 消息类型
	if msgType != "" {
		query["msg_type"] = msgType
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindMessages", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Message
	opts := options.Find().SetSort(bson.D{{Key: "send_time", Value: -1}})
	if limit != 0 {
		if skip != 0 {
			opts.SetSkip((skip - 1) * limit).SetLimit(limit)
		} else {
			opts.SetLimit(limit)
		}
	}

	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindMessages", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var m Message
		err := cur.Decode(&m)
		if err != nil {
			utils.ErrorLog("error FindMessages", err.Error())
			return nil, err
		}
		result = append(result, m)
	}

	return result, nil
}

// FindUpdateMessage 获取系统更新通知
func FindUpdateMessage(db, domain, recipient string, now_time time.Time) (m Message, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MessagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}
	query["msg_type"] = "update"
	query["status"] = "unread"
	query["recipient"] = recipient
	query["domain"] = domain
	query["end_time"] = bson.M{
		"$gte": now_time,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindUpdateMessage", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result Message
	opts := options.Find().SetSort(bson.D{{Key: "send_time", Value: -1}})
	opts.SetLimit(1)

	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindUpdateMessage", err.Error())
		return result, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var m Message
		err := cur.Decode(&m)
		if err != nil {
			utils.ErrorLog("error FindUpdateMessage", err.Error())
			return result, err
		}
		result = m
	}

	return result, nil
}

// AddMessage 添加通知
func AddMessage(db string, m *Message) (id string, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MessagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 标注过时系统消息为已读
	if m.MsgType == "update" {
		query := bson.M{
			"msg_type":  "update",
			"status":    "unread",
			"recipient": m.Recipient,
			"domain":    m.Domain,
		}
		update := bson.M{
			"$set": bson.M{
				"status": "read",
			},
		}
		_, err := c.UpdateMany(ctx, query, update)
		if err != nil {
			utils.ErrorLog("error AddMessage", err.Error())
			return "", err
		}
	}

	m.ID = primitive.NewObjectID()
	m.MessageID = m.ID.Hex()

	queryJSON, _ := json.Marshal(m)
	utils.DebugLog("AddMessage", fmt.Sprintf("Message: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, m)
	if err != nil {
		utils.ErrorLog("error AddMessage", err.Error())
		return m.MessageID, err
	}
	return m.MessageID, nil
}

// ChangeStatus 变更通知状态
func ChangeStatus(db, messageID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MessagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(messageID)
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ChangeStatus", fmt.Sprintf("query: [ %s ]", queryJSON))

	update := bson.M{
		"$set": bson.M{
			"status": "read",
		},
	}

	_, err := c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ChangeStatus", err.Error())
		return err
	}
	return nil
}

// DeleteMessage 硬删除通知
func DeleteMessage(db, messageID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MessagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(messageID)
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteMessage", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err := c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteMessage", err.Error())
		return err
	}
	return nil
}

// DeleteMessages 硬删除多个通知
func DeleteMessages(db string, userID string) error {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MessagesCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteHelps", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteHelps", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		query := bson.M{
			"recipient": userID,
			"status":    "read",
		}

		queryJSON, _ := json.Marshal(query)
		utils.DebugLog("DeleteMessages", fmt.Sprintf("query: [ %s ]", queryJSON))

		_, err = c.DeleteMany(sc, query)
		if err != nil {
			utils.ErrorLog("error DeleteHelps", err.Error())
			return err
		}

		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteHelps", err.Error())
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

	return nil
}
