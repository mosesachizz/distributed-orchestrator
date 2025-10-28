package util

import (
	"log"
	"os"
	"strings"
	"time"
)

type Logger struct {
	*log.Logger
	level string
}

var (
	DebugLogger *Logger
	InfoLogger  *Logger
	WarnLogger  *Logger
	ErrorLogger *Logger
)

func InitLogging(level string) {
	level = strings.ToUpper(level)
	
	flags := log.Ldate | log.Ltime | log.Lshortfile
	
	DebugLogger = &Logger{
		Logger: log.New(os.Stdout, "DEBUG: ", flags),
		level:  level,
	}
	InfoLogger = &Logger{
		Logger: log.New(os.Stdout, "INFO: ", flags),
		level:  level,
	}
	WarnLogger = &Logger{
		Logger: log.New(os.Stdout, "WARN: ", flags),
		level:  level,
	}
	ErrorLogger = &Logger{
		Logger: log.New(os.Stderr, "ERROR: ", flags),
		level:  level,
	}
}

func (l *Logger) Debug(v ...interface{}) {
	if l.shouldLog("DEBUG") {
		l.Logger.Output(2, fmt.Sprint(v...))
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.shouldLog("DEBUG") {
		l.Logger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.shouldLog("INFO") {
		l.Logger.Output(2, fmt.Sprint(v...))
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.shouldLog("INFO") {
		l.Logger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if l.shouldLog("WARN") {
		l.Logger.Output(2, fmt.Sprint(v...))
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.shouldLog("WARN") {
		l.Logger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.shouldLog("ERROR") {
		l.Logger.Output(2, fmt.Sprint(v...))
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.shouldLog("ERROR") {
		l.Logger.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		"DEBUG": 4,
		"INFO":  3,
		"WARN":  2,
		"ERROR": 1,
	}
	return levels[level] <= levels[l.level]
}

func LogRequest(method, path string, status int, duration time.Duration) {
	InfoLogger.Infof("%s %s %d %v", method, path, status, duration)
}

func LogTaskEvent(taskID, event string) {
	InfoLogger.Infof("Task %s: %s", taskID, event)
}

func LogNodeEvent(nodeID, event string) {
	InfoLogger.Infof("Node %s: %s", nodeID, event)
}