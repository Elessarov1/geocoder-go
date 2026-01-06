package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

type loggerKey struct{}

var (
	ctxLoggerKey = loggerKey{}
	loggingLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	programStart = time.Now()

	debugColor = color.New(color.FgHiBlack)
	infoColor  = color.New(color.FgBlue)
	warnColor  = color.New(color.FgYellow)
	errorColor = color.New(color.FgRed)
	fatalColor = color.New(color.FgHiRed)
	panicColor = color.New(color.FgHiMagenta)
	nameColor  = color.New(color.FgHiBlue)
)

func NewDevLogger() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.EncoderConfig.EncodeLevel = consoleColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = consoleDeltaEncoder()
	cfg.EncoderConfig.EncodeName = func(s string, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(nameColor.Sprint(s))
	}
	cfg.Level = loggingLevel
	cfg.OutputPaths = []string{"stdout"}

	return cfg
}

func NewProdLogger() zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.Level = loggingLevel
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Sampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
	cfg.OutputPaths = []string{"stdout"}

	return cfg
}

func NewLogger() (*zap.Logger, error) {
	var cfg zap.Config
	if term.IsTerminal(int(os.Stdout.Fd())) {
		cfg = NewDevLogger()
	} else {
		cfg = NewProdLogger()
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger, nil
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(ctxLoggerKey).(*zap.Logger); ok {
		return logger
	}
	return nil
}

func SetDebug() {
	loggingLevel.SetLevel(zap.DebugLevel)
}

func consoleColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString(debugColor.Sprint("D"))
	case zapcore.InfoLevel:
		enc.AppendString(infoColor.Sprint("I"))
	case zapcore.WarnLevel:
		enc.AppendString(warnColor.Sprint("W"))
	case zapcore.ErrorLevel:
		enc.AppendString(errorColor.Sprint("E"))
	case zapcore.FatalLevel:
		enc.AppendString(fatalColor.Sprint("F"))
	case zap.PanicLevel:
		enc.AppendString(panicColor.Sprint("P"))
	default:
		enc.AppendString("U")
	}
}

func consoleDeltaEncoder() zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		duration := t.Sub(programStart)
		seconds := duration / time.Second
		milliseconds := (duration % time.Second) / time.Millisecond
		secColor := color.New(color.Faint)
		msecColor := color.New(color.FgHiBlack)
		enc.AppendString(secColor.Sprintf("%03d", seconds) + msecColor.Sprintf(".%02d", milliseconds/10))
	}
}
