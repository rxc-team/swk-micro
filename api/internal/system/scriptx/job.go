package scriptx

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/micro/go-micro/v2/client"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/srv/manage/proto/script"
)

type ScriptJob struct {
	s DataPatchScript
}

func NewJob(sid string) *ScriptJob {
	return &ScriptJob{
		s: createScript(sid),
	}
}

func (j *ScriptJob) ExecJob() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("panic#scriptx/job.go:ExecJob#", err)
		}
	}()

	err := j.s.Run()
	if err != nil {
		loggerx.ErrorLog("execJob", err.Error())
	}
}

func Init() {
	var list []script.AddRequest
	var scriptIDs []string

	// 2021-09-29 追加了用户权限的相关后台验证，需要追加用户与角色关系
	list = append(list, script.AddRequest{
		ScriptId:      "0929",
		ScriptName:    "ユーザーとロールの関係を追加します",
		ScriptDesc:    "ユーザー権限の検証を追加しました。ユーザーとロールの関係を追加する必要があります",
		ScriptType:    "function",
		ScriptVersion: "2.4.1",
		CreatedAt:     "2021-09-29",
		Writer:        "system",
		Database:      "system",
	})

	script1110, err := ioutil.ReadFile("assets/scripts/script1110.js")
	if err != nil {
		panic(err)
	}

	script1216, err := ioutil.ReadFile("assets/scripts/script1216.js")
	if err != nil {
		panic(err)
	}

	script0124, err := ioutil.ReadFile("assets/scripts/script0124.js")
	if err != nil {
		panic(err)
	}
	script0419, err := ioutil.ReadFile("assets/scripts/script0419.js")
	if err != nil {
		panic(err)
	}
	script0420, err := ioutil.ReadFile("assets/scripts/script0420.js")
	if err != nil {
		panic(err)
	}

	script20220517, err := ioutil.ReadFile("assets/scripts/script20220517.js")
	if err != nil {
		panic(err)
	}

	script20220518, err := ioutil.ReadFile("assets/scripts/script20220518.js")
	if err != nil {
		panic(err)
	}

	script20220616, err := ioutil.ReadFile("assets/scripts/script20220616.js")
	if err != nil {
		panic(err)
	}
	script20220620, err := ioutil.ReadFile("assets/scripts/script20220620.js")
	if err != nil {
		panic(err)
	}
	script20220708, err := ioutil.ReadFile("assets/scripts/script20220708.js")
	if err != nil {
		panic(err)
	}
	script20220805, err := ioutil.ReadFile("assets/scripts/script20220805.js")
	if err != nil {
		panic(err)
	}
	script20220810, err := ioutil.ReadFile("assets/scripts/script20220810.js")
	if err != nil {
		panic(err)
	}
	script20230807, err := ioutil.ReadFile("assets/scripts/script20230807.js")
	if err != nil {
		panic(err)
	}
	/* script20230904, err := ioutil.ReadFile("assets/scripts/script20230904.js")
	if err != nil {
		panic(err)
	} */
	script20230926, err := ioutil.ReadFile("assets/scripts/script20230926.js")
	if err != nil {
		panic(err)
	}
	script20230927, err := ioutil.ReadFile("assets/scripts/script20230927.js")
	if err != nil {
		panic(err)
	}
	script20231017, err := ioutil.ReadFile("assets/scripts/script20231017.js")
	if err != nil {
		panic(err)
	}
	script20231024, err := ioutil.ReadFile("assets/scripts/script20231024.js")
	if err != nil {
		panic(err)
	}
	script20231130, err := ioutil.ReadFile("assets/scripts/script20231130.js")
	if err != nil {
		panic(err)
	}
	script20231201, err := ioutil.ReadFile("assets/scripts/script20231201.js")
	if err != nil {
		panic(err)
	}
	script20231204, err := ioutil.ReadFile("assets/scripts/script20231204.js")
	if err != nil {
		panic(err)
	}
	script20240410, err := ioutil.ReadFile("assets/scripts/script20240410.js")
	if err != nil {
		panic(err)
	}

	// 2021-11-10 修改了自动採番字段实现规则，需要重新设置该字段的数据
	list = append(list, script.AddRequest{
		ScriptId:      "1110",
		ScriptName:    "autonumフィールド値をリセットします",
		ScriptDesc:    "autonumフィールドの実装ルールを変更しました。フィールドのデータをリセットする必要があります",
		ScriptType:    "javascript",
		ScriptData:    "{}",
		ScriptFunc:    string(script1110),
		ScriptVersion: "2.4.2",
		CreatedAt:     "2021-11-10",
		Writer:        "system",
		Database:      "system",
	})

	// 2021-12-16 修改流程多语言的从属位置，需要重新设置流程多语言情报和流程名称情报
	list = append(list, script.AddRequest{
		ScriptId:      "1216",
		ScriptName:    "プロセスの多言語情報とプロセス名情報をリセットします",
		ScriptDesc:    "プロセスの多言語の所属位置を変更しました。プロセスの多言語情報とプロセス名情報をリセットする必要があります",
		ScriptType:    "javascript",
		ScriptData:    "{}",
		ScriptFunc:    string(script1216),
		ScriptVersion: "2.4.3",
		CreatedAt:     "2021-12-16",
		Writer:        "system",
		Database:      "system",
	})

	// 2022-01-24 添加数据履历映射下载，需要更新台账映射情报
	list = append(list, script.AddRequest{
		ScriptId:      "0124",
		ScriptName:    "台帳マッピング情報を更新します",
		ScriptDesc:    "履歴マッピングダウンロード機能を追加するため、台帳マッピング情報を更新する必要があります",
		ScriptType:    "javascript",
		ScriptData:    "{}",
		ScriptFunc:    string(script0124),
		ScriptVersion: "2.4.3",
		CreatedAt:     "2022-01-24",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-03-24 修改台账的文件在minio中的路径，并同步修改数据库的路径
	list = append(list, script.AddRequest{
		ScriptId:      "0324",
		ScriptName:    "台帳のファイルのパスを更新します",
		ScriptDesc:    "minioで台帳ファイルのパスを変更し、データベースのパスを同期的に変更します",
		ScriptType:    "function",
		ScriptVersion: "2.4.5",
		CreatedAt:     "2022-03-24",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-04-19 添加台账的唯一属性的字段id的数组，包括自动采番字段的字段id
	list = append(list, script.AddRequest{
		ScriptId:      "0419",
		ScriptName:    "台帳情報の一意のフィールド情報を追加します",
		ScriptDesc:    "自動インクリメントフィールドを含む、一意のフィールド情報の配列を台帳に追加します",
		ScriptType:    "javascript",
		ScriptData:    "{}",
		ScriptFunc:    string(script0419),
		ScriptVersion: "2.4.5",
		CreatedAt:     "2022-04-19",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-04-20 添加台账的lookup字段所关联的台账的信息的数组
	list = append(list, script.AddRequest{
		ScriptId:      "0420",
		ScriptName:    "台帳を追加するための関連情報の配列",
		ScriptDesc:    "台帳のルックアップフィールドに関連付けられた台帳情報の配列を追加します",
		ScriptType:    "javascript",
		ScriptData:    "{}",
		ScriptFunc:    string(script0420),
		ScriptVersion: "2.4.5",
		CreatedAt:     "2022-04-20",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-05-17 修正履历集合的索引问题
	list = append(list, script.AddRequest{
		ScriptId:      "20220517",
		ScriptName:    "履歴のインデックスを修正",
		ScriptDesc:    "履歴のインデックスを修正し、間違ったインデックスを削除します",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20220517),
		ScriptVersion: "2.4.5",
		CreatedAt:     "2022-05-17",
		Writer:        "system",
		Database:      "system",
	})

	// 2022-05-18 修改app的config信息的存放位置
	list = append(list, script.AddRequest{
		ScriptId:      "20220518",
		ScriptName:    "アプリの設定情報のストレージコレクションを変更",
		ScriptDesc:    "アプリの設定情報のストレージコレクションを変更し、元の設定コレクションを削除します。",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20220518),
		ScriptVersion: "2.4.6",
		CreatedAt:     "2022-05-18",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-06-16 承認のために承認履歴情報のストレージを変更する
	list = append(list, script.AddRequest{
		ScriptId:      "20220616",
		ScriptName:    "承認のために承認履歴情報のストレージを変更します",
		ScriptDesc:    "承認テーブルの承認履歴情報を変更し、変更された履歴書のセットを追加します。",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20220616),
		ScriptVersion: "2.4.6",
		CreatedAt:     "2022-06-16",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-06-20 通常の履歴書のデータパッチ
	list = append(list, script.AddRequest{
		ScriptId:      "20220620",
		ScriptName:    "通常の履歴書のデータパッチ",
		ScriptDesc:    "通常の履歴書のデータパッチ、および通常の履歴書の保存方法を変更する。",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20220620),
		ScriptVersion: "2.4.6",
		CreatedAt:     "2022-06-20",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-07-08 通常の履歴書のデータパッチ
	list = append(list, script.AddRequest{
		ScriptId:      "20220708",
		ScriptName:    "履歴データのインデックスを追加。",
		ScriptDesc:    "履歴データのインデックスを追加。",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20220708),
		ScriptVersion: "2.4.6",
		CreatedAt:     "2022-07-08",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-08-05 タイ語マルチリンガルを追加
	list = append(list, script.AddRequest{
		ScriptId:      "20220805",
		ScriptName:    "タイ語マルチリンガルを追加",
		ScriptDesc:    "タイ語マルチリンガルを追加",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20220805),
		ScriptVersion: "2.4.7",
		CreatedAt:     "2022-08-05",
		Writer:        "system",
		Database:      "system",
	})
	// 2022-08-10 許可アクションの多言語データパッチ
	list = append(list, script.AddRequest{
		ScriptId:      "20220810",
		ScriptName:    "許可アクションの多言語データパッチ",
		ScriptDesc:    "許可アクションの多言語データパッチ",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20220810),
		ScriptVersion: "2.4.7",
		CreatedAt:     "2022-08-10",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-08-07 年と月のフィールドの更新
	list = append(list, script.AddRequest{
		ScriptId:      "20230807",
		ScriptName:    "年と月のフィールドの更新",
		ScriptDesc:    "年と月のフィールドの更新",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20230807),
		ScriptVersion: "1.1.1",
		CreatedAt:     "2023-08-07",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-09-04 フィールドの更新
	/* list = append(list, script.AddRequest{
		ScriptId:      "20230904",
		ScriptName:    "フィールドの更新",
		ScriptDesc:    "フィールドの更新",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20230904),
		ScriptVersion: "1.1.1",
		CreatedAt:     "2023-09-04",
		Writer:        "system",
		Database:      "system",
	}) */
	// 2023-09-26 利益剰余金と適用開始時点の残存リース料のフィールドの更新
	list = append(list, script.AddRequest{
		ScriptId:      "20230926",
		ScriptName:    "利益剰余金と適用開始時点の残存リース料のフィールドの更新",
		ScriptDesc:    "利益剰余金と適用開始時点の残存リース料のフィールドの更新",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20230926),
		ScriptVersion: "1.1.1",
		CreatedAt:     "2023-09-26",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-09-27 バックアップフィールドの削除
	list = append(list, script.AddRequest{
		ScriptId:      "20230927",
		ScriptName:    "バックアップフィールドの削除",
		ScriptDesc:    "バックアップフィールドの削除",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20230927),
		ScriptVersion: "1.1.1",
		CreatedAt:     "2023-09-27",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-10-17 年と月のフィールドの確認
	list = append(list, script.AddRequest{
		ScriptId:      "20231017",
		ScriptName:    "年と月のフィールドの確認",
		ScriptDesc:    "年と月のフィールドの確認",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20231017),
		ScriptVersion: "1.1.1",
		CreatedAt:     "2023-10-17",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-10-24 利益剰余金と適用開始時点の残存リース料のフィールドの追加
	list = append(list, script.AddRequest{
		ScriptId:      "20231024",
		ScriptName:    "利益剰余金と適用開始時点の残存リース料のフィールドの追加",
		ScriptDesc:    "利益剰余金と適用開始時点の残存リース料のフィールドの追加",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20231024),
		ScriptVersion: "1.1.1",
		CreatedAt:     "2023-10-24",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-11-30 フィールドの日付が表示されない
	list = append(list, script.AddRequest{
		ScriptId:      "20231130",
		ScriptName:    "フィールドの日付が表示されない",
		ScriptDesc:    "フィールドの日付が表示されない",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20231130),
		ScriptVersion: "1.1.4",
		CreatedAt:     "2023-11-30",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-12-01 フィールド年月ソートの変更
	list = append(list, script.AddRequest{
		ScriptId:      "20231201",
		ScriptName:    "フィールド年月ソートの変更",
		ScriptDesc:    "フィールド年月ソートの変更",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20231201),
		ScriptVersion: "1.1.4",
		CreatedAt:     "2023-12-01",
		Writer:        "system",
		Database:      "system",
	})
	// 2023-12-04 フィールドunique_fieldsの変更
	list = append(list, script.AddRequest{
		ScriptId:      "20231204",
		ScriptName:    "フィールドunique_fieldsの変更",
		ScriptDesc:    "フィールドunique_fieldsの変更",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20231204),
		ScriptVersion: "1.1.4",
		CreatedAt:     "2023-12-04",
		Writer:        "system",
		Database:      "system",
	})
	// 2024-04-10 適用開始期首月の変更
	list = append(list, script.AddRequest{
		ScriptId:      "20240410",
		ScriptName:    "適用開始期首月の変更",
		ScriptDesc:    "適用開始期首月の変更",
		ScriptType:    "javascript",
		ScriptData:    "{\"db\":\"pit\"}",
		ScriptFunc:    string(script20240410),
		ScriptVersion: "1.1.4",
		CreatedAt:     "2024-04-10",
		Writer:        "system",
		Database:      "system",
	})

	// 去除重复数据并生成唯一索引
	for _, script := range list {
		scriptIDs = append(scriptIDs, script.GetScriptId())
	}
	err = deleteDuplicateAndAddIndex(scriptIDs)
	if err != nil {
		loggerx.ErrorLog("deleteDuplicateAndAddIndex", err.Error())
		return
	}

	// 执行命令
	for _, req := range list {
		if !exist(req.Database, req.ScriptId) {
			add(req)
		}
	}
}

func exist(db, scriptId string) bool {
	scriptService := script.NewScriptService("manage", client.DefaultClient)
	ctxTmp, cancel := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancel()

	var req script.FindScriptJobRequest
	req.ScriptId = scriptId
	req.Database = db
	response, err := scriptService.FindScriptJob(ctxTmp, &req)
	if err != nil {
		return false
	}

	if response.ScriptJob != nil {
		return true
	}

	return false
}

func add(req script.AddRequest) error {
	scriptService := script.NewScriptService("manage", client.DefaultClient)
	ctxTmp, cancel := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancel()

	_, err := scriptService.AddScriptJob(ctxTmp, &req)
	if err != nil {
		loggerx.ErrorLog("add", err.Error())
		return err
	}

	return nil
}

func deleteDuplicateAndAddIndex(scriptIds []string) error {
	scriptService := script.NewScriptService("manage", client.DefaultClient)
	ctxTmp, cancel := context.WithTimeout(context.TODO(), 120*time.Second)
	defer cancel()
	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Minute * 10
	}
	var deleteReq script.DeleteScriptsRequest
	deleteReq.Database = "system"
	deleteReq.ScriptIds = scriptIds
	_, err := scriptService.DeleteDuplicateAndAddIndex(ctxTmp, &deleteReq, opss)
	if err != nil {
		loggerx.ErrorLog("delete", err.Error())
		return err
	}
	return nil
}
