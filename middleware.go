package securehttp

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/peruri-dev/inalog"
)

const (
	defaultRateLimit    = 500
	defaultRateDuration = 1 * time.Second
)

func (conf *serverConfig) preMiddleware(app *fiber.App) {
	app.Use(InjectRequestID())

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e any) {
			pc, _, line, _ := runtime.Caller(3)
			inalog.LogWith(inalog.WithCfg{Ctx: c.Context()}).
				Error("PANIC",
					slog.Any("panic", e),
					slog.String("trace", fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), line)),
				)
		},
	}))

	// secure headers
	app.Use(helmet.New())

	// CORS—restrict origins in production
	if conf.config.CorsConfig.Enabled {
		origins := "*"
		methods := "OPTIONS,GET,POST"
		headers := "Content-Type,Authorization,Cookie,X-Real-IP,X-Forwarded-For"

		if conf.config.CorsConfig.Origins != "" {
			origins = conf.config.CorsConfig.Origins
		}
		if conf.config.CorsConfig.Methods != "" {
			methods = conf.config.CorsConfig.Methods
		}
		if conf.config.CorsConfig.Headers != "" {
			headers = conf.config.CorsConfig.Headers
		}

		app.Use(cors.New(cors.Config{
			AllowOrigins:     origins,
			AllowMethods:     methods,
			AllowHeaders:     headers,
			AllowCredentials: true,
		}))
	}

	// rate limiting
	if conf.config.RateLimit.Enabled {
		maxRL := defaultRateLimit
		expRL := defaultRateDuration

		if conf.config.RateLimit.Duration != 0 {
			expRL = conf.config.RateLimit.Duration
		}

		if conf.config.RateLimit.Size != 0 {
			maxRL = conf.config.RateLimit.Size
		}

		app.Use(limiter.New(limiter.Config{
			Max:                    maxRL,
			Expiration:             expRL,
			LimiterMiddleware:      limiter.SlidingWindow{},
			SkipSuccessfulRequests: true,
			LimitReached: func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusTooManyRequests)
			},
		}))
	}

	// profiling API
	if conf.config.EnableProfiling {
		app.Use(pprof.New())
	}

	// otel fiber middleware
	app.Use(otelfiber.Middleware())

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		c.Locals(inalog.CtxKeyHttp, inalog.FiberCtxHttpBuilder(c))
		c.Locals(inalog.CtxKeyDevice, inalog.FiberCtxDeviceBuilder(c))

		inalog.FiberHTTPLog(inalog.FiberHTTPLogParam{
			FiberCtx:  c,
			StartTime: start,
		})

		return c.Next()
	})
}

// request ID injection via NanoID
func InjectRequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID, err := gonanoid.New()
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		c.Locals("requestid", reqID)
		c.Set("X-Request-ID", reqID)
		return c.Next()
	}
}
