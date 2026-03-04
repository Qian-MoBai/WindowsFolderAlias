package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	fileName       = "desktop.ini"
	classLabel     = "[.ShellClassInfo]"
	aliasNameLabel = "LocalizedResourceName="
)

// 判断是否是文件夹
func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// 运行命令
func runCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("出现错误: ", err)
		return ""
	}
	return string(output)
}

// 设置文件夹别名
func setFolderAlias(path, aliasName string) {
	var filePath string
	if strings.LastIndex(path, `\`) == len(path)-1 {
		filePath = path + fileName
	} else {
		filePath = path + `\` + fileName
	}
	file, err := os.ReadFile(filePath)
	content := string(file)
	if !strings.Contains(content, classLabel) {
		if content == "" {
			content = classLabel
		} else {
			content += "\n" + classLabel
		}
	}
	if !strings.Contains(content, aliasNameLabel) {
		content += "\n" + aliasNameLabel
	}
	// 分割内容为多行
	lines := strings.Split(content, "\n")
	var newLines []string
	// 过滤掉旧的 LocalizedResourceName 行
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), aliasNameLabel) {
			newLines = append(newLines, line)
		}
	}
	// 添加新的别名行
	newLines = append(newLines, aliasNameLabel+aliasName)
	content = strings.Join(newLines, "\n")
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Println("写入失败，请检查是否拥有写入权限")
		return
	}
	output := runCommand("attrib", "/S", path)
	fields := strings.Fields(output)
	if fields[1] != "SH" {
		runCommand("attrib", "+S", "+H", path)
	}
}

// 重启资源管理器
func restartExplorer() {
	runCommand("taskkill", "/F", "/IM", "explorer.exe")
	cmd := exec.Command("explorer.exe")
	err := cmd.Start()
	if err != nil {
		fmt.Println("出现错误: ", err)
		return
	}
}
