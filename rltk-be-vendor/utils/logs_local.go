package utils

import (
	"io"
	"os"
	
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

var logGlobal *logrus.Logger

type FileHook struct {
	File *os.File
}

func NewFileHook(filePath string) (*FileHook, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileHook{File: file}, nil
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	_, err := hook.File.Write([]byte(entry.Message + "\n"))
	return err
}

func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func Logger(filePath string) {
	// Create an instance of the FileHook
	fileHook, err := NewFileHook(filePath)
	if err != nil {
		panic(err)
	}

	// Set up Logrus with the FileHook
	logGlobal = logrus.New()
	logGlobal.AddHook(&writer.Hook{
		Writer:    fileHook.File,
		LogLevels: logrus.AllLevels,
	})
	logGlobal.SetOutput(io.MultiWriter(os.Stdout, fileHook.File)) // Optional: log to both file and console

	// Log messages using Logrus
	logGlobal.Info("Connection established with logger service for the vendor module")
}

func GetLogger() *logrus.Logger {
	return logGlobal
}
