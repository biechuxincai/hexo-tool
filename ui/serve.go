package ui

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/sys/windows"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// runCommandInDir 执行指定文件夹中的终端命令，并返回输出
func runCommandInDir(dir string, cmdStr string, myCmd *MyCmd) error {
	// 执行终端命令

	myCmd.cmd = exec.Command("cmd.exe", "/C", cmdStr) // 适用于类 Unix 系统，Windows 上可替换为 cmd.exe
	var serverCmd = myCmd.cmd
	serverCmd.Dir = dir

	// 捕获标准输出和错误输出
	var stderr bytes.Buffer
	serverCmd.Stderr = &stderr
	stdout, err := serverCmd.StdoutPipe()
	// 执行命令
	err = serverCmd.Start()
	outputArea.SetText("服务正在启动中...")
	go printOutput(stdout)
	if err != nil {
		return fmt.Errorf(stderr.String())
	}

	return nil
}

// 打印 Web 服务的输出
func printOutput(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		output := scanner.Text()
		fmt.Println(output)
		updateServiceOutput(output)
	}
	if err := scanner.Err(); err != nil {
		updateServiceOutput(fmt.Sprintf("读取输出出错: %v", err))
	}
}

// 更新服务输出显示
func updateServiceOutput(output string) {
	// 更新服务输出，考虑线程安全
	// 创建一个从 ISO-8859-1 到 UTF-8 的转换器
	//decoder := charmap.ISO8859_1.NewDecoder()
	//data := []byte(outputArea.Text + "\n" + output)
	//// 创建一个 Reader 来进行编码转换
	//reader := transform.NewReader(bytes.NewReader(data), decoder)
	//
	//// 读取并转换数据
	//utf8Data, _ := ioutil.ReadAll(reader)
	outputArea.SetText(outputArea.Text + "\n" + output)
}

// 显示状态条
func showStatusBar(statusContainer *fyne.Container, myCmd *MyCmd) {
	var serverCmd = myCmd.cmd
	// 创建服务状态标签
	statusLabel := widget.NewLabel("服务正在进行中...")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	// 创建停止按钮
	stopButton := widget.NewButton("停止", func() {

		if serverCmd != nil && serverCmd.Process != nil {
			statusLabel.SetText("停止服务中...")
			outputArea.SetText("停止服务中...")
			// Windows 下通过 Terminate 结束进程
			//err := cmd.Process.Kill()
			err := killProcessTree(serverCmd.Process.Pid)
			if err != nil {
				fmt.Println("杀死进程失败:", err)
			} else {
				fmt.Println("进程及其子进程已杀死")
			}

			if err != nil {
				outputArea.SetText(fmt.Sprintf("发送停止信号失败: %v", err))
			} else {
				err = killProcess(4000)
				if err != nil {
					outputArea.SetText(fmt.Sprintf("服务终止失败: %v", err))
				} else {
					statusLabel.SetText("服务已停止")
					outputArea.SetText("web服务已停止")
				}
			}
			//cmd.Process.Kill()
		}
		serverCmd = nil
		hideStatusBar(statusContainer) // 点击停止后隐藏状态条
	})

	// 创建状态条并添加到容器中
	statusBar := container.NewBorder(nil, nil, nil, stopButton, statusLabel)
	statusContainer.Objects = []fyne.CanvasObject{statusBar}
	statusContainer.Refresh()
}

// 隐藏状态条
func hideStatusBar(statusContainer *fyne.Container) {
	// 清空状态条容器
	statusContainer.Objects = nil
	statusContainer.Refresh()

	//// 启用启动按钮
	//if startButton != nil {
	//	startButton.Enable()
	//}
}

// Read file content
func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		builder.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func killProcessTree(pid int) error {
	// Windows API 调用，确保所有子进程被终止
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	err = windows.TerminateProcess(handle, uint32(0))
	if err != nil {
		return err
	}

	return nil
}

func killProcess(port int) error {

	// 获取占用指定端口的进程信息
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("获取进程信息失败:", err)
		return err
	}

	// 解析 netstat 输出
	lines := strings.Split(string(output), "\n")
	var pid int
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) {
			fields := strings.Fields(line)
			if len(fields) > 4 {
				pidStr := fields[len(fields)-1]
				pid, err = strconv.Atoi(pidStr)
				if err != nil {
					fmt.Println("解析 PID 失败:", err)
					return err
				}
				break
			}
		}
	}

	if pid == 0 {
		fmt.Println("没有找到占用端口", port, "的进程")
		return errors.New("没有找到占用端口")
	}

	fmt.Println("找到占用端口", port, "的进程 PID:", pid)

	// 终止进程
	taskkillCmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	err = taskkillCmd.Run()
	if err != nil {
		fmt.Println("停止进程失败:", err)
		return err
	} else {
		fmt.Println("进程已停止")
	}
	return nil
}
