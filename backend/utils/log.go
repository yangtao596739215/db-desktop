package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// 全局日志实例
	Logger *logrus.Logger
)

// 初始化日志配置
func init() {
	Logger = logrus.New()
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Logger.SetOutput(os.Stdout)
}

// 设置日志级别
func SetLogLevel(level logrus.Level) {
	Logger.SetLevel(level)
}

// 设置日志输出文件
func SetLogFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	Logger.SetOutput(file)
	return nil
}

// 获取调用者信息的辅助函数
func getCallerInfo() string {
	// 跳过当前函数和调用它的函数，获取真正的调用者
	pc, file, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	// 跳过 utils 包的文件
	if strings.Contains(file, "/utils/") {
		// 继续查找调用栈
		pc, file, _, ok = runtime.Caller(3)
		if !ok {
			return ""
		}
	}

	// 获取相对路径
	wd, _ := os.Getwd()
	relPath, err := filepath.Rel(wd, file)
	if err != nil {
		relPath = filepath.Base(file)
	}

	// 获取函数名
	funcName := runtime.FuncForPC(pc).Name()
	if funcName != "" {
		funcName = filepath.Base(funcName)
	}

	return fmt.Sprintf("%s %s", relPath, funcName)
}

// 格式化日志方法

// Debugf 调试级别格式化日志
func Debugf(format string, args ...interface{}) {
	caller := getCallerInfo()
	Logger.Debugf("[%s] %s", caller, fmt.Sprintf(format, args...))
}

// Infof 信息级别格式化日志
func Infof(format string, args ...interface{}) {
	caller := getCallerInfo()
	Logger.Infof("[%s] %s", caller, fmt.Sprintf(format, args...))
}

// Warnf 警告级别格式化日志
func Warnf(format string, args ...interface{}) {
	caller := getCallerInfo()
	Logger.Warnf("[%s] %s", caller, fmt.Sprintf(format, args...))
}

// Errorf 错误级别格式化日志
func Errorf(format string, args ...interface{}) {
	caller := getCallerInfo()
	Logger.Errorf("[%s] %s", caller, fmt.Sprintf(format, args...))
}

// Fatalf 致命错误级别格式化日志
func Fatalf(format string, args ...interface{}) {
	caller := getCallerInfo()
	Logger.Fatalf("[%s] %s", caller, fmt.Sprintf(format, args...))
}
