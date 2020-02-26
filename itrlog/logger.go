// Package itrlog is the custom logger for Go using Zap and Lumberjack libraries.
package itrlog

import (
	"gokopy/lumberjack"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogTimeFormat formats the event timestamp.
const LogTimeFormat string = "Jan-02-2006 03:04:05 PM"

// InitLog initialize the zap and lumberjack logger library.
func InitLog(maxSize, maxAge int) *zap.Logger {
	logFile := "logs/gokopy_logs_" + time.Now().Format("01-02-2006") + ".log"

	// lumberjack.Logger is already safe for concurrent use, so we don't need to lock it.
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:  logFile,
		MaxSize:   maxSize, // megabytes
		MaxAge:    maxAge,  // days
		LocalTime: true,    // use the local machine's timezone
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		zap.InfoLevel,
	)
	return zap.New(core)
}
