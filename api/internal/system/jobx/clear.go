package jobx

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-plugins/broker/rabbitmq/v2"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// ClearHandler 清除备份数据
type ClearHandler struct {
}

// Run 执行数据库备份操作
func (b *ClearHandler) Run(schedule *schedule.Schedule) (result string, err error) {
	var now string
	if schedule.Spec != "" {
		// 提取计划时区名称
		scheduleTimezoneName := schedule.Spec[strings.Index(schedule.Spec, "=")+1 : strings.Index(schedule.Spec, " ")]
		// 通过时区名称获取时区
		scheduleTimezone, err := time.LoadLocation(scheduleTimezoneName)
		if err != nil {
			loggerx.SystemLog(true, true, "run", err.Error())
			return "", err
		}
		// 获取指定时区的时间
		now = time.Now().In(scheduleTimezone).Format("2006-01-02")
	} else {
		// 获取本地时区的时间
		now = time.Now().Local().Format("2006-01-02")
	}
	if schedule.StartTime > now {
		loggerx.SystemLog(true, true, "run", errNotExecutionTime.Error())
		return "", errNotExecutionTime
	}
	if schedule.EndTime < now {
		// 过期了，删除该任务
		body, err := json.Marshal(schedule)
		if err != nil {
			return "", err
		}

		br := rabbitmq.NewBroker()

		br.Publish("job.delete", &broker.Message{
			Body: body,
		})
		// 提示任务已过期
		loggerx.SystemLog(true, true, "run", errExpired.Error())
		return "", errExpired
	}

	err = dbBackupClear(schedule)
	if err != nil {
		loggerx.SystemLog(true, true, "run", err.Error())
		return "", err
	}

	return "ok", nil
}

func dbBackupClear(schedule *schedule.Schedule) error {
	db := schedule.Params["db"]
	domain := schedule.Params["domain"]
	appID := schedule.Params["app_id"]
	jobID := "job_" + time.Now().Format("20060102150405")
	userID := schedule.CreatedBy

	go func() {

		CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      schedule.ScheduleName,
			ScheduleId:   schedule.ScheduleId,
			Origin:       "-",
			UserId:       userID,
			ShowProgress: false,
			Message:      "ジョブを作成します",
			TaskType:     "db-backup-clean",
			Steps:        []string{"start", "find-backups", "delete-record-file", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})

		// 发送消息 查找备份记录
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "バックアップファイルを探す",
			CurrentStep: "find-backups",
			Database:    db,
		}, userID)

		req := backup.FindBackupsRequest{
			BackupType: "database",
			Database:   db,
		}

		backupService := backup.NewBackupService("manage", client.DefaultClient)

		response, err := backupService.FindBackups(context.TODO(), &req)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 处理失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "find-backups",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 查找备份记录
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "7日前にバックアップファイルを削除します",
			CurrentStep: "delete-record-file",
			Database:    db,
		}, userID)

		var now string
		if schedule.Spec != "" {
			// 提取计划时区名称
			scheduleTimezoneName := schedule.Spec[strings.Index(schedule.Spec, "=")+1 : strings.Index(schedule.Spec, " ")]
			// 通过时区名称获取时区
			scheduleTimezone, err := time.LoadLocation(scheduleTimezoneName)
			if err != nil {
				path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
				// 发送消息 处理失败，终止任务
				ModifyTask(task.ModifyRequest{
					JobId:       jobID,
					Message:     err.Error(),
					CurrentStep: "delete-record-file",
					EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
					ErrorFile: &task.File{
						Url:  path.MediaLink,
						Name: path.Name,
					},
					Database: db,
				}, userID)
				return
			}
			// 获取指定时区的时间
			now = time.Now().In(scheduleTimezone).AddDate(0, 0, -7).Format("2006-01-02")
		} else {
			// 获取本地时区的时间
			now = time.Now().Local().AddDate(0, 0, -7).Format("2006-01-02")
		}

		//  超级域名（从proship的桶中下载备份文件）
		minioClient, err := storagecli.NewClient("backups")
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 处理失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "delete-record-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		for _, bk := range response.GetBackups() {
			createDate := bk.GetCreatedAt()[0:10]
			if createDate < now {

				req := backup.HardDeleteBackupsRequest{
					BackupIdList: []string{bk.GetBackupId()},
					Database:     db,
				}

				_, err := backupService.HardDeleteBackups(context.TODO(), &req)
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 处理失败，终止任务
					ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "delete-record-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)
					return
				}

				err = minioClient.DeleteObject(bk.GetFileName())
				if err != nil {
					path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
					// 发送消息 处理失败，终止任务
					ModifyTask(task.ModifyRequest{
						JobId:       jobID,
						Message:     err.Error(),
						CurrentStep: "delete-record-file",
						EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
						ErrorFile: &task.File{
							Url:  path.MediaLink,
							Name: path.Name,
						},
						Database: db,
					}, userID)
					return
				}

			}
		}

		// 发送消息 恢复成功，任务结束
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ジョブ実行の成功",
			CurrentStep: "end",
			EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
			Database:    db,
		}, userID)
	}()
	return nil
}
