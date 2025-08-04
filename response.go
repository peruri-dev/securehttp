package securehttp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/peruri-dev/errs"
	"github.com/peruri-dev/inalog"
)

func SuccessResponse(c *fiber.Ctx, data interface{}) error {
	response := Build{
		Data:   data,
		Errors: []any{},
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func ErrorResponse(c *fiber.Ctx, err error) error {
	log := inalog.LogWith(inalog.WithCfg{Ctx: c.Context()})

	parsed := errs.ParseCodex(err)
	resp := ErrResponse{
		ID:     c.Locals("requestid").(string),
		Status: parsed.Status,
		Code:   parsed.CustomCode,
		Title:  parsed.Title,
		Detail: parsed.Detail,
	}

	response := Build{
		Errors: []any{resp},
	}

	if parsed.Status >= 500 {
		log.Error(
			parsed.Original.Error(),
			inalog.ErrorCtx(parsed.Original),
			inalog.ErrorTrace(errs.PrintStackJson(err)),
		)
	} else {
		log.Warn(
			parsed.Original.Error(),
			inalog.ErrorCtx(parsed.Original),
			inalog.ErrorTrace(errs.PrintStackJson(err)),
		)
	}

	return c.Status(parsed.Status).JSON(response)
}
