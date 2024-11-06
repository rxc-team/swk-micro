package storagex

import (
	"errors"
	"fmt"
	"os"

	"rxcsoft.cn/pit3/api/internal/common/cmdx"
	"rxcsoft.cn/utils/config"
)

// MinioConf 存储配置
type MinioConf struct {
	Name   string
	Bucket string
	Path   string
}

// NewConf 返回一个backup对象
func NewConf(bucket, path string) *MinioConf {
	// 获取mongo的配置
	cfName := os.Getenv("CONFIG_NAME")

	cf := MinioConf{
		Name:   cfName,
		Bucket: bucket,
		Path:   path,
	}

	return &cf
}

func (cf *MinioConf) setName() string {
	storageConfig, err := config.GetStorageConf()
	if err != nil {
		panic(errors.New("storage config has error"))
	}
	bucketName := storageConfig.Bucket
	if len(cf.Bucket) > 0 {
		bucketName = fmt.Sprintf("%s-%s", storageConfig.Bucket, cf.Bucket)
	}

	return fmt.Sprintf("%s:%s", cf.Name, bucketName)
}

// MinioCopy 执行文件备份操作
func (cf *MinioConf) MinioCopy() error {

	cmd := "rclone" + " " +
		"copy" + " " +
		cf.setName() + " " +
		cf.Path

	return cmdx.ExecCommand(cmd)
}

// MinioSync 执行文件恢复操作
func (cf *MinioConf) MinioSync() error {

	cmd := "rclone" + " " +
		"sync" + " " +
		cf.Path + " " +
		cf.setName()

	return cmdx.ExecCommand(cmd)
}
