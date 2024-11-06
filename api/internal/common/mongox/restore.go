package mongox

import (
	"strings"

	"rxcsoft.cn/pit3/api/internal/common/cmdx"
	"rxcsoft.cn/utils/config"
)

// MongoResotre mongodb恢复参数
type MongoResotre struct {
	host           string
	port           string
	database       string
	username       string
	password       string
	replicaSetName string
	DbSuffix       string
	Oplog          bool
	DumpPath       string
}

// NewResotre 返回一个backup对象
func NewResotre() *MongoResotre {
	// 获取mongo的配置
	cfe := config.GetConf(config.MongoKey)

	backup := MongoResotre{
		host:           cfe.Host,
		port:           cfe.Port,
		database:       cfe.Database,
		username:       cfe.Username,
		password:       cfe.Password,
		replicaSetName: cfe.ReplicaSetName,
	}

	return &backup
}

func (mongo *MongoResotre) setName() string {
	if len(mongo.DbSuffix) > 0 {
		return "--nsInclude " + mongo.database + "_" + mongo.DbSuffix + ".*"
	}
	return "--nsInclude " + mongo.database + ".*"
}

func (mongo *MongoResotre) setCredential() string {
	opts := []string{}
	if len(mongo.username) > 0 {
		opts = append(opts, "--username="+mongo.username)
	}
	if len(mongo.password) > 0 {
		opts = append(opts, `--password=`+"\""+mongo.password+"\"")
	}
	return strings.Join(opts, " ")
}

func (mongo *MongoResotre) setURL() string {
	opts := []string{}
	hosts := strings.Split(mongo.host, ",")
	hostURI := strings.Builder{}
	if len(hosts) > 1 {
		hostURI.WriteString("--host=")
		hostURI.WriteString(mongo.replicaSetName)
		hostURI.WriteString("/")
		for i, host := range hosts {
			hostURI.WriteString(host)
			if i < len(hosts)-1 {
				hostURI.WriteString(",")
			}
		}

		opts = append(opts, hostURI.String())
		opts = append(opts, "--authenticationDatabase=admin")
	} else {
		hostURI.WriteString("--host=")
		if mongo.replicaSetName != "" {
			hostURI.WriteString(mongo.replicaSetName)
			hostURI.WriteString("/")
		}
		hostURI.WriteString(mongo.host)
		opts = append(opts, hostURI.String())
		if !strings.Contains(mongo.host, ":") {
			opts = append(opts, "--port="+mongo.port)
		}
		opts = append(opts, "--authenticationDatabase=admin")
	}

	return strings.Join(opts, " ")
}

func (mongo *MongoResotre) setOplog() string {
	if mongo.Oplog {
		return "--oplog"
	}

	return ""
}

// MongoRestore 执行恢复操作
func (mongo *MongoResotre) MongoRestore() error {

	cmd := "mongorestore" + " " +
		mongo.setName() + " " +
		mongo.setCredential() + " " +
		mongo.setURL() + " " +
		"--drop " +
		mongo.DumpPath

	return cmdx.ExecCommand(cmd)
}
