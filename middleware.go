package securehttp

import (
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
	app.Use(recover.New())

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		c.Locals(inalog.CtxKeyHttp, inalog.FiberCtxHttpBuilder(c))
		c.Locals(inalog.CtxKeyDevice, inalog.FiberCtxDeviceBuilder(c))

		err := c.Next()
		inalog.FiberHTTPLog(inalog.FiberHTTPLogParam{
			FiberCtx:  c,
			StartTime: start,
		})

		return err
	})

	// secure headers
	app.Use(helmet.New())

	// CORS—restrict origins in production
	if conf.config.CorsConfig.Enabled {
		origins := "*"
		methods := "OPTIONS,GET,POST"

		if conf.config.CorsConfig.Origins != "" {
			origins = conf.config.CorsConfig.Origins
		}
		if conf.config.CorsConfig.Methods != "" {
			methods = conf.config.CorsConfig.Methods
		}

		app.Use(cors.New(cors.Config{
			AllowOrigins: origins,
			AllowMethods: methods,
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

	app.Use(InjectRequestID())

	// profiling API
	if conf.config.EnableProfiling {
		app.Use(pprof.New())
	}

	// otel fiber middleware
	app.Use(otelfiber.Middleware())
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
