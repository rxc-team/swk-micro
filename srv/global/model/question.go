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
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"rxcsoft.cn/pit3/srv/global/proto/question"
	"rxcsoft.cn/pit3/srv/global/utils"

	"rxcsoft.cn/utils/helpers"
	database "rxcsoft.cn/utils/mongo"
)

// 集合
var (
	QuestionsCollection = "questions"
)

// Question 问题
type Question struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	QuestionID     string             `json:"question_id" bson:"question_id"`
	Title          string             `json:"title" bson:"title"`
	Type           string             `json:"type" bson:"type"`
	Function       string             `json:"function" bson:"function"`
	Images         []string           `json:"images" bson:"images"`
	Content        string             `json:"content" bson:"content"`
	Domain         string             `json:"domain" bson:"domain"`
	QuestionerName string             `json:"questioner_name" bson:"questioner_name"`
	ResponderID    string             `json:"responder_id" bson:"responder_id"`
	ResponderName  string             `json:"responder_name" bson:"responder_name"`
	Postscripts    []Postscript       `json:"postscripts" bson:"postscripts"`
	Status         string             `json:"status" bson:"status"`
	Locations      string             `json:"locations" bson:"locations"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy      string             `json:"created_by" bson:"created_by"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy      string             `json:"updated_by" bson:"updated_by"`
}

// Postscript 追记
type Postscript struct {
	Postscripter     string    `json:"postscripter" bson:"postscripter"`
	PostscripterName string    `json:"postscripter_name" bson:"postscripter_name"`
	Avatar           string    `json:"avatar" bson:"avatar"`
	Content          string    `json:"content" bson:"content"`
	Link             string    `json:"link" bson:"link"`
	Images           []string  `json:"images" bson:"images"`
	PostscriptedAt   time.Time `json:"postscripted_at" bson:"postscripted_at"`
}

// ToProto 转换为proto数据
func (q *Question) ToProto() *question.Question {
	var ps []*question.Postscript
	for _, p := range q.Postscripts {
		ps = append(ps, p.ToProto())
	}

	return &question.Question{
		QuestionId:     q.QuestionID,
		Title:          q.Title,
		Type:           q.Type,
		Function:       q.Function,
		Images:         q.Images,
		Content:        q.Content,
		Domain:         q.Domain,
		QuestionerName: q.QuestionerName,
		ResponderId:    q.ResponderID,
		ResponderName:  q.ResponderName,
		Postscripts:    ps,
		Locations:      q.Locations,
		Status:         q.Status,
		CreatedAt:      q.CreatedAt.String(),
		CreatedBy:      q.CreatedBy,
		UpdatedAt:      q.UpdatedAt.String(),
		UpdatedBy:      q.UpdatedBy,
	}
}

// ToProto 转换为proto数据
func (p *Postscript) ToProto() *question.Postscript {
	return &question.Postscript{
		Postscripter:     p.Postscripter,
		PostscripterName: p.PostscripterName,
		Avatar:           p.Avatar,
		Content:          p.Content,
		Link:             p.Link,
		Images:           p.Images,
		PostscriptedAt:   p.PostscriptedAt.String(),
	}
}

// FindQuestion 获取单个问题
func FindQuestion(questionID string) (q Question, err error) {
	client := database.New()
	opts := options.Collection()
	opts.SetReadPreference(readpref.Primary())
	c := client.Database(database.Db).Collection(QuestionsCollection, opts)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var result Question
	objectID, err := primitive.ObjectIDFromHex(questionID)
	if err != nil {
		utils.ErrorLog("error FindQuestion", err.Error())
		return result, err
	}
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindQuestion", fmt.Sprintf("query: [ %s ]", queryJSON))

	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		utils.ErrorLog("error FindQuestion", err.Error())
		return result, err
	}

	return result, nil
}

