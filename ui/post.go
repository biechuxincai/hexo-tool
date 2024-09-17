package ui

import (
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"os"
	"os/exec"
	"path/filepath"
)

func newPost(path string, tree *DirTree, w *fyne.Window) {
	// 创建输入框弹窗
	inputDialog := dialog.NewEntryDialog("新建文件", "请输入文件名", func(fileName string) {
		// 点击确认后的逻辑：创建文件
		err := createFile(path, fileName)
		if err != nil {
			dialog.ShowError(err, *w)
		} else {
			dialog.ShowInformation("文件创建成功", fmt.Sprintf("文件 '%s' 已创建", fileName), *w)
			fullPath := filepath.Join(path, "source", "_posts", fileName+".md")

			// 更新目录树并展开新文件所在的目录
			//tree.config.Path.Set(path)
			tree.OpenBranch(fullPath)
			//tree.config.RootEntry = Prepare(path)
			//done := make(chan bool)
			//tree.DoRefresh(done)
			//<-done
			tree.FileTree.Select(filepath.ToSlash(fullPath)) // 在树中选中新建的文件
			//
			//tree.FileTree.OpenBranch(filepath.ToSlash(filepath.Join(path)))                     // 自动展开新文件的父目录
			//tree.FileTree.OpenBranch(filepath.ToSlash(filepath.Join(path, "source")))           // 自动展开新文件的父目录
			//tree.FileTree.OpenBranch(filepath.ToSlash(filepath.Join(path, "source", "_posts"))) // 自动展开新文件的父目录

		}
	}, *w)
	inputDialog.Resize(fyne.NewSize(300, 100))
	// 设置输入框提示
	inputDialog.SetPlaceholder("请输入文件名")
	inputDialog.Show()

}

// 创建文件的函数
func createFile(path string, fileName string) error {

	// 检查文件名是否为空
	if fileName == "" {
		return fmt.Errorf("文件名不能为空")
	}
	_, err := os.Stat(filepath.Join(path, "source", "_posts", fileName+".md"))
	if !os.IsNotExist(err) {
		return fmt.Errorf("文件已存在")
	}

	cmd := exec.Command("cmd.exe", "/C", "hexo new post "+fileName) // 适用于类 Unix 系统，Windows 上可替换为 cmd.exe
	cmd.Dir = path

	// 捕获标准输出和错误输出
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	// 执行命令
	err = cmd.Run()
	outputArea.SetText(stdout.String())
	if err != nil {
		return fmt.Errorf(stderr.String())
	}

	outputArea.SetText("创建文件成功：" + fileName)
	return nil
}
