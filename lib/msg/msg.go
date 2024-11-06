package msg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// MessageType 消息类型
type MessageType int32

// Msg 消息
type Msg struct {
	Info   map[string]string `json:"info"`
	Warn   map[string]string `json:"warning"`
	Error  map[string]string `json:"error"`
	Logger map[string]string `json:"logger"`
}

const (
	// Info 提示
	Info MessageType = 0
	// Warn 警告
	Warn MessageType = 1
	// Error 错误
	Error MessageType = 2
	// Logger 出力系统日志
	Logger MessageType = 3
)

// 消息key
const (
	E001 = "E_001"
	E002 = "E_002"
	E003 = "E_003"
	E004 = "E_004"
	E005 = "E_005"
	E006 = "E_006"
	E007 = "E_007"
	E008 = "E_008"
	E009 = "E_009"
	E010 = "E_010"
	E011 = "E_011"
	E012 = "E_012"
	E013 = "E_013"
	E014 = "E_014"
	E099 = "E_099"

	W001 = "W_001"
	W002 = "W_002"
	W003 = "W_003"
	W004 = "W_004"
	W005 = "W_005"

	I001 = "I_001"
	I002 = "I_002"
	I003 = "I_003"
	I004 = "I_004"
	I005 = "I_005"
	I006 = "I_006"
	I007 = "I_007"
	I008 = "I_008"
	I009 = "I_009"
	I010 = "I_010"
	I011 = "I_011"
	I012 = "I_012"
	I013 = "I_013"
	I014 = "I_014"
	I015 = "I_015"
	I016 = "I_016"
	I017 = "I_017"
	I018 = "I_018"
	I100 = "I_100"
	I101 = "I_101"
	I102 = "I_102"
	I103 = "I_103"
	I104 = "I_104"
	I105 = "I_105"
	I106 = "I_106"
	I107 = "I_107"
	I108 = "I_108"
	I109 = "I_109"
	I110 = "I_110"
	I111 = "I_111"
	I112 = "I_112"
	I113 = "I_113"
	I114 = "I_114"
	I115 = "I_115"
	I116 = "I_116"
	I117 = "I_117"

	L001 = "L_001"
	L002 = "L_002"
	L003 = "L_003"
	L004 = "L_004"
	L005 = "L_005"
	L006 = "L_006"
	L007 = "L_007"
	L008 = "L_008"
	L009 = "L_009"
	L010 = "L_010"
	L011 = "L_011"
	L012 = "L_012"
	L013 = "L_013"
	L014 = "L_014"
	L015 = "L_015"
	L016 = "L_016"
	L017 = "L_017"
	L018 = "L_018"
	L019 = "L_019"
	L020 = "L_020"
	L021 = "L_021"
	L022 = "L_022"
	L023 = "L_023"
	L024 = "L_024"
	L025 = "L_025"
	L026 = "L_026"
	L027 = "L_027"
	L028 = "L_028"
	L029 = "L_029"
	L030 = "L_030"
	L031 = "L_031"
	L032 = "L_032"
	L033 = "L_033"
	L034 = "L_034"
	L035 = "L_035"
	L036 = "L_036"
	L037 = "L_037"
	L038 = "L_038"
	L039 = "L_039"
	L040 = "L_040"
	L041 = "L_041"
	L042 = "L_042"
	L043 = "L_043"
	L044 = "L_044"
	L045 = "L_045"
	L046 = "L_046"
	L047 = "L_047"
	L048 = "L_048"
	L049 = "L_049"
	L050 = "L_050"
	L051 = "L_051"
	L052 = "L_052"
	L053 = "L_053"
	L054 = "L_054"
	L055 = "L_055"
	L056 = "L_056"
	L057 = "L_057"
	L058 = "L_058"
	L059 = "L_059"
	L060 = "L_060"
	L061 = "L_061"
	L062 = "L_062"
	L063 = "L_063"
	L064 = "L_064"
	L065 = "L_065"
	L066 = "L_066"
	L067 = "L_067"
	L068 = "L_068"
	L069 = "L_069"
	L070 = "L_070"
	L071 = "L_071"
	L072 = "L_072"
	L073 = "L_073"
	L074 = "L_074"
)

// LoadMsg 加载动态语言
func LoadMsg() {

	en, e1 := readMsg("./message-en.json")
	if e1 != nil {
		fmt.Printf("e1:%v", e1)
		return
	}
	enStore := NewMessageStore("en")
	enStore.Set(en)

	zh, e2 := readMsg("./message-zh.json")
	if e2 != nil {
		fmt.Printf("e2:%v", e2)
		return
	}
	zhStore := NewMessageStore("zh")
	zhStore.Set(zh)

	ja, e3 := readMsg("./message-ja.json")
	if e3 != nil {
		fmt.Printf("e3:%v", e3)
		return
	}

	jaStore := NewMessageStore("ja")
	jaStore.Set(ja)
}

// GetMsg 获取消息
func GetMsg(lang string, tp MessageType, key string, params ...string) string {

	langCd := "ja"

	switch lang {
	case "zh-CN":
		langCd = "zh"
	case "en-US":
		langCd = "en"
	case "ja-JP":
		langCd = "ja"
	default:
		langCd = "ja"
	}

	store := NewMessageStore(langCd)

	var t string

	switch tp {
	case Info:
		t = "info"
	case Warn:
		t = "warn"
	case Error:
		t = "error"
	case Logger:
		t = "logger"
	default:
		t = "info"
	}

	mStr := store.Get(t, key)

	res, e := Format(mStr, params...)
	if e != nil {
		return ""
	}

	return res
}

// GetObjMsg 获取消息
func GetObjMsg(lang string, tp MessageType, key string, params map[string]string) string {

	langCd := "ja"

	switch lang {
	case "zh-CN":
		langCd = "zh"
	case "en-US":
		langCd = "en"
	case "ja-JP":
		langCd = "ja"
	default:
		langCd = "ja"
	}

	store := NewMessageStore(langCd)

	var t string

	switch tp {
	case Info:
		t = "info"
	case Warn:
		t = "warn"
	case Error:
		t = "error"
	case Logger:
		t = "logger"
	default:
		t = "info"
	}

	mStr := store.Get(t, key)

	res, e := objFormat(mStr, params)
	if e != nil {
		return ""
	}

	return res
}

func readMsg(filePath string) (Msg, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Errorf("Open file failed [Err:%s]", err.Error())
		return Msg{}, err
	}
	defer jsonFile.Close()

	var m Msg
	if err := json.NewDecoder(jsonFile).Decode(&m); err != nil {
		log.Errorf("Json decode failed [Err:%s]", err.Error())
		return Msg{}, err
	}

	return m, nil
}

// Format 格式化处理
func Format(temp string, params ...string) (string, error) {
	if len(temp) == 0 {
		log.Errorf("template string is null")
		return "", errors.New("template string is null")
	}

	for i := 0; i < len(params); i++ {
		k := fmt.Sprintf("{%v}", i)
		v := params[i]
		temp = strings.Replace(temp, k, v, 1)
	}

	return temp, nil
}

// objFormat 格式化处理
func objFormat(temp string, params map[string]string) (string, error) {
	if len(temp) == 0 {
		log.Errorf("template string is null")
		return "", errors.New("template string is null")
	}

	tpl, err := template.New("test").Parse(temp)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = tpl.Execute(&out, params)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
