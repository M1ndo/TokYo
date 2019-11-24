// Created By r2dr0dn
// Date: 20/07/2019
// Don't Copy The Code Without Giving me The Credits Nerd !!
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/r2dr0dn/TokYo/pkg/app"
)

func main() {
	cfg := app.DefaultConfig()
	err := cfg.ReadFile("config.json")
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	a, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Local server: http://%s", addr)
	err = a.Run()
	if err != nil {
		log.Fatal(err)
	}
}
