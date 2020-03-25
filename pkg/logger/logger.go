package logger

import (
	golog "log"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	crlog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Trace is set to true if 'trace' level logging was configured. Since Zap
// doesn't have a trace level, this uses the debug level to fake it.
var Trace bool

var (
	skipWrapper   = zap.AddCallerSkip(1)
	unskipWrapper = zap.AddCallerSkip(-1)
	nopLogger     = zap.NewNop().Sugar()
)

func newLogger(level string, options ...zap.Option) *zap.Logger {
	if level == "trace" {
		Trace = true
		level = "debug"
	}

	l := zap.DebugLevel
	l.Set(level)

	cfg := zap.NewDevelopmentConfig()
	cfg.Development = false
	cfg.Level = zap.NewAtomicLevelAt(l)
	logger, err := cfg.Build(options...)
	if err != nil {
		golog.Fatalf("cannot initialize ZAP logger: %v", err)
	}
	return logger
}

// New creates a new logger with our setup
func New(level string, options ...zap.Option) *zap.SugaredLogger {
	return newLogger(level, options...).Sugar()
}

// NewControllerLogger returns a new logger, which does not include the ctxlog wrapper in
// caller annotations and sets up controller-runtime to use it
func NewControllerLogger(level string) *zap.SugaredLogger {
	logger := newLogger(level, skipWrapper)

	crlog.SetLogger(zapr.NewLogger(logger))

	return logger.Sugar()
}

// Unskip removes the CallerSkip for the wrapper, i.e. ctxlog
func Unskip(log *zap.SugaredLogger, name string) *zap.SugaredLogger {
	return log.Named(name).Desugar().WithOptions(unskipWrapper).Sugar()
}

// TraceFilter returns the debug logger if trace is enabled, otherwise a nop logger
func TraceFilter(log *zap.SugaredLogger, name string) *zap.SugaredLogger {
	if Trace {
		return Unskip(log, name)
	}
	return nopLogger
}
