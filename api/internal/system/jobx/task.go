package jobx

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/micro/go-micro/v2/client"
	"github.com/robfig/cron/v3"
	"github.com/rxc-team/dcron"
	"gopkg.in/gomail.v2"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/global/proto/mail"
	config "rxcsoft.cn/pit3/srv/global/proto/mail-config"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
	cf "rxcsoft.cn/utils/config"
)

type (
	// Job 任务
	Job struct{}

	// Result 任务结果
	Result struct {
		Result     string
		Err        error
		RetryTimes int64
	}
)

var (
	// 定时任务调度管理器
	serviceCron *dcron.Dcron

	// 并发队列, 限制同时运行的任务数量
	concurrencyQueue ConcurrencyQueue
)

// getRedisConf 获取redis的连接信息
func getRedisConf() *redis.Options {
	// 获取mongo的配置
	cfe := cf.GetConf(cf.RedisKey)
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfe.Host, cfe.Port),
		Password: cfe.Password,
		PoolSize: 10,
	}
}

// Initialize 初始化任务, 从数据库取出所有任务, 添加到定时任务并运行
func (j Job) Initialize() {
	opt := getRedisConf()

	serviceCron = dcron.NewDcron("job-server1", opt)
	serviceCron.Start()
	concurrencyQueue = ConcurrencyQueue{queue: make(chan struct{}, 500)}

	loggerx.SystemLog(false, false, "Initialize", "start init job")
	taskNum := 0

	// 添加系统任务
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)
	var req schedule.SchedulesRequest
	req.RunNow = false
	req.Database = "system"
	response, err := scheduleService.FindSchedules(context.TODO(), &req)
	if err != nil {
		loggerx.SystemLog(true, true, "Initialize", err.Error())
		return
	}

	scheduleList := response.GetSchedules()
	for _, item := range scheduleList {
		j.Add(item)
		taskNum++
	}

	// 获取已有的顾客
	customerService := customer.NewCustomerService("manage", client.DefaultClient)
	var csReq customer.FindCustomersRequest
	csRes, err := customerService.FindCustomers(context.TODO(), &csReq)
	if err != nil {
		loggerx.SystemLog(true, true, "Initialize", err.Error())
		return
	}
	// 获取已经创建的任务列表
	for _, cs := range csRes.GetCustomers() {
		// 获取当前顾客下的任务
		var req schedule.SchedulesRequest
		req.RunNow = false
		req.Database = cs.GetCustomerId()
		response, err := scheduleService.FindSchedules(context.TODO(), &req)
		if err != nil {
			loggerx.SystemLog(true, true, "Initialize", err.Error())
			return
		}

		scheduleList := response.GetSchedules()
		for _, item := range scheduleList {
			j.Add(item)
			taskNum++
		}
	}

	loggerx.SystemLog(false, false, "Initialize", fmt.Sprintf("Init job success, %d job add to the scheduler", taskNum))
}

// BatchAdd 批量添加任务
func (j Job) BatchAdd(schedules []*schedule.Schedule) {
	for _, item := range schedules {
		j.RemoveAndAdd(item)
	}
}

// RemoveAndAdd 删除任务后添加
func (j Job) RemoveAndAdd(schedule *schedule.Schedule) {
	j.Remove(schedule)
	j.Add(schedule)
}

// Add 添加任务
func (j Job) Add(sc *schedule.Schedule) int {
	taskFunc := createJob(sc)
	if taskFunc == nil {
		loggerx.SystemLog(true, true, "Add", "create job has error")
		return 0
	}

	id, err := serviceCron.AddFunc(sc.ScheduleId, sc.Spec, taskFunc)
	if err != nil {
		loggerx.SystemLog(true, true, "Add", fmt.Sprintf("add job to the scheduler, has error", err.Error()))
		return 0
	}

	// 打印出下一次执行时间
	nextTime := serviceCron.Next(sc.ScheduleId)
	loggerx.SystemLog(false, false, "Add", fmt.Sprintf("job[%v] next time %v", sc.ScheduleName, nextTime))

	// 修改任务的对应的状态
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)
	var mReq schedule.ModifyRequest
	mReq.ScheduleId = sc.GetScheduleId()
	mReq.EntryId = strconv.Itoa(int(id))
	// 从共通中获取
	mReq.Writer = sc.GetCreatedBy()
	mReq.Database = sc.Params["db"]

	_, err = scheduleService.ModifySchedule(context.TODO(), &mReq)
	if err != nil {
		return int(id)
	}

	return int(id)
}

// HealthCheck 检查任务健康状态
func (j Job) HealthCheck(id string) bool {
	entry := serviceCron.GetJobInfo(id)

	return entry.ID != 0
}

