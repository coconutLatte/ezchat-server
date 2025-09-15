package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log   *zap.Logger
	sugar *zap.SugaredLogger
)

// Options carries logger configuration values.
type Options struct {
	Level      string
	Encoding   string
	File       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	Console    bool
}

func getZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Init initializes the global zap logger using application config.
func Init(lc Options) error {

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	var encoder zapcore.Encoder
	if strings.ToLower(lc.Encoding) == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	var cores []zapcore.Core

	// File sink with lumberjack if file path provided
	if lc.File != "" {
		lj := &lumberjack.Logger{
			Filename:   lc.File,
			MaxSize:    lc.MaxSizeMB,
			MaxBackups: lc.MaxBackups,
			MaxAge:     lc.MaxAgeDays,
			Compress:   lc.Compress,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(lj), getZapLevel(lc.Level)))
	}

	// Console stdout
	if lc.Console || lc.File == "" {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), getZapLevel(lc.Level)))
	}

	core := zapcore.NewTee(cores...)
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	sugar = log.Sugar()
	return nil
}

func L() *zap.Logger {
	return log
}

func S() *zap.SugaredLogger {
	return sugar
}
