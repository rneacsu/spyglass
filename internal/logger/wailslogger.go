package logger

import "go.uber.org/zap"

type WailsLogger struct {
	logger *zap.SugaredLogger
}

func NewWailsLogger(logger *zap.SugaredLogger) *WailsLogger {
	return &WailsLogger{logger: logger.Named("wails")}
}

func (wl *WailsLogger) Print(message string) {
	wl.logger.Info(message)
}

func (wl *WailsLogger) Trace(message string) {
	wl.logger.Debug(message)
}

func (wl *WailsLogger) Debug(message string) {
	wl.logger.Debug(message)
}

func (wl *WailsLogger) Info(message string) {
	wl.logger.Info(message)
}

func (wl *WailsLogger) Warning(message string) {
	wl.logger.Warn(message)
}

func (wl *WailsLogger) Error(message string) {
	wl.logger.Error(message)
}

func (wl *WailsLogger) Fatal(message string) {
	wl.logger.Fatal(message)
}
