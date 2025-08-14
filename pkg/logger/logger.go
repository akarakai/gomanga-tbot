package logger

import "go.uber.org/zap"

var Log *zap.SugaredLogger

func LoggerInit() {
	rawLogger, _ := zap.NewDevelopment()
	Log = rawLogger.Sugar()
}
