/*
 * @Descripttion:
 * @Author: Rxc 陳平
 * @Date: 2020-10-19 09:20:34
 * @LastEditors: Rxc 陳平
 * @LastEditTime: 2020-10-19 13:27:45
 */
package model

import (
	"testing"

	"rxcsoft.cn/utils/config"
	database1 "rxcsoft.cn/utils/mongo"
)

func TestAddConfig(t *testing.T) {
	mongo := config.DB{
		Host:           "172.18.32.40",
		Port:           "27017",
		Username:       "root",
		Database:       "pit",
		Password:       "admin",
		ReplicaSetName: "mongos",
		Source:         "",
	}

	// before
	database1.StartMongodb(mongo)
	type args struct {
		db string
		mc *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		//TODO: Add test cases.
		{
			name: "添加一条配置",
			args: args{
				db: "5f87c039972895febab68b4b",
				mc: &Config{
					ConfigID:  "5f87bfcff9599894d7511102",
					Mail:      "chenping@rxcsoft.cn",
					Password:  "Rxc@2018",
					Host:      "smtp.yunyou.top",
					Port:      "465",
					Ssl:       "false",
					CreatedBy: "5f86c6fb95e2cb1ed02c571b",
					UpdatedBy: "5f86c6fb95e2cb1ed02c571b",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AddConfig(tt.args.db, tt.args.mc)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if gotId != "" {
			// 	t.Errorf("AddConfig() = %v, want %v", gotId, nil)
			// }
		})
	}
}
