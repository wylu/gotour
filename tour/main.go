package main

import (
	"flag"
	"log"
	"os"

	"github.com/wylu/gotour/tour/cmd"
)

func showFlagExample() {
	var name string
	flag.StringVar(&name, "name", "Go 语言编程之旅", "帮助信息")
	flag.StringVar(&name, "n", "Go 语言编程之旅", "帮助信息")
	flag.Parse()
	log.Printf("name: %v\n", name)
	log.Printf("os.Args: %v\n", os.Args)
}

func showFlagSubCommand() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return
	}

	var name string

	switch args[0] {
	case "go":
		goCmd := flag.NewFlagSet("go", flag.ExitOnError)
		goCmd.StringVar(&name, "name", "Golang", "帮助信息")
		goCmd.Parse(args[1:])
	case "py":
		pyCmd := flag.NewFlagSet("py", flag.ExitOnError)
		pyCmd.StringVar(&name, "name", "Python", "帮助信息")
		pyCmd.Parse(args[1:])
	}

	log.Printf("name: %v\n", name)
}

func main() {
	// 1.1 打开工具之旅（flag 基本使用和长短选项）
	// showFlagExample()

	// 1.1 打开工具之旅（子命令的实现）
	// showFlagSubCommand()

	// 1.2 单词格式转换
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("cmd.Execute err: %v", err)
	}
}
