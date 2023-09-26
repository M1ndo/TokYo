// Created By ybenel
// Date: 20/07/2019 || Updated in 2023/02/09
package main

import (
	"fmt"
	app "github.com/M1ndo/TokYo/pkg/app"
	"os"
)

func main() {
	cfg := app.DefaultConfig()
	err := cfg.ReadFile("config.json")
	if err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
	}
	app.ParseFlags(cfg)
	a, err := app.NewApp(cfg)
	logger := a.GetLogger()
	if err != nil {
		logger.Log.Fatal(err)
		a.Debug.Logger.Debug(err)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Log.Infof("Local server: http://%s", addr)
	err = a.Run()
	if err != nil {
		logger.Log.Fatal(err)
		a.Debug.Logger.Debug(err)
	}
}
