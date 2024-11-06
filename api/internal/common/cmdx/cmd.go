package cmdx

import (
	"fmt"
	"runtime"

	execute "github.com/alexellis/go-execute/pkg/v1"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
)

func ExecCommand(command string) error {
	sysType := runtime.GOOS
	if sysType == "linux" {
		// apline系統
		params := []string{"-c", command}
		cmd := execute.ExecTask{
			Command:     "/bin/ash",
			Args:        params,
			StreamStdio: true,
		}
		//显示运行的命令
		// fmt.Println(cmd.Args)

		result, err := cmd.Execute()
		if err != nil {
			loggerx.ErrorLog("execCommand", err.Error())
			return err
		}

		if result.ExitCode != 0 {
			err := fmt.Errorf("exit code: %d, error: %s", result.ExitCode, result.Stderr)
			loggerx.ErrorLog("execCommand", err.Error())
			return err
		}

		loggerx.DebugLog("execCommand", fmt.Sprintf("stdout: %s, stderr: %s, exit-code: %d\n", result.Stdout, result.Stderr, result.ExitCode))

		return nil
	}
	if sysType == "darwin" {
		// mac系統
		params := []string{"-c", command}
		cmd := execute.ExecTask{
			Command:     "/bin/bash",
			Args:        params,
			StreamStdio: true,
		}
		//显示运行的命令
		fmt.Println(cmd.Args)

		result, err := cmd.Execute()
		if err != nil {
			loggerx.ErrorLog("execCommand", err.Error())
			return err
		}

		if result.ExitCode != 0 {
			err := fmt.Errorf("exit code: %d, error: %s", result.ExitCode, result.Stderr)
			loggerx.ErrorLog("execCommand", err.Error())
			return err
		}

		loggerx.DebugLog("execCommand", fmt.Sprintf("stdout: %s, stderr: %s, exit-code: %d\n", result.Stdout, result.Stderr, result.ExitCode))

		return nil
	}

	if sysType == "windows" {
		// windows系統
		params := []string{"/c", command}
		cmd := execute.ExecTask{
			Command:     "cmd",
			Args:        params,
			StreamStdio: true,
		}
		//显示运行的命令
		// fmt.Println(cmd.Args)

		result, err := cmd.Execute()
		if err != nil {
			loggerx.ErrorLog("execCommand", err.Error())
			return err
		}

		if result.ExitCode != 0 {
			err := fmt.Errorf("exit code: %d, error: %s", result.ExitCode, result.Stderr)
			loggerx.ErrorLog("execCommand", err.Error())
			return err
		}

		loggerx.DebugLog("execCommand", fmt.Sprintf("result: %v", result))

		return nil
	}
	return nil
}
