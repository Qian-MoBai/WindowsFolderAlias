package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	var path string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("输入需要别名的文件夹路径: ")
	if scanner.Scan() {
		path = scanner.Text()
	} else if isDir(path) == false {
		fmt.Println("请输入正确的文件夹路径")
		return
	}
	var aliasName string
	fmt.Print("输入别名: ")
	if scanner.Scan() {
		aliasName = scanner.Text()
	} else {
		fmt.Println("请输入正确的别名")
		return
	}
	setFolderAlias(path, aliasName)
	fmt.Println("设置成功，重启资源管理器生效")
	restartExplorer()
}
