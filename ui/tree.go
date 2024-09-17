package ui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DirTree struct {
	config        *Config
	Path          binding.String
	ParentWindow  fyne.Window
	FileTree      *widget.Tree
	CurrentFile   fyne.URI
	EditWidget    *widget.Entry
	PreviewWidget *widget.RichText
	selectedNode  string
}

type myCanvas struct {
	*widget.Label
	id           string
	onRightClick func(uid widget.TreeNodeID, e *fyne.PointEvent)
	onLeftClick  func(uid widget.TreeNodeID)
}

func (canvas *myCanvas) Tapped(e *fyne.PointEvent) {
	fmt.Printf("%s", e)
	canvas.onLeftClick(canvas.id)
}

// 右键处理
func (canvas *myCanvas) TappedSecondary(e *fyne.PointEvent) {
	fmt.Printf("%s", canvas.id)
	canvas.onRightClick(canvas.id, e)
}

//
//	func (t *DirTree) MouseDown(m *desktop.MouseEvent) {
//		if m.Button == desktop.MouseButtonSecondary {
//			t.onRightClick(t.selectedNode)
//		}
//	}
//
// func (e *DirTree) MouseUp(m *desktop.MouseEvent) {
//
// }

func (t *DirTree) TappedSecondary(e *fyne.PointEvent) {
	fmt.Printf("%s", e)
	fmt.Printf("%s", t.selectedNode)
}
func (tree *DirTree) DoRefresh(done chan bool) {
	var rootEntry *Entry
	var ctx context.Context
	var cancel context.CancelFunc
	// Prepare the context
	ctx, cancel = context.WithCancel(context.Background())

	// Prepare the root entry
	rootEntry = tree.config.RootEntry
	// Get the exclusions
	exclusions := fyne.CurrentApp().Preferences().StringListWithFallback("exclusions", []string{})

	// Build the settings object
	settings := ProcessSettings{
		Context:      ctx,
		CurrentEntry: rootEntry,
		Exclusions:   exclusions,
	}

	// Start a goroutine which builds the tree in another subroutine and regularly refreshes the tree
	go func() {
		go func() {
			BuildTreeRecursive(settings)
			defer cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				tree.FileTree.Refresh()
				if done != nil {
					done <- true
				}
				return
			case <-time.After(5 * time.Second):
				tree.FileTree.Refresh()
			}
		}
	}()

	//tree.FileTree.Refresh()
}
func (tree *DirTree) OpenBranch(newPath string) {
	path := tree.config.RootEntry.Path
	// 更新目录树并展开新文件所在的目录
	tree.config.RootEntry = Prepare(path)
	done := make(chan bool)
	tree.DoRefresh(done)
	<-done
	tree.FileTree.Select(filepath.ToSlash(newPath)) // 在树中选中新建的文件
	_, afterS, found := strings.Cut(filepath.ToSlash(newPath), path)
	if found {
		arr := strings.Split(afterS, "/")
		subPath := ""
		for i := 0; i < len(arr); i++ {
			//println(subPath + arr[i])
			subPath = filepath.Join(subPath, arr[i])
			tree.FileTree.OpenBranch(filepath.ToSlash(filepath.Join(path, subPath)))
		}
	}
}
func (tree *DirTree) Render() fyne.CanvasObject {
	// Folder selection
	showContextMenu := func(selectedNode widget.TreeNodeID, e *fyne.PointEvent) {
		if selectedNode == "" {
			return
		}
		contextMenu := fyne.NewMenu("",
			fyne.NewMenuItem("删除", func() {
				dialog.ShowConfirm("确认删除", "确定要删除这个文件/目录吗?", func(ok bool) {
					if ok {
						err := os.RemoveAll(selectedNode)
						if err != nil {
							dialog.ShowError(err, tree.ParentWindow)
						} else {
							tree.OpenBranch(selectedNode)
						}
					}
				}, tree.ParentWindow)
			}),
			fyne.NewMenuItem("重命名", func() {
				dialog.ShowEntryDialog("重命名", "请输入新名称:", func(newName string) {
					if newName != "" {
						oldPath := selectedNode
						suffixName := filepath.Ext(oldPath)
						newPath := filepath.Join(filepath.Dir(oldPath), newName+suffixName)
						err := os.Rename(oldPath, newPath)
						if err != nil {
							dialog.ShowError(err, tree.ParentWindow)
						} else {
							tree.OpenBranch(newPath)
						}
					}
				}, tree.ParentWindow)
			}),
		)
		// 显示上下文菜单
		widget.ShowPopUpMenuAtPosition(contextMenu, tree.ParentWindow.Canvas(), e.AbsolutePosition)
	}
	leftChick := func(selectedNode widget.TreeNodeID) {
		//tree.selectedNode = selectedNode
		tree.FileTree.Select(selectedNode)
	}

	if tree.Path == nil {
		return nil
	}

	sortBySetting := SortByName
	//todo  无能为力 sb框架
	tree.FileTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			rootEntry := tree.config.RootEntry
			if id == "" {
				return []widget.TreeNodeID{rootEntry.Path}
			}
			ids := []widget.TreeNodeID{}
			currEntry := getEntryFromTreeId(rootEntry, id)

			// Sorting
			entries := []*Entry{}

			// Separately sort folders, then files, then merge them
			dirEntries := []*Entry{}
			dirEntries = append(dirEntries, currEntry.Folders...)
			SortEntries(sortBySetting, dirEntries)

			fileEntries := []*Entry{}
			fileEntries = append(fileEntries, currEntry.Files...)
			SortEntries(sortBySetting, fileEntries)

			entries = append(entries, dirEntries...)
			entries = append(entries, fileEntries...)

			// Add all the ids
			for _, entry := range entries {
				ids = append(ids, entry.Path)
			}
			return ids
		},
		func(id widget.TreeNodeID) bool {
			rootEntry := tree.config.RootEntry
			currEntry := getEntryFromTreeId(rootEntry, id)
			return currEntry.IsFolder
		},
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewFileIcon(storage.NewFileURI("/"))

			return container.NewBorder(
				nil,
				nil,
				container.NewHBox(icon, &myCanvas{widget.NewLabel(""), ``, nil, nil}),
				nil,
			)
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {

			rootEntry := tree.config.RootEntry
			currEntry := getEntryFromTreeId(rootEntry, id)
			text := currEntry.Name
			if branch {
				text += fmt.Sprintf(" : %d / %d", len(currEntry.Folders), len(currEntry.Files))
			}

			rootContainer := o.(*fyne.Container)
			leftContainer := rootContainer.Objects[0].(*fyne.Container)
			canvas := leftContainer.Objects[1].(*myCanvas)
			canvas.SetText(text)
			canvas.id = id
			// 创建上下文菜单
			canvas.onRightClick = showContextMenu
			canvas.onLeftClick = leftChick
			//canvas.icon.SetURI(storage.NewFileURI(currEntry.Path))
			fileIcon := leftContainer.Objects[0].(*widget.FileIcon)
			//nameLabel := leftContainer.Objects[1].(*widget.Label)
			fileIcon.SetURI(storage.NewFileURI(currEntry.Path))
			//nameLabel.SetText(text)

		})
	//编辑区域
	tree.EditWidget = widget.NewMultiLineEntry()
	tree.EditWidget.Wrapping = fyne.TextWrapWord
	//selectedNode := tree.selectedNode

	tree.PreviewWidget = widget.NewRichTextFromMarkdown("")
	tree.EditWidget.OnChanged = tree.PreviewWidget.ParseMarkdown
	tree.FileTree.OnSelected = func(id widget.TreeNodeID) {
		rootEntry := tree.config.RootEntry
		//先保存
		tree.saveFunc(false)
		currEntry := getEntryFromTreeId(rootEntry, id)
		tree.CurrentFile = storage.NewFileURI(currEntry.Path)
		if !currEntry.IsFolder {
			//selectedNode = id
			path := currEntry.Path
			if !strings.HasSuffix(strings.ToLower(path), ".md") {
				tree.EditWidget.SetText("please use the .md extension")
				return
			}
			content, err := readFile(id)
			if err != nil {
				tree.EditWidget.SetText("无法读取文件内容: " + err.Error())
				return
			}
			tree.PreviewWidget.ParseMarkdown(content)
			tree.EditWidget.SetText(content)
		} else {
			tree.EditWidget.SetText("请选择一个文件查看内容")
		}
	}
	// 监听快捷键
	tree.ParentWindow.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}, func(shortcut fyne.Shortcut) {
		tree.saveFunc(true)
	})

	return tree.FileTree

}

func (tree *DirTree) saveFunc(warn bool) {

	if tree.CurrentFile != nil {
		info, err := os.Stat(tree.CurrentFile.Path())
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if info.IsDir() {
			return
		}
		write, err := storage.Writer(tree.CurrentFile)
		if err != nil {
			dialog.ShowError(err, tree.ParentWindow)
			return
		}

		write.Write([]byte(tree.EditWidget.Text))
		defer write.Close()
		if warn {
			dialog.ShowInformation("Saved", "保存成功", tree.ParentWindow)
		}
	}
}
func getEntryFromTreeId(rootEntry *Entry, path string) *Entry {
	if rootEntry == nil {
		return &Entry{Path: "LOADING", Name: "LOADING"}
	}
	if path == rootEntry.Path {
		return rootEntry
	}
	return rootEntry.GetChildFromPath(path)
}
