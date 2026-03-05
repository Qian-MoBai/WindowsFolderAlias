package main

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

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

// 运行命令 (隐藏窗口)
func runCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	// 隐藏 cmd窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		log.Println("出现错误：", err)
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
	// 先删除旧文件，避免追加模式导致的问题
	os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	// 创建 UTF-16 LE 编码器
	writer := transform.NewWriter(file, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder())
	_, err = writer.Write([]byte(content))
	if err != nil {
		return err
	}
	// 确保所有数据都写入
	return writer.Close()
}

// 设置文件夹别名
func setFolderAlias(path, aliasName string) error {
	// 检查路径分隔符并拼接文件名
	var filePath string
	if strings.HasSuffix(path, `\`) || strings.HasSuffix(path, `/`) {
		filePath = path + fileName
	} else {
		if strings.Contains(path, `\`) {
			filePath = path + `\` + fileName
		} else if strings.Contains(path, "/") {
			filePath = path + `/` + fileName
		} else {
			return errors.New("请输入正确的路径")
		}
	}
	// 读取现有内容
	content, err := readUTF16LEFile(filePath)
	if err != nil {
		return errors.New("读取失败，请检查是否拥有读取权限")
	}
	// 构建新的文件内容
	var lines []string
	if content != "" {
		lines = strings.Split(content, "\n")
	}
	// 过滤掉旧的 LocalizedResourceName 行
	var newLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// 去除 BOM 字符 (U+FEFF)
		trimmedLine = strings.TrimPrefix(trimmedLine, "\ufeff")
		// 保留非空行且不包含 LocalizedResourceName 的行
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, aliasNameLabel) {
			newLines = append(newLines, trimmedLine)
		}
	}
	// 确保 [.ShellClassInfo] 标签存在且只存在一次
	hasClassLabel := false
	for _, line := range newLines {
		if line == classLabel {
			hasClassLabel = true
			break
		}
	}
	if !hasClassLabel {
		// 在开头添加标签
		newLines = append([]string{classLabel}, newLines...)
	}
	// 添加新的别名行
	newLines = append(newLines, aliasNameLabel+aliasName)
	content = strings.Join(newLines, "\n")
	// 写入文件
	err = writeUTF16LEFile(filePath, content)
	if err != nil {
		return errors.New("写入失败，请检查是否拥有写入权限")
	}
	// 设置系统隐藏属性 (不检查输出)
	runCommand("attrib", "+S", "+H", filePath)
	return nil
}

// 重启资源管理器
func restartExplorer() {
	// 静默终止资源管理器 (不检查输出)
	runCommand("taskkill", "/F", "/IM", "explorer.exe")
	// 启动新的资源管理器实例
	expCmd := exec.Command("explorer.exe")
	expCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := expCmd.Start()
	if err != nil {
		log.Println("出现错误:", err)
		return
	}
}