// Remove 删除任务
func (j Job) Remove(sc *schedule.Schedule) error {
	loggerx.SystemLog(false, false, "Remove", fmt.Sprintf("Remove the [%s] job from the scheduler", sc.ScheduleName))
	serviceCron.Remove(sc.ScheduleId)

	// 删除当前任务
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)
	var mReq schedule.DeleteRequest
	mReq.ScheduleIds = []string{sc.ScheduleId}
	mReq.Database = sc.Params["db"]

	_, err := scheduleService.DeleteSchedule(context.TODO(), &mReq)
	if err != nil {
		return err
	}

	return nil
}

// Stop 停止当前任务
func (j Job) Stop(sc *schedule.Schedule) error {
	loggerx.SystemLog(false, false, "Remove", fmt.Sprintf("Remove the [%s] job from the scheduler", sc.ScheduleName))
	serviceCron.Remove(sc.ScheduleId)

	// 修改任务的对应的状态
	scheduleService := schedule.NewScheduleService("task", client.DefaultClient)
	var mReq schedule.ModifyRequest
	mReq.ScheduleId = sc.ScheduleId
	mReq.EntryId = "0"
	// 从共通中获取
	mReq.Writer = sc.GetCreatedBy()
	mReq.Database = sc.Params["db"]

	_, err := scheduleService.ModifySchedule(context.TODO(), &mReq)
	if err != nil {
		return err
	}

	return nil
}

// WaitAndExit 等待所有任务结束后退出
func (j Job) WaitAndExit() {
	serviceCron.Stop()
}

// Run 直接运行任务
func (j Job) Run(schedule *schedule.Schedule) {
	go createJob(schedule)()
}

// 创建任务
func createJob(schedule *schedule.Schedule) cron.FuncJob {
	handler := createHandler(schedule)
	if handler == nil {
		return nil
	}
	taskFunc := func() {
		ok := beforeExecJob(schedule)
		if !ok {
			return
		}

		concurrencyQueue.Add()
		defer concurrencyQueue.Done()

		loggerx.SystemLog(false, false, "execJob", fmt.Sprintf("strart execution the [%s] job", schedule.ScheduleName))
		taskResult := execJob(handler, schedule)
		loggerx.SystemLog(false, false, "execJob", fmt.Sprintf("execution the [%s] job has success", schedule.ScheduleName))
		afterExecJob(schedule, taskResult)
	}

	loggerx.SystemLog(false, false, "createJob", fmt.Sprintf("create a job [%s]", schedule.ScheduleName))

	return taskFunc
}

func createHandler(schedule *schedule.Schedule) Handler {
	var handler Handler = nil
	switch schedule.ScheduleType {
	case "db-backup":
		handler = new(BackHandler)
	case "db-restore":
		handler = new(RestoreHandler)
	case "db-backup-clean":
		handler = new(ClearHandler)
	}

	return handler
}

// 任务前置操作
func beforeExecJob(s *schedule.Schedule) bool {
	// rand.Seed(time.Now().Unix())
	// nodes := runInstance.getNodes(s.Params["db"], s.ScheduleId)
	// node := nodes[rand.Intn(len(nodes)-1)]

	// if node == nodeID {
	// 	return true
	// }
	return true
}

// 任务执行后置操作
func afterExecJob(schedule *schedule.Schedule, taskResult Result) {
	// // 删除任务运行实例
	// if schedule.Multi == 0 {
	// 	// runInstance.done(schedule.ScheduleId)
	// }

	loggerx.SystemLog(false, false, "createJob", fmt.Sprintf("the [%s] job has end", schedule.ScheduleName))

	// 发送邮件
	go SendNotification(schedule, taskResult)
}

// SendNotification 发送任务结果通知
func SendNotification(schedule *schedule.Schedule, taskResult Result) {
	db := schedule.Params["db"]
	userID := schedule.CreatedBy

	userService := user.NewUserService("manage", client.DefaultClient)
	var req user.FindUserRequest
	req.Database = db
	req.UserId = userID

	res, err := userService.FindUser(context.TODO(), &req)
	if err != nil {
		loggerx.SystemLog(false, true, "SendNotification", fmt.Sprintf("find user has error: %v", err))
		return
	}

	email := res.GetUser().GetNoticeEmail()
	if len(email) > 0 {
		// 定义收件人
		mailTo := []string{
			email,
		}
		// 定义抄送人
		mailCcTo := []string{}
		// 邮件主题
		subject := "Task completed"
		// 邮件正文
		// body := fmt.Sprintf("The [%s] job from schedule has executed successfully", schedule.ScheduleName)
		tpl := template.Must(template.ParseFiles("assets/html/task.html"))
		params := map[string]string{
			"name": schedule.ScheduleName,
		}

		var out bytes.Buffer
		err := tpl.Execute(&out, params)
		if err != nil {
			loggerx.SystemLog(false, true, "SendNotification", fmt.Sprintf("send mail has error: %v", err))
			return
		}

		er := sendMail(db, mailTo, mailCcTo, subject, out.String(), "")
		if er != nil {
			return
		}
	}
}

