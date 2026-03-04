package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
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

// 读取 UTF-16 LE 文件
func readUTF16LEFile(filename string) (string, error) {
	// 检查文件是否为空
	fileInfo, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			create, err := os.Create(filename)
			if err != nil {
				return "", err
			}
			defer create.Close()
			return "", nil
		}
		return "", err
	}
	if fileInfo.Size() == 0 {
		// 文件为空
		return "", nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// 创建 UTF-16 LE 解码器
	reader := transform.NewReader(file, unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder())
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// 写入 UTF-16 LE 文件
func writeUTF16LEFile(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	// 创建 UTF-16 LE 编码器
	writer := transform.NewWriter(file, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder())
	_, err = writer.Write([]byte(content))
	defer writer.Close()
	return err
}

// 设置文件夹别名
func setFolderAlias(path, aliasName string) error {
	// 检查路径分隔符并拼接文件名
	var filePath string
	if strings.LastIndex(path, `\`) == len(path)-1 || strings.LastIndex(path, `/`) == len(path)-1 {
		filePath = path + fileName
	} else {
		if strings.Index(path, `\`) != -1 {
			filePath = path + `\` + fileName
		} else if strings.Index(path, `/`) != -1 {
			filePath = path + `/` + fileName
		} else {
			return errors.New("请输入正确的路径")
		}
	}
	content, err := readUTF16LEFile(filePath)
	if err != nil {
		return errors.New("读取失败，请检查是否拥有读取权限")
	}
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
	err = writeUTF16LEFile(filePath, content)
	if err != nil {
		return errors.New("写入失败，请检查是否拥有写入权限")
	}
	output := runCommand("attrib", "/S", filePath)
	fields := strings.Fields(output)
	if fields[1] != "SH" {
		runCommand("attrib", "+S", "+H", filePath)
	}
	return nil
}

// 重启资源管理器
func restartExplorer() {
	runCommand("taskkill", "/F", "/IM", "explorer.exe")
	cmd := exec.Command("explorer.exe")
	err := cmd.Start()
	if err != nil {
		log.Println("出现错误: ", err)
		return
	}
}
