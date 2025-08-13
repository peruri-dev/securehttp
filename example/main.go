package main

import (
	"fmt"
	"time"

	"github.com/peruri-dev/inalog"
	"github.com/peruri-dev/securehttp"
)

func main() {
	inalog.Init(inalog.Cfg{Source: true})
	server := securehttp.Init(securehttp.Config{
		BodyLimit:       10000,
		EnablePrefork:   false,
		ReadBufferSize:  10000,
		WriteBufferSize: 10000,
		EnableProfiling: true,
		RateLimit: securehttp.RateLimit{
			Enabled:  true,
			Duration: 1 * time.Second,
			Size:     100,
		},
		CorsConfig: securehttp.CorsConfig{
			Enabled: true,
			Methods: "OPTIONS,GET,POST",
			Origins: "*",
		},
		AppVersion:   "local",
		TimeoutRead:  5 * time.Minute,
		TimeoutWrite: 5 * time.Minute,
		DefaultPath: securehttp.DefaultPath{
			Status:  "/status",
			Health:  "/healthz",
			Ready:   "/readyz",
			Upstart: "/upstart",
			OASPath: "/oas/spec.json",
		},
	})

	// put every service that should be closed before shutdown
	server.DeferClose(func() {
		fmt.Println("this will be called before die")
	})

	server.Run(8008)
}
