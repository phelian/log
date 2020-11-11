package log

import "context"

// Logger interface taken from https://github.com/golang/go/issues/28412
type Logger interface {
	// all levels + Prin
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Warnln(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})

	// Prefix - chainable so it can create a logger instance, safe for concurrent use
	Prefix(prefix string) Logger
	// Prefixf to avoid having to use fmt.Sprintf whenever using this
	Prefixf(fromat string, v ...interface{}) Logger
	// fields for other formatters, acts like prefix be default
	WithField(k, v interface{}) Logger

	// Context stuff as explained earlier
	Ctx(ctx context.Context) (Logger, context.Context)
	SetCtxCallback(func(ctx context.Context, logger Logger) context.Context)
}
