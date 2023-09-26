// Date 2023/09/11
// Created By ybenel
// Logger for the whole app.
package log

import (
	glg "github.com/kpango/glg"
)

type Logger struct {
	Log *glg.Glg
}

func (log *Logger) NewLogger() (*Logger, error) {
	infoWriter := glg.FileWriter("logs/log_server.log", 0644)
	errWriter := glg.FileWriter("logs/error_server.log", 0644)
	// defer infoWriter.Close()
	// defer errWriter.Close()
	log.Log = glg.Get().
		SetMode(glg.BOTH).
		AddLevelWriter(glg.INFO, infoWriter).
		AddLevelWriter(glg.WARN, infoWriter).
		AddLevelWriter(glg.LOG, infoWriter).
		AddLevelWriter(glg.PRINT, infoWriter).
		AddLevelWriter(glg.OK, infoWriter).
		AddLevelWriter(glg.DEBG, infoWriter).
		AddLevelWriter(glg.FATAL, errWriter).
		AddLevelWriter(glg.FAIL, errWriter).
		AddLevelWriter(glg.ERR, errWriter)
	return log, nil
}
