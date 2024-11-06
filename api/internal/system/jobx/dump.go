package jobx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-plugins/broker/rabbitmq/v2"
	"rxcsoft.cn/pit3/api/internal/common/filex"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/mongox"
	"rxcsoft.cn/pit3/api/internal/common/storagex"
	"rxcsoft.cn/pit3/srv/manage/proto/backup"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	"rxcsoft.cn/pit3/srv/task/proto/task"
	storagecli "rxcsoft.cn/utils/storage/client"
)

// BackHandler 备份
type BackHandler struct {
}

// Run 执行数据库备份操作
func (b *BackHandler) Run(schedule *schedule.Schedule) (result string, err error) {
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
	err = dbdoump(schedule)
	if err != nil {
		loggerx.SystemLog(true, true, "run", err.Error())
		return "", err
	}
	return "ok", nil
}

func dbdoump(schedule *schedule.Schedule) error {
	db := schedule.Params["db"]
	domain := schedule.Params["domain"]
	appID := schedule.Params["app_id"]
	jobID := "job_" + time.Now().Format("20060102150405")
	userID := schedule.CreatedBy

	// 获取当前时间戳
	timestamp := time.Now().Format("20060102150405")

	go func() {

		CreateTask(task.AddRequest{
			JobId:        jobID,
			JobName:      schedule.ScheduleName,
			ScheduleId:   schedule.ScheduleId,
			Origin:       "-",
			UserId:       userID,
			ShowProgress: false,
			Message:      "ジョブを作成します",
			TaskType:     "db-backup",
			Steps:        []string{"start", "build-cmd-execute", "zip-file", "save-file", "create-backup-record", "end"},
			CurrentStep:  "start",
			Database:     db,
			AppId:        appID,
		})

		// 发送消息 编辑备份命令,并执行
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "cmdをビルドしてcmdを実行します",
			CurrentStep: "build-cmd-execute",
			Database:    db,
		}, userID)

		req := backup.AddBackupRequest{
			BackupName: schedule.ScheduleName,
			BackupType: "database",
			Database:   db,
			Writer:     schedule.CreatedBy,
		}

		// 获取已有的顾客
		customerService := customer.NewCustomerService("manage", client.DefaultClient)
		var csReq customer.FindCustomersRequest
		csReq.InvalidatedIn = "true"
		csRes, err := customerService.FindCustomers(context.TODO(), &csReq)
		if err != nil {
			loggerx.SystemLog(true, true, "Initialize", err.Error())
			return
		}

		backupInfo := make(map[string]string)

		backupInfo["system"] = domain

		// 备份dev数据
		{
			// 备份system数据库数据
			sback := mongox.NewBackup()
			sback.DbSuffix = "system"
			sback.Oplog = true
			sback.Out = "./backups/" + timestamp + "/" + "system"
			// 开始备份
			err := sback.Mongodump()
			if err != nil {
				os.Remove("./backups/" + timestamp)
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

			// 备份全局数据库数据
			back := mongox.NewBackup()
			back.Oplog = true
			back.Out = "./backups/" + timestamp + "/" + "system"
			// 开始备份
			err = back.Mongodump()
			if err != nil {
				os.Remove("./backups/" + timestamp)
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

			filesPath := "./backups/" + timestamp + "/" + "system" + "/files/"
			// 创建文件夹
			filex.Mkdir(filesPath)
			// 备份文件
			reclone := storagex.NewConf(domain, filesPath)
			// 文件备份-顾客单位
			err = reclone.MinioCopy()
			if err != nil {
				os.Remove("./backups/" + timestamp)
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
		}

		// 获取已经创建的任务列表
		for _, cs := range csRes.GetCustomers() {

			backupInfo[cs.GetCustomerId()] = cs.GetDomain()

			back := mongox.NewBackup()
			back.DbSuffix = cs.GetCustomerId()
			back.Oplog = true
			back.Out = "./backups/" + timestamp + "/" + cs.GetCustomerId()
			// 开始备份
			err := back.Mongodump()
			if err != nil {
				os.Remove("./backups/" + timestamp)
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

			filesPath := "./backups/" + timestamp + "/" + cs.GetCustomerId() + "/files/"
			// 创建文件夹
			filex.Mkdir(filesPath)
			// 备份文件
			reclone := storagex.NewConf(cs.GetDomain(), filesPath)
			// 文件备份-顾客单位
			err = reclone.MinioCopy()
			if err != nil {
				os.Remove("./backups/" + timestamp)
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
		}

		err = saveFile(backupInfo, "backups/"+timestamp+"/backup.json")
		if err != nil {
			os.Remove("./backups/" + timestamp)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "zip-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 发送消息 编辑备份命令,并执行
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "バックアップファイルを圧縮します",
			CurrentStep: "zip-file",
			Database:    db,
		}, userID)

		dir := "backups/" + timestamp

		zipFileName := "backups/backups_" + timestamp + ".zip"
		// 压缩文件
		filex.Zip(dir, zipFileName)

		fo, err := os.Open(zipFileName)
		if err != nil {
			os.Remove("./backups/" + timestamp)
			os.Remove(zipFileName)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "zip-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}
		defer fo.Close()
		// 发送消息 编辑备份命令,并执行
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "ファイルをクラウドに保存",
			CurrentStep: "save-file",
			Database:    db,
		}, userID)
		minioClient, err := storagecli.NewClient("backups")
		if err != nil {
			os.RemoveAll(zipFileName)
			os.RemoveAll(dir)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		filePath := path.Join("zip", "backups_"+timestamp+".zip")
		file, err := minioClient.SavePublicObject(fo, filePath, "application/x-zip-compressed")
		if err != nil {
			os.RemoveAll(zipFileName)
			os.RemoveAll(dir)
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "save-file",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

		// 删除压缩文件
		fo.Close()
		os.RemoveAll(zipFileName)
		os.RemoveAll(dir)

		// 发送消息 编辑备份命令,并执行
		ModifyTask(task.ModifyRequest{
			JobId:       jobID,
			Message:     "バックアップレコードを作成します",
			CurrentStep: "create-backup-record",
			File: &task.File{
				Url:  file.MediaLink,
				Name: file.Name,
			},
			Database: db,
		}, userID)

		// 设置文件信息
		req.FileName = file.Name
		req.Size = file.Size
		req.FilePath = file.MediaLink
		backupService := backup.NewBackupService("manage", client.DefaultClient)

		_, err = backupService.AddBackup(context.TODO(), &req)
		if err != nil {
			path := filex.WriteAndSaveFile(domain, appID, []string{err.Error()})
			// 发送消息 获取数据失败，终止任务
			ModifyTask(task.ModifyRequest{
				JobId:       jobID,
				Message:     err.Error(),
				CurrentStep: "create-backup-record",
				EndTime:     time.Now().UTC().Format("2006-01-02 15:04:05"),
				ErrorFile: &task.File{
					Url:  path.MediaLink,
					Name: path.Name,
				},
				Database: db,
			}, userID)
			return
		}

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

// saveFile 创建并保存文件本地
func saveFile(data map[string]string, fileName string) (e error) {

	// 保存该数据到文件
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	bytesBuffer := &bytes.Buffer{}

	writer := bufio.NewWriter(bytesBuffer)
	writer.Write(jsonData)
	writer.Flush()

	// 创建文件
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(bytesBuffer.Bytes())
	if err != nil {
		return err
	}
	return nil
}
