package eventx

import (
	"encoding/json"

	"github.com/micro/go-micro/v2/broker"
	"rxcsoft.cn/pit3/api/internal/system/jobx"
	"rxcsoft.cn/pit3/srv/task/proto/schedule"
)

// AddJob 创建job
func AddJob(event broker.Event) error {
	// 从body中获取数据
	var sc *schedule.Schedule
	err := json.Unmarshal(event.Message().Body, &sc)
	if err != nil {
		return err
	}

	// 添加任务
	job := new(jobx.Job)

	if sc.RunNow {
		job.Run(sc)
	} else {
		job.Add(sc)
	}

	event.Ack()

	return nil
}

// StopJob 停止job
func StopJob(event broker.Event) error {
	// 从body中获取数据
	var sc *schedule.Schedule
	err := json.Unmarshal(event.Message().Body, &sc)
	if err != nil {
		return err
	}

	// 添加任务
	job := new(jobx.Job)
	job.Stop(sc)

	event.Ack()

	return nil
}

// DeleteJob 删除job
func DeleteJob(event broker.Event) error {
	// 从body中获取数据
	var sc *schedule.Schedule
	err := json.Unmarshal(event.Message().Body, &sc)
	if err != nil {
		return err
	}

	// 删除任务
	job := new(jobx.Job)
	err = job.Remove(sc)
	if err != nil {
		return err
	}

	event.Ack()

	return nil
}
