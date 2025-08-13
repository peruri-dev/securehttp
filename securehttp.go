package securehttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/peruri-dev/inalog"
)

type DefaultPath struct {
	Status  string
	Health  string
	Ready   string
	Upstart string
	OASPath string
}

type RateLimit struct {
	Enabled  bool
	Duration time.Duration
	Size     int
}

type CorsConfig struct {
	Enabled bool
	Methods string
	Origins string
}

type Config struct {
	BodyLimit       int
	EnablePrefork   bool
	ReadBufferSize  int
	WriteBufferSize int
	EnableProfiling bool
	RateLimit       RateLimit
	CorsConfig      CorsConfig
	AppVersion      string
	TimeoutRead     time.Duration
	TimeoutWrite    time.Duration
	DefaultPath     DefaultPath
}

type IServer interface {
	Run(int)
	App() *fiber.App
	DeferClose(func())
}

type Apps struct{}

type serverConfig struct {
	app    *fiber.App
	config Config
	close  func()
}

func Init(c Config) *serverConfig {
	fiberConf := fiber.Config{
		StrictRouting: true,
		ServerHeader:  "Peruri-SecureHTTP",
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
	}

	if c.TimeoutRead != 0 {
		fiberConf.ReadTimeout = c.TimeoutRead
	}

	if c.TimeoutWrite != 0 {
		fiberConf.WriteTimeout = c.TimeoutWrite
	}

	if c.BodyLimit == 0 {
		fiberConf.BodyLimit = 1024 * 1024
	}

	if c.EnablePrefork {
		fiberConf.Prefork = true
	}

	fiberApp := fiber.New(fiberConf)
	newServer := &serverConfig{
		app:    fiberApp,
		config: c,
	}

	newServer.preMiddleware(fiberApp)

	return newServer
}

func (c *serverConfig) DeferClose(fn func()) {
	c.close = fn
}

func (c *serverConfig) setDefaultPath() {
	app := c.app

	statusPath := "/status"
	if c.config.DefaultPath.Status != "" {
		statusPath = c.config.DefaultPath.Status
	}

	appVersion := c.config.AppVersion
	app.Get(statusPath, func(c *fiber.Ctx) error {
		return c.Status(200).JSON(map[string]string{
			"AppVersion": appVersion,
		})
	})

	healthPath := "/healthz"
	if c.config.DefaultPath.Status != "" {
		healthPath = c.config.DefaultPath.Health
	}

	// liveness probe
	app.Get(healthPath, func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	readyPath := "/readyz"
	if c.config.DefaultPath.Status != "" {
		readyPath = c.config.DefaultPath.Ready
	}

	// readiness probe
	app.Get(readyPath, func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	upstartPath := "/upstart"
	if c.config.DefaultPath.Status != "" {
		upstartPath = c.config.DefaultPath.Upstart
	}

	// startup probe
	app.Get(upstartPath, func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
}

func (c *serverConfig) SetOAS(jsonFile []byte) {
	oasPath := "/oas/spec.json"
	if c.config.DefaultPath.OASPath != "" {
		oasPath = c.config.DefaultPath.OASPath
	}

	c.app.Get(oasPath, func(c *fiber.Ctx) error {
		c.Type("json")
		return c.Send(jsonFile)
	})
}

func (c *serverConfig) Run(port int) {
	log := inalog.Log()
	// Channel to capture serve errors
	errCh := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		c.setDefaultPath()

		// override default 404 page
		c.app.Use(func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusNotFound).JSON(Build{
				Errors: []any{ErrResponse{
					ID:     c.Locals("requestid").(string),
					Status: 404,
					Code:   "HTTP-404",
					Title:  "not found",
					Detail: "the page you are looking for is not exist",
				}},
			})
		})

		// ListenAndServe will block until error or shutdown
		err := c.app.Listen(fmt.Sprintf(":%d", port))
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server listen error: %w", err)
		}
		close(errCh)
	}()

	// Listen for OS interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case err := <-errCh:
		// Server encountered a fatal error
		log.Fatal(fmt.Sprintf("Server error: %v", err))

	case sig := <-sigCh:
		// Received interrupt, initiate graceful shutdown
		log.Notice(fmt.Sprintf("Received signal %v, shutting down...", sig))

		// Create context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		c.close()

		// Shutdown server
		if err := c.app.ShutdownWithContext(ctx); err != nil {
			log.Fatal(fmt.Sprintf("Graceful shutdown failed: %v", err))
		}
		log.Notice("Server stopped cleanly")
	}
}

func (c *serverConfig) App() *fiber.App {
	return c.app
}
