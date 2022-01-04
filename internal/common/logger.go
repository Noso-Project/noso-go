package common

import (
	"io"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger
var logWriter io.Writer

func InitLogger(w io.Writer) {
	// TODO: Add Production, Development, and Test loggers
	// TODO: Make custom logger that includes includes calling function in log for Dev/Test loggers
	// if logWriter == nil {
	// 	logWriter = os.Stdout
	// }
	if logger == nil {
		writeSyncer := getLogWriter()
		encoder := getEncoder()
		core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

		l := zap.New(core, zap.AddCaller())
		logger = l.Sugar()
		logger.Debug("Logger initialized")

		// config := zap.NewDevelopmentConfig()
		// config.EncoderConfig.EncodeLevel = func(_ zapcore.Level, _ zapcore.PrimitiveArrayEncoder) {}
		// config.EncoderConfig.EncodeTime = func(_ time.Time, _ zapcore.PrimitiveArrayEncoder) {}
		// // config.OutputPaths = []string{"stdout"}
		// zapLogger, _ := config.Build()
		// logger = zapLogger.Sugar()
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	// encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeLevel = func(_ zapcore.Level, _ zapcore.PrimitiveArrayEncoder) {}
	encoderConfig.EncodeTime = func(_ time.Time, _ zapcore.PrimitiveArrayEncoder) {}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	if logWriter == nil {
		logWriter = &lumberjack.Logger{
			Filename:   "./test.log",
			MaxSize:    1,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   false,
		}
	}
	return zapcore.AddSync(logWriter)
}
