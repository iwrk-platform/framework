package frontend

import (
	"github.com/gofiber/fiber/v2"
	"github.com/iwrk-platform/formam/v3"
)

type BaseRequest struct {
	Id      string `json:"id,omitempty"`
	LastKey string `json:"lastKey"`
	Limit   int64  `json:"limit"`
	Offset  int64  `json:"offset"`
	Action  string `json:"action"`
}

func ParseForm[T any](ctx *fiber.Ctx, item T) *fiber.Error {
	if mpd, err := ctx.MultipartForm(); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	} else {
		decodedValues := formam.NewDecoder(&formam.DecoderOptions{TagName: "json", IgnoreUnknownKeys: true})
		if err = decodedValues.Decode(mpd.Value, item); err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
	}
	return nil
}
