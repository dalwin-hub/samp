package utils

// // package utils

// // import (
// // 	"time"

// // 	"github.com/aws/aws-sdk-go/aws"
// // 	"github.com/aws/aws-sdk-go/aws/credentials"
// // 	"github.com/aws/aws-sdk-go/aws/session"
// // 	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
// // 	"github.com/sirupsen/logrus"
// // )

// // var logGlobal *logrus.Logger

// // type CloudWatchHook struct {
// // 	GroupName  string
// // 	StreamName string
// // 	svc        *cloudwatchlogs.CloudWatchLogs
// // }

// // func NewCloudWatchHook(groupName, streamName string, svc *cloudwatchlogs.CloudWatchLogs) *CloudWatchHook {
// // 	return &CloudWatchHook{
// // 		GroupName:  groupName,
// // 		StreamName: streamName,
// // 		svc:        svc,
// // 	}
// // }

// // func (hook *CloudWatchHook) Fire(entry *logrus.Entry) error {
// // 	logEvent := &cloudwatchlogs.InputLogEvent{
// // 		Timestamp: aws.Int64(entry.Time.UnixNano() / int64(time.Millisecond)),
// // 		Message:   aws.String(entry.Message),
// // 	}

// // 	_, err := hook.svc.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
// // 		LogGroupName:  aws.String(hook.GroupName),
// // 		LogStreamName: aws.String(hook.StreamName),
// // 		LogEvents:     []*cloudwatchlogs.InputLogEvent{logEvent},
// // 	})

// // 	return err
// // }

// // func (hook *CloudWatchHook) Levels() []logrus.Level {
// // 	return logrus.AllLevels
// // }

// // func Logger(regionValue, accessKey, secretAccessKey, groupValue, streamValue string) {
// // 	// Set up an AWS session
// // 	sess := session.Must(session.NewSession(&aws.Config{
// // 		Region:      aws.String(regionValue),
// // 		Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
// // 	}))

// // 	// Create a CloudWatchLogs client
// // 	svc := cloudwatchlogs.New(sess)

// // 	// Create an instance of the CloudWatchHook
// // 	groupName := groupValue
// // 	streamName := streamValue
// // 	cloudWatchHook := NewCloudWatchHook(groupName, streamName, svc)

// // 	// Set the CloudWatchHook as the logrus hook
// // 	logGlobal = logrus.New()
// // 	logGlobal.AddHook(cloudWatchHook)

// // 	// Log messages using Logrus
// // 	logGlobal.Info("Connection established with logger service for the vendor module")

// // }

// // func GetLogger() *logrus.Logger {
// // 	return logGlobal
// // }

// package utils

// import (
// 	"io"
// 	"os"
// 	"time"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/credentials"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
// 	"github.com/sirupsen/logrus"
// 	"github.com/sirupsen/logrus/hooks/writer"
// )

// var logGlobal *logrus.Logger

// type CloudWatchHook struct {
// 	GroupName  string
// 	StreamName string
// 	svc        *cloudwatchlogs.CloudWatchLogs
// }

// type FileHook struct {
// 	File *os.File
// }

// func NewCloudWatchHook(groupName, streamName string, svc *cloudwatchlogs.CloudWatchLogs) *CloudWatchHook {
// 	return &CloudWatchHook{
// 		GroupName:  groupName,
// 		StreamName: streamName,
// 		svc:        svc,
// 	}
// }

// func (hook *CloudWatchHook) Fire(entry *logrus.Entry) error {
// 	logEvent := &cloudwatchlogs.InputLogEvent{
// 		Timestamp: aws.Int64(entry.Time.UnixNano() / int64(time.Millisecond)),
// 		Message:   aws.String(entry.Message),
// 	}

// 	_, err := hook.svc.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
// 		LogGroupName:  aws.String(hook.GroupName),
// 		LogStreamName: aws.String(hook.StreamName),
// 		LogEvents:     []*cloudwatchlogs.InputLogEvent{logEvent},
// 	})

// 	return err
// }

// func (hook *CloudWatchHook) Levels() []logrus.Level {
// 	return logrus.AllLevels
// }

// func NewFileHook(filePath string) (*FileHook, error) {
// 	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &FileHook{File: file}, nil
// }

// func (hook *FileHook) Fire(entry *logrus.Entry) error {
// 	_, err := hook.File.Write([]byte(entry.Message + "\n"))
// 	return err
// }

// func (hook *FileHook) Levels() []logrus.Level {
// 	return logrus.AllLevels
// }

// func Logger(regionValue, accessKey, secretAccessKey, groupValue, streamValue, filePath string) {
// 	// Set up an AWS session
// 	sess := session.Must(session.NewSession(&aws.Config{
// 		Region:      aws.String(regionValue),
// 		Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
// 	}))

// 	// Create a CloudWatchLogs client
// 	svc := cloudwatchlogs.New(sess)

// 	// Create an instance of the CloudWatchHook
// 	groupName := groupValue
// 	streamName := streamValue
// 	cloudWatchHook := NewCloudWatchHook(groupName, streamName, svc)

// 	// Create an instance of the FileHook
// 	fileHook, err := NewFileHook(filePath)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Set the hooks for logrus
// 	logGlobal = logrus.New()
// 	logGlobal.AddHook(cloudWatchHook)
// 	logGlobal.AddHook(&writer.Hook{ // Using writer.Hook for file logging
// 		Writer:    fileHook.File,
// 		LogLevels: logrus.AllLevels,
// 	})
// 	logGlobal.SetOutput(io.MultiWriter(os.Stdout, fileHook.File)) // Optional: log to both file and console

// 	// Log messages using Logrus
// 	logGlobal.Info("Connection established with logger service for the vendor module")
// }

// func GetLogger() *logrus.Logger {
// 	return logGlobal
// }
