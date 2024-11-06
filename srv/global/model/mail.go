package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"rxcsoft.cn/pit3/srv/global/proto/mail"
	"rxcsoft.cn/pit3/srv/global/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	MailsCollection = "mails"
)

// Mail 邮件
type Mail struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	Sender     string             `json:"sender" bson:"sender"`
	Recipients []string           `json:"recipients" bson:"recipients"`
	Ccs        []string           `json:"ccs" bson:"ccs"`
	Subject    string             `json:"subject" bson:"subject"`
	Content    string             `json:"content" bson:"content"`
	Annex      string             `json:"annex" bson:"annex"`
	SendTime   string             `json:"send_time" bson:"send_time"`
}

// ToProto 转换为proto数据
func (m *Mail) ToProto() *mail.Mail {
	return &mail.Mail{
		Sender:     m.Sender,
		Recipients: m.Recipients,
		Ccs:        m.Ccs,
		Subject:    m.Subject,
		Content:    m.Content,
		Annex:      m.Annex,
		SendTime:   m.SendTime,
	}
}

// FindMails 获取邮件
func FindMails(db, recipient, cc, subject, content, annex, sendTime string) (m []Mail, err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MailsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	// 收件人不为空
	if recipient != "" {
		query["recipients"] = bson.M{"$in": []string{recipient}}
	}

	// 抄送人不为空
	if cc != "" {
		query["ccs"] = bson.M{"$in": []string{cc}}
	}

	// 主题不为空
	if subject != "" {
		query["subject"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(subject), Options: "m"}}
	}

	// 内容不为空
	if content != "" {
		query["content"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(content), Options: "m"}}
	}

	// 附件不为空
	if annex != "" {
		query["annex"] = annex
	}

	// 发送时间不为空
	if sendTime != "" {
		query["send_time"] = sendTime
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindMails", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Mail
	cur, err := c.Find(ctx, query)
	if err != nil {
		utils.ErrorLog("error FindCustomers", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var ma Mail
		err := cur.Decode(&ma)
		if err != nil {
			utils.ErrorLog("error FindCustomers", err.Error())
			return nil, err
		}
		result = append(result, ma)
	}

	return result, nil
}

// AddMail 添加邮件记录
func AddMail(db string, m *Mail) (err error) {
	client := database.New()
	c := client.Database(database.GetDBName(db)).Collection(MailsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m.ID = primitive.NewObjectID()

	queryJSON, _ := json.Marshal(m)
	utils.DebugLog("FindMails", fmt.Sprintf("Mail: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, m)
	if err != nil {
		utils.ErrorLog("error AddMail", err.Error())
		return err
	}

	return nil
}
