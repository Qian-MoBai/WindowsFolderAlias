package main

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 文件夹路径容器
func folderPathContainer(window fyne.Window) (fyne.CanvasObject, *widget.Entry) {
	folderLabel := widget.NewLabel("需要配置别名的文件夹路径：")
	folderInput := widget.NewEntry()
	folderInput.SetPlaceHolder("请输入文件夹路径")
	folderInput.Validator = func(s string) error {
		// 校验空
		if strings.TrimSpace(s) == "" {
			return errors.New("请选择文件夹路径")
		}
		// 校验两种斜杠混合使用
		if strings.HasSuffix(s, `\`) && strings.HasSuffix(s, `/`) {
			return errors.New("两种斜杠混合")
		}
		return nil
	}
	folderButton := widget.NewButtonWithIcon("选择文件夹", theme.FolderIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				folderInput.SetText(uri.Path())
			} else if err != nil {
				dialog.ShowError(err, window)
			}
		}, window)
	})
	return container.NewBorder(nil, nil, folderLabel, folderButton, folderInput), folderInput
}

// 别名输入框
func folderAliasContainer() (*fyne.Container, *widget.Entry) {
	aliasLabel := widget.NewLabel("请输入别名：")
	aliasInput := widget.NewEntry()
	aliasInput.SetPlaceHolder("请输入别名")
	aliasInput.Validator = func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("请输入别名")
		}
		// 检查是否包含 Windows 文件名非法字符：\ / : * ? " < > |
		if strings.ContainsAny(s, "\\/:*?\"<>|") {
			return errors.New("别名不能包含以下字符：\\ / : * ? \" < > |")
		}
		return nil
	}
	return container.New(layout.NewFormLayout(), aliasLabel, aliasInput), aliasInput
}

// 操作容器
func operateContainer(window fyne.Window, folderInput *widget.Entry, aliasInput *widget.Entry) *fyne.Container {
	// 确认按钮
	confirmButton := widget.NewButton("确认", func() {
		if err := folderInput.Validate(); err != nil {
			dialog.ShowError(err, window)
			return
		}
		if err := aliasInput.Validate(); err != nil {
			dialog.ShowError(err, window)
			return
		}
		if isDir(folderInput.Text) {
			err := setFolderAlias(folderInput.Text, aliasInput.Text)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			dialog.ShowInformation("提示", "操作成功", window)
		} else {
			dialog.ShowError(errors.New("请选择文件夹路径"), window)
		}
	})
	// 取消按钮
	cancelButton := widget.NewButton("取消", func() {
		window.Close()
	})
	// 重启资源管理器按钮
	restartButton := widget.NewButton("重启资源管理器", func() {
		dialog.ShowConfirm("提示", "是否重启资源管理器？", func(b bool) {
			if b {
				restartExplorer()
			}
		}, window)
	})
	return container.NewVBox(confirmButton, restartButton, cancelButton)
}

// 启动 GUI
func startGUI() {
	a := app.NewWithID("com.mobai.windows.folder.alias")
	w := a.NewWindow("Windows 文件夹别名")
	// 加载窗口图标
	icon, err := fyne.LoadResourceFromPath("icon.png")
	if err == nil {
		w.SetIcon(icon)
	}
	w.Resize(fyne.NewSize(520, 410))
	pathContainer, folderInput := folderPathContainer(w)
	aliasContainer, aliasInput := folderAliasContainer()
	content := container.NewVBox(pathContainer, aliasContainer, operateContainer(w, folderInput, aliasInput))
	w.SetContent(content)
	w.ShowAndRun()
}