// Handler 任务执行接口
type Handler interface {
	Run(schedule *schedule.Schedule) (string, error)
}

// 执行具体任务
func execJob(handler Handler, schedule *schedule.Schedule) Result {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic#service/Job.go:execJob#", err)
		}
	}()
	output, err := handler.Run(schedule)
	if err == nil {
		return Result{Result: output, Err: err, RetryTimes: 1}
	}
	// // 默认只运行任务一次
	// var execTimes int64 = 1
	// if schedule.RetryTimes > 0 {
	// 	execTimes += schedule.RetryTimes
	// }
	// var i int64 = 0
	// var output string
	// var err error
	// for i < execTimes {
	// 	output, err = handler.Run(schedule)
	// 	if err == nil {
	// 		return Result{Result: output, Err: err, RetryTimes: i}
	// 	}
	// 	i++
	// 	if i < execTimes {
	// 		fmt.Printf("任务执行失败#任务id-%s#重试第%d次#输出-%s#错误-%s", schedule.ScheduleId, i, output, err.Error())
	// 		if schedule.RetryInterval > 0 {
	// 			time.Sleep(time.Duration(schedule.RetryInterval) * time.Second)
	// 		} else {
	// 			// 默认重试间隔时间，每次递增1分钟
	// 			time.Sleep(time.Duration(i) * time.Minute)
	// 		}
	// 	}
	// }

	return Result{Result: output, Err: err, RetryTimes: schedule.RetryTimes}
}

// sendMail 发邮件
func sendMail(db string, mailTo []string, mailCcTo []string, subject string, body string, attachFile string) error {
	// 获取邮箱服务器连接信息
	configService := config.NewConfigService("global", client.DefaultClient)

	var req config.FindConfigsRequest
	req.Database = "system"
	res, err := configService.FindConfigs(context.TODO(), &req)
	if err != nil {
		return err
	}

	// 邮箱配置信息不存在
	if len(res.Configs) < 1 {
		return fmt.Errorf("config info not exist")
	}
	// 邮箱配置信息复数存在
	if len(res.Configs) > 1 {
		return fmt.Errorf("too much config")
	}

	// 连接邮箱服务器信息校验
	if res.Configs[0].Host == "" {
		return fmt.Errorf("The Host cannot be empty")
	}
	if res.Configs[0].Port == "" {
		return fmt.Errorf("The Port cannot be empty")
	}
	if res.Configs[0].Mail == "" {
		return fmt.Errorf("The Mail cannot be empty")
	}
	if res.Configs[0].Password == "" {
		return fmt.Errorf("The Password cannot be empty")
	}
	// 转换端口类型为int
	port, e := strconv.Atoi(res.Configs[0].Port)
	if e != nil {
		return e
	}
	// 连接邮箱服务器
	d := gomail.NewDialer(res.Configs[0].Host, port, res.Configs[0].Mail, res.Configs[0].Password)

	// 编辑邮件信息
	m := gomail.NewMessage()
	// 设置带别名送件人
	nickName := "Proship"
	m.SetHeader("From", m.FormatAddress(res.Configs[0].Mail, nickName))
	// 设置收件人
	if len(mailTo) > 0 {
		m.SetHeader("To", mailTo...)
	} else {
		return fmt.Errorf("Recipient cannot be empty")
	}
	// 设置抄送人
	if len(mailCcTo) > 0 {
		m.SetHeader("Cc", mailCcTo...)
	}
	// 设置邮件主题
	if subject != "" {
		m.SetHeader("Subject", subject)
	}
	// 设置邮件正文
	if body != "" {
		m.SetBody("text/html", body)
	}
	// 设置附件名
	if attachFile != "" {
		pos := strings.LastIndex(attachFile, "/")
		name := attachFile[pos+1:]
		m.Attach(attachFile,
			gomail.Rename(name),
		)
	}

	// 发送编辑好的邮件
	er := d.DialAndSend(m)
	if er != nil {
		return er
	}

	// 添加邮件发送记录
	mailService := mail.NewMailService("global", client.DefaultClient)

	var AddReq mail.AddMailRequest

	AddReq.Database = db

	AddReq.Sender = res.Configs[0].Mail
	AddReq.Recipients = mailTo
	AddReq.Ccs = mailCcTo
	AddReq.Subject = subject
	AddReq.Content = body
	AddReq.Annex = attachFile
	AddReq.SendTime = time.Now().Format("2006-01-02 15:04:05")

	_, addErr := mailService.AddMail(context.TODO(), &AddReq)
	if addErr != nil {
		return addErr
	}
	return nil
}
