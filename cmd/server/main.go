package main

import (
	"os"
)

func main() {
	// Load config (get path from env)
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.json"
	}
	// cfg, err := config.Load(cfgPath)

	// make new balancer

	// new server ...

	//
}
