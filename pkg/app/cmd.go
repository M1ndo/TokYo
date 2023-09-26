package app

import (
	"flag"
)

func ParseFlags(cfg *Config) {
	host := flag.String("host", "", "Server host")
	port := flag.Int("port", 0, "Server port")
	torpass := flag.String("torpass", "", "Tor Controller Password")
	torhost := flag.String("torhost", "", "Tor Controller Host")
	torport := flag.Int("torport", 0, "Tor Controller Port")
	flag.Parse()
	switch {
	case *host != "":
		cfg.Server.Host = *host
	case *port != 0:
		cfg.Server.Port = *port
	case *torpass != "":
		cfg.Tor.Controller.Password = *torpass
	case *torhost != "":
		cfg.Tor.Controller.Host = *torhost
	case *torport != 0:
		cfg.Tor.Controller.Port = *torport
	}
}