// FindQuestions 获取多个问题
func FindQuestions(title, questionType, function, status, domain string) (q []Question, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(QuestionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := bson.M{}

	// 问题标题非空
	if title != "" {
		query["title"] = bson.M{"$regex": primitive.Regex{Pattern: helpers.Escape(title), Options: "m"}}
	}
	// 问题类型非空
	if questionType != "" {
		query["type"] = questionType
	}
	// 问题发生位置(机能别)非空
	if function != "" {
		query["function"] = function
	}
	// 问题处理状态非空
	if status != "" {
		query["status"] = status
	}
	// 域名非空
	if domain != "" {
		query["domain"] = domain
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("FindQuestions", fmt.Sprintf("query: [ %s ]", queryJSON))

	var result []Question
	sortItem := bson.D{
		{Key: "created_at", Value: -1},
	}
	opts := options.Find().SetSort(sortItem)
	cur, err := c.Find(ctx, query, opts)
	if err != nil {
		utils.ErrorLog("error FindQuestions", err.Error())
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var q Question
		err := cur.Decode(&q)
		if err != nil {
			utils.ErrorLog("error FindQuestions", err.Error())
			return nil, err
		}
		result = append(result, q)
	}

	return result, nil
}

// AddQuestion 添加问题
func AddQuestion(q *Question) (id string, err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(QuestionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	q.ID = primitive.NewObjectID()
	q.QuestionID = q.ID.Hex()
	if q.Postscripts == nil {
		q.Postscripts = make([]Postscript, 0)
	}

	queryJSON, _ := json.Marshal(q)
	utils.DebugLog("AddQuestion", fmt.Sprintf("Question: [ %s ]", queryJSON))

	_, err = c.InsertOne(ctx, q)
	if err != nil {
		utils.ErrorLog("error AddMessage", err.Error())
		return q.QuestionID, err
	}
	return q.QuestionID, nil
}

// ModifyQuestion 更新问题
func ModifyQuestion(q *Question) (err error) {
	client := database.New()
	c := client.Database(database.Db).Collection(QuestionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(q.QuestionID)
	if err != nil {
		utils.ErrorLog("error ModifyCustomer", err.Error())
		return err
	}
	query := bson.M{
		"_id": objectID,
	}

	change := bson.M{
		"updated_at": q.UpdatedAt,
		"updated_by": q.UpdatedBy,
	}

	// 问题状态不为空的场合
	if q.Status != "" {
		change["status"] = q.Status
	}

	update := bson.D{
		{Key: "$set", Value: change},
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("ModifyQuestion", fmt.Sprintf("query: [ %s ]", queryJSON))

	updateSON, _ := json.Marshal(update)
	utils.DebugLog("ModifyQuestion", fmt.Sprintf("update: [ %s ]", updateSON))

	_, err = c.UpdateOne(ctx, query, update)
	if err != nil {
		utils.ErrorLog("error ModifyQuestion", err.Error())
		return err
	}

	// 追记
	if len(q.Postscripts) > 0 {
		objectID, err := primitive.ObjectIDFromHex(q.QuestionID)
		if err != nil {
			utils.ErrorLog("error ModifyQuestion", err.Error())
			return err
		}
		qp := bson.M{
			"_id": objectID,
		}

		up := bson.M{
			"$addToSet": bson.M{
				"postscripts": q.Postscripts[0],
			},
		}

		queryJSON, _ := json.Marshal(qp)
		utils.DebugLog("ModifyQuestion", fmt.Sprintf("query: [ %s ]", queryJSON))

		updateSON, _ := json.Marshal(up)
		utils.DebugLog("ModifyQuestion", fmt.Sprintf("update: [ %s ]", updateSON))

		_, err = c.UpdateOne(ctx, qp, up)
		if err != nil {
			utils.ErrorLog("error postscripts", err.Error())
			return err
		}
	}
	return nil
}

// DeleteQuestion 硬删除问题
func DeleteQuestion(questionID string) error {
	client := database.New()
	c := client.Database(database.Db).Collection(QuestionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(questionID)
	query := bson.M{
		"_id": objectID,
	}

	queryJSON, _ := json.Marshal(query)
	utils.DebugLog("DeleteQuestion", fmt.Sprintf("query: [ %s ]", queryJSON))

	_, err := c.DeleteOne(ctx, query)
	if err != nil {
		utils.ErrorLog("error DeleteQuestion", err.Error())
		return err
	}

	return nil
}

// DeleteQuestions 硬删除多个问题
func DeleteQuestions(questionIDList []string) error {
	client := database.New()
	c := client.Database(database.Db).Collection(QuestionsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	session, err := client.StartSession()
	if err != nil {
		utils.ErrorLog("error DeleteQuestions", err.Error())
		return err
	}
	if err = session.StartTransaction(); err != nil {
		utils.ErrorLog("error DeleteQuestions", err.Error())
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		for _, questionID := range questionIDList {
			objectID, err := primitive.ObjectIDFromHex(questionID)
			if err != nil {
				utils.ErrorLog("error DeleteQuestions", err.Error())
				return err
			}
			query := bson.M{
				"_id": objectID,
			}
			queryJSON, _ := json.Marshal(query)
			utils.DebugLog("DeleteQuestions", fmt.Sprintf("query: [ %s ]", queryJSON))

			_, err = c.DeleteOne(ctx, query)
			if err != nil {
				utils.ErrorLog("error DeleteQuestions", err.Error())
				return err
			}
		}
		if err = session.CommitTransaction(sc); err != nil {
			if err != nil {
				session.AbortTransaction(ctx)
				utils.ErrorLog("error DeleteQuestions", err.Error())
				return err
			}
		}
		return nil
	}); err != nil {
		session.AbortTransaction(ctx)
		utils.ErrorLog("error DeleteQuestions", err.Error())
		return err
	}
	session.EndSession(ctx)
	return nil
}
