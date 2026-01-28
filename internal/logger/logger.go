// Package logger 提供日志记录功能
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Logger 定义日志记录器结构
type Logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	successLogger *log.Logger
	file          *os.File
	writers       []io.Writer
}

// NewLogger 创建新的日志记录器
func NewLogger() (*Logger, error) {
	// 创建日志文件
	logFile, err := os.OpenFile("import.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("无法创建日志文件: %w", err)
	}

	// 创建多输出writer
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// 创建不同级别的日志记录器
	infoLogger := log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger := log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime)
	successLogger := log.New(multiWriter, "[SUCCESS] ", log.Ldate|log.Ltime)

	return &Logger{
		infoLogger:    infoLogger,
		errorLogger:   errorLogger,
		successLogger: successLogger,
		file:          logFile,
		writers:       []io.Writer{os.Stdout, logFile},
	}, nil
}

// NewLoggerWithWriter 使用自定义writer创建日志记录器（用于测试）
func NewLoggerWithWriter(writers ...io.Writer) *Logger {
	multiWriter := io.MultiWriter(writers...)

	infoLogger := log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime)
	errorLogger := log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime)
	successLogger := log.New(multiWriter, "[SUCCESS] ", log.Ldate|log.Ltime)

	return &Logger{
		infoLogger:    infoLogger,
		errorLogger:   errorLogger,
		successLogger: successLogger,
		writers:       writers,
	}
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	return l.file.Close()
}

// Info 记录信息级别的日志
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error 记录错误级别的日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Success 记录成功级别的日志
func (l *Logger) Success(format string, v ...interface{}) {
	l.successLogger.Printf(format, v...)
}

// Fatal 记录致命错误并退出程序
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
	l.Close()
	os.Exit(1)
}

// GetLogFilePath 获取日志文件的完整路径
func (l *Logger) GetLogFilePath() string {
	absPath, err := filepath.Abs(l.file.Name())
	if err != nil {
		return l.file.Name()
	}
	return absPath
}
