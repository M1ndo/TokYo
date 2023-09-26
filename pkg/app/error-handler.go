package app

import "github.com/snwfdhmp/errlog"
import "fmt"

type ErrorHandler struct {
	Error string
}

type DebugConfig struct {
	Config *errlog.Config
	Logger errlog.Logger
}

var (
	//DefaultLoggerPrintFunc is fmt.Printf without return values
	DefaultLoggerPrintFunc = func(format string, data ...interface{}) {
		fmt.Printf(format+"\n", data...)
	}
)

func (Conf *DebugConfig) DefaultConfig() *errlog.Config {
	Conf.Config =  &errlog.Config{
			PrintFunc:          DefaultLoggerPrintFunc,
			LinesBefore:        6,
			LinesAfter:         3,
			PrintError:         true,
			PrintSource:        true,
			PrintStack:         false,
			ExitOnDebugSuccess: true,
	}
	return Conf.Config
}

func (c *DebugConfig) NewDebug() errlog.Logger {
	return errlog.NewLogger(c.Config)
}
