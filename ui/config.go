package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type Config struct {
	Path         binding.String
	ParentWindow fyne.Window
	DoRefresh    []Refresh
	RootEntry    *Entry
}

type Refresh interface {
	DoRefresh(done chan bool)
}

func (c *Config) Render() fyne.CanvasObject {
	c.Path = binding.BindPreferenceString("path", fyne.CurrentApp().Preferences())
	folderEdit := widget.NewEntryWithData(c.Path)
	path, err := c.Path.Get()
	if err == nil {
		c.RootEntry = Prepare(path)
	}
	folderBrowseButton := widget.NewButton("Browse", func() {
		folderDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if lu != nil {
				c.Path.Set(lu.Path())
				c.RootEntry = Prepare(lu.Path())
				for i := range c.DoRefresh {
					c.DoRefresh[i].DoRefresh(nil)
				}
			}
		}, c.ParentWindow)
		currentPath, err := c.Path.Get()
		if err != nil {
			uri := storage.NewFileURI(currentPath)
			l, err := storage.ListerForURI(uri)
			if err != nil {
				folderDialog.SetLocation(l)
			}
		}
		folderDialog.Show()
	})
	return container.NewVBox(
		container.NewBorder(
			nil, nil, widget.NewLabel("Path"), folderBrowseButton, folderEdit,
		),
	)
}
