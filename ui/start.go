package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
	"hexo-tool/ui/theme"
	"log"
	"os"
	"os/exec"
)

// var serverCmd *exec.Cmd
var outputArea *widget.Entry

type MyCmd struct {
	cmd *exec.Cmd
}
type mainWindow struct {
	window          *fyne.Window
	config          *Config
	tree            *DirTree
	statusContainer *fyne.Container
}

var serverCmd *MyCmd = nil
var deployCmd *MyCmd = nil

func Run() {
	// 创建一个新的 Fyne 应用
	myApp := app.NewWithID("hexo.tools")
	//myApp.SetIcon(resourceIconPng)
	myApp.Settings().SetTheme(&theme.MyTheme{})
	myWindow := myApp.NewWindow("Hexo Tools")
	// 初始化状态条容器（默认没有状态条）
	statusContainer := container.NewVBox()
	// 文本区域，用于显示命令输出
	outputArea = widget.NewMultiLineEntry()
	outputArea.SetMinRowsVisible(10)
	//配置
	config := Config{ParentWindow: myWindow}
	configContent := config.Render()
	//目录树
	dirTree := DirTree{config: &config, Path: config.Path, ParentWindow: myWindow}
	treeContent := dirTree.Render()
	dirTree.DoRefresh(nil)
	win := mainWindow{window: &myWindow, config: &config, tree: &dirTree, statusContainer: statusContainer}
	//初始化菜单，文件目录树， 编辑区
	myWindow.SetMainMenu(win.renderMenu())

	config.DoRefresh = append(config.DoRefresh, &dirTree)

	myWindow.SetContent(container.NewBorder(
		// Top
		container.NewVBox(
			configContent,   // 主窗口内容
			statusContainer, // 动态状态条
		),

		outputArea, nil, nil,
		// Fill
		NewAdaptiveSplit(treeContent, container.NewHSplit(container.NewScroll(dirTree.EditWidget), container.NewScroll(dirTree.PreviewWidget))),
	))

	myWindow.Resize(fyne.NewSize(600, 600))
	myWindow.ShowAndRun()

}

func (window *mainWindow) renderMenu() *fyne.MainMenu {
	myWindow := window.window
	config := window.config
	dirTree := window.tree
	statusContainer := window.statusContainer
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Post File", func() {
				//cloud.ShowSettings(a, w)
				selectedPath := config.RootEntry.Path
				if config.RootEntry.Path == "" {
					dialog.ShowInformation("error", "请先选择文件夹并输入命令", *myWindow)
					return
				}
				newPost(selectedPath, dirTree, myWindow)

			}),
			fyne.NewMenuItem("Deploy", func() {
				if serverCmd != nil {

					err := killProcess(4000)
					if err != nil {
						outputArea.SetText(fmt.Sprintf("web服务终止失败: %v", err))
					} else {
						outputArea.SetText("web服务已停止")
						serverCmd = nil
						hideStatusBar(statusContainer) // 点击停止后隐藏状态条
					}
				}

				selectedPath := config.RootEntry.Path
				if config.RootEntry.Path == "" {
					dialog.ShowInformation("error", "请先选择文件夹并输入命令", *myWindow)
					return
				}
				if deployCmd != nil {
					dialog.ShowInformation("error", "服务正在部署中，请稍等", *myWindow)
					return
				} else {
					deployCmd = &MyCmd{cmd: nil}
				}

				err := runCommandInDir(selectedPath, "npx hexo clean && npx hexo g && npx hexo deploy", deployCmd)
				if err != nil {
					outputArea.SetText(fmt.Sprintf("执行命令失败: %s\n错误: %v", "npx hexo deploy", err))
					return
				}
				go func() {
					err := deployCmd.cmd.Wait()
					if err != nil {
						outputArea.SetText(fmt.Sprintf("执行命令失败: %s\n错误: %v", "npx hexo deploy", err))
						return
					}
					outputArea.SetText(fmt.Sprintf("部署成功"))
					deployCmd = nil
				}()
				//cloud.ShowSettings(a, w)
			}),
			fyne.NewMenuItem("Server", func() {
				//cloud.ShowSettings(a, w)
				selectedPath := config.RootEntry.Path
				if config.RootEntry.Path == "" {
					dialog.ShowInformation("error", "请先选择文件夹并输入命令", *myWindow)
					return
				}

				if serverCmd != nil {
					dialog.ShowInformation("error", "服务正在启动中，请勿重复启动", *myWindow)
					return
				} else {
					serverCmd = &MyCmd{cmd: nil}
				}

				err := runCommandInDir(selectedPath, "npx hexo serve", serverCmd)
				if err != nil {
					outputArea.SetText(fmt.Sprintf("执行命令失败: %s\n错误: %v", "npx hexo serve", err))
					return
				}
				showStatusBar(statusContainer, serverCmd) // 点击启动按钮时显示状态条
				//startButton.Disable()          // 禁用启动按钮以避免重复启动

			}),
			fyne.NewMenuItem("Help", func() {
				//cloud.ShowSettings(a, w)
			}),
		),
		fyne.NewMenu("Edit",
			fyne.NewMenuItem("Save   ctrl+s", func() {
				dirTree.saveFunc(true)
			}),
		),
	)

}

func watcher() {
	// 设置要监控的目录
	directory := "D:\\000\\software\\MSYS2\\tmp"
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 启动 goroutine 来监控文件变化
	go func() {
		for {
			select {
			//case event, ok := <-watcher.Events:
			//	if !ok {
			//		return
			//	}

			// 更新文件操作历史
			//operation := fmt.Sprintf("%s - %s", event.Op.String(), event.Name)
			//history.SetText(history.Text + operation + "\n")

			// 更新文件列表
			//updateFileList(fileList, directory)

			//case err, ok := <-watcher.Errors:
			//	if !ok {
			//		return
			//	}
			//history.SetText(history.Text + "ERROR: " + err.Error() + "\n")
			}
		}
	}()

	// 添加要监控的目录
	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(err)
	}
}

// 更新文件列表显示当前监控目录中的文件
func updateFileList(fileList *widget.List, directory string) {
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Println("Error reading directory:", err)
		return
	}

	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = file.Name()
	}

	fileList.Length = func() int { return len(fileNames) }
	fileList.UpdateItem = func(i int, o fyne.CanvasObject) {
		o.(*widget.Label).SetText(fileNames[i])
	}
	fileList.Refresh()
}
