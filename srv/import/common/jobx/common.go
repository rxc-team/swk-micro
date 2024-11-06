package jobx

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/srv/import/common/loggerx"
	"rxcsoft.cn/pit3/srv/import/system/wsx"
	"rxcsoft.cn/pit3/srv/task/proto/task"
)

// 任务数据
type Task struct {
	JobId        string   `json:"job_id"`
	JobName      string   `json:"job_name,omitempty"`
	Origin       string   `json:"origin,omitempty"`
	UserId       string   `json:"user_id,omitempty"`
	ShowProgress bool     `json:"show_progress,omitempty"`
	Progress     int64    `json:"progress,omitempty"`
	StartTime    string   `json:"start_time,omitempty"`
	EndTime      string   `json:"end_time,omitempty"`
	Message      string   `json:"message,omitempty"`
	File         *File    `json:"file,omitempty"`
	ErrorFile    *File    `json:"error_file,omitempty"`
	TaskType     string   `json:"task_type,omitempty"`
	Steps        []string `json:"steps,omitempty"`
	CurrentStep  string   `json:"current_step,omitempty"`
	ScheduleId   string   `json:"schedule_id,omitempty"`
	AppId        string   `json:"app_id,omitempty"`
	Insert       int64    `json:"insert,omitempty"`
	Update       int64    `json:"update,omitempty"`
	Total        int64    `json:"total,omitempty"`
}

type File struct {
	Url  string `json:"url,omitempty"`
	Name string `json:"name,omitempty"`
}

var (
	errNotExecutionTime = errors.New("not to this task execution time")
	errExpired          = errors.New("this task has expired")
)

func CreateTask(req task.AddRequest) error {
	taskService := task.NewTaskService("task", client.DefaultClient)
	_, err := taskService.AddTask(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("ModifyTask", err.Error())
		return err
	}

	tk := Task{
		JobId:        req.JobId,
		JobName:      req.JobName,
		Origin:       req.Origin,
		UserId:       req.UserId,
		ShowProgress: req.ShowProgress,
		Progress:     req.Progress,
		StartTime:    req.StartTime,
		Message:      req.Message,
		TaskType:     req.TaskType,
		Steps:        req.Steps,
		CurrentStep:  req.CurrentStep,
		ScheduleId:   req.ScheduleId,
		AppId:        req.AppId,
	}

	content, err := json.Marshal(tk)
	if err != nil {
		loggerx.ErrorLog("ModifyTask", err.Error())
		return err
	}

	wsx.SendMsg(wsx.MessageParam{
		Sender:    "system",
		Recipient: req.GetUserId(),
		MsgType:   "job",
		Content:   string(content),
	})

	return nil
}

func ModifyTask(req task.ModifyRequest, userId string) error {
	taskService := task.NewTaskService("task", client.DefaultClient)
	_, err := taskService.ModifyTask(context.TODO(), &req)
	if err != nil {
		loggerx.ErrorLog("ModifyTask", err.Error())
		return err
	}

	var file File
	if req.File != nil {
		file.Url = req.File.Url
		file.Name = req.File.Name
	}

	var efile File
	if req.ErrorFile != nil {
		efile.Url = req.ErrorFile.Url
		efile.Name = req.ErrorFile.Name
	}

	tk := Task{
		JobId:       req.JobId,
		Progress:    req.Progress,
		EndTime:     req.EndTime,
		Message:     req.Message,
		File:        &file,
		ErrorFile:   &efile,
		CurrentStep: req.CurrentStep,
		Insert:      req.Insert,
		Update:      req.Update,
		Total:       req.Total,
	}

	content, err := json.Marshal(tk)
	if err != nil {
		loggerx.ErrorLog("ModifyTask", err.Error())
		return err
	}
	wsx.SendMsg(wsx.MessageParam{
		Sender:    "system",
		Recipient: userId,
		MsgType:   "job",
		Content:   string(content),
	})

	return nil
}
