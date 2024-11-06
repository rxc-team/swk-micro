package scriptx

import "errors"

// Script datapatch处理接口
type DataPatchScript interface {
	Run() error
}

func createScript(sid string) DataPatchScript {
	var script DataPatchScript = nil
	switch sid {
	case "0929":
		script = new(Script0929)
	// case "1110":
	// 	script = new(Script1110)
	// case "1216":
	// 	script = new(Script1216)
	case "0324":
		script = new(Script0324)
	default:
		panic(errors.New("job has not found"))
	}

	return script
}
