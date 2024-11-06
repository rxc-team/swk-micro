package jobx

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-plugins/broker/rabbitmq/v2"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/mongox"
	"rxcsoft.cn/pit3/api/internal/common/storagex"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	"rxcsoft.cn/utils/mq"
)

// RestoreHandler 恢复
type RestoreHandler struct {
}

// Run 执行数据库备份操作
func (b *RestoreHandler) Run(schedule *schedule.Schedule) (result string, err error) {
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

	err = dbrestore(schedule)
	if err != nil {
		loggerx.SystemLog(true, true, "run", err.Error())
		return "", err
	}

	domain := schedule.Params["domain"]
	clientIP := schedule.Params["client_ip"]
	userID := schedule.CreatedBy

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.FindUserRequest
	req.Type = 0
	req.UserId = userID
	req.Database = schedule.Params["db"]
	userInfo, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		return "", err
	}
	params := map[string]string{
		"user_name":   userInfo.GetUser().GetUserName(),
		"backup_name": schedule.ScheduleName,
	}
	loggerx.JobProcessLog(userID, domain, clientIP, "Restore", msg.L016, params)

	return "ok", nil
}

func dbrestore(schedule *schedule.Schedule) error {
	db := schedule.Params["db"]
	backupDb := schedule.Params["backup_db"]
	backupDomain := schedule.Params["backup_domain"]
	domain := schedule.Params["domain"]
	appID := schedule.Params["app_id"]
	localPath := schedule.Params["local_path"]
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
			TaskType:     "db-restore",
			Steps:        []string{"start", "unzip-file", "build-cmd-execute", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})
		// 发送消息 解压文件
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "バックアップファイルをローカルに解凍します",
			CurrentStep: "unzip-file",
			Database:    db,
		}, userID)

		// 解压文件
		path, err := filex.UnZip(localPath, filepath.Dir(localPath), "utf-8")
		if err != nil {
			os.Remove(localPath)
			os.RemoveAll(filepath.Dir(localPath))
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 处理失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "unzip-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		err = os.Remove(localPath)
		if err != nil {
			os.RemoveAll(path)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息，处理失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "unzip-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 编辑恢复命令,并执行
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "cmdをビルドしてcmdを実行します",
			CurrentStep: "build-cmd-execute",
			Database:    db,
		}, userID)

		restore := mongox.NewResotre()
		restore.DbSuffix = backupDb
		// mongo := config.GetConf("mongo")

		restore.DumpPath = "./" + path + "/" + backupDb
		// 开始恢复
		err = restore.MongoRestore()
		if err != nil {
			os.RemoveAll(path)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "build-cmd-execute",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 文件恢复
		reclone := storagex.NewConf(backupDomain, path+"/"+backupDb+"/files")

		// 文件恢复-顾客单位
		err = reclone.MinioSync()
		if err != nil {
			os.RemoveAll(path)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "backup-file-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}
		// 删除解压后的问题
		os.RemoveAll(path)

		// 删除任务
		body, err := json.Marshal(schedule)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "backup-file-data",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		bk := mq.NewBroker()

		bk.Publish("job.delete", &broker.Message{
			Body: body,
		})

		// 发送消息 写入保存文件成功，返回下载路径，任务结束
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
