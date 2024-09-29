package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.SugaredLogger
)

func InitGlobalLogger(isDev bool) error {
	var cfg zap.Config

	if isDev {
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevel()
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level.SetLevel(zap.InfoLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	globalLogger = logger.Sugar()

	return nil
}

func GlobalLogger() *zap.SugaredLogger {
	return globalLogger
}

func GlobalSync() error {
	return globalLogger.Sync()
}

func Log(lvl zapcore.Level, args ...interface{}) {
	globalLogger.Log(lvl, args...)
}

func Logf(lvl zapcore.Level, template string, args ...interface{}) {
	globalLogger.Logf(lvl, template, args...)
}

func Logw(lvl zapcore.Level, msg string, keysAndValues ...interface{}) {
	globalLogger.Logw(lvl, msg, keysAndValues...)
}

func Info(args ...interface{}) {
	globalLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	globalLogger.Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	globalLogger.Infow(msg, keysAndValues...)
}

func Debug(args ...interface{}) {
	globalLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	globalLogger.Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	globalLogger.Debugw(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	globalLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	globalLogger.Debugf(template, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	globalLogger.Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	globalLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	globalLogger.Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	globalLogger.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	globalLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	globalLogger.Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	globalLogger.Fatalw(msg, keysAndValues...)
}
