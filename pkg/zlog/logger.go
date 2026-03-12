package zlog

import (
    "os"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.Logger

func init() {
    InitLogger("logs/im.log", zapcore.DebugLevel)
}

func InitLogger(logPath string, level zapcore.Level) {
    customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
        enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
    }

    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "time",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeTime:     customTimeEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }

    consoleConfig := encoderConfig
    consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)

    fileConfig := encoderConfig
    fileConfig.EncodeLevel = zapcore.CapitalLevelEncoder
    fileEncoder := zapcore.NewJSONEncoder(fileConfig)

    hook := &lumberjack.Logger{
        Filename:   logPath,
        MaxSize:    100,
        MaxBackups: 30,
        MaxAge:     7,
        Compress:   true,
    }

    // write to console and log files
    core := zapcore.NewTee(
        zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
        zapcore.NewCore(fileEncoder, zapcore.AddSync(hook), level),
    )

    // use AddCallerSkip to skip zlog pkg
    log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func Debug(msg string, fields ...zap.Field) {
    log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
    log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
    log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
    log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
    log.Fatal(msg, fields...)
}

func Sync() {
    _ = log.Sync()
}
