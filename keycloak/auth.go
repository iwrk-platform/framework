package keycloak

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"path"
	"time"
)

func (c *client) authorization() error {
	args := fiber.AcquireArgs()
	args.Set("client_id", c.clientId)
	args.Set("client_secret", c.clientSecret)
	args.Set("grant_type", "client_credentials")
	resp := &token_response{}
	status, _, err := fiber.Post(c.address + "/" + path.Join("realms", c.realm, "protocol/openid-connect/token")).
		Form(args).Struct(resp)

	if status != 200 || len(err) > 0 {
		if len(resp.ErrorDescription) > 0 || len(resp.ErrorCode) > 0 {
			err = append(err, errors.New(resp.ErrorCode), errors.New(resp.ErrorDescription))
		}
		return errors.Join(err...)
	}

	c.token = &token{
		AccessToken:  resp.AccessToken,
		TokenType:    resp.TokenType,
		Expiry:       time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
		RefreshToken: resp.RefreshToken,
		SessionState: resp.SessionState,
		ClientID:     c.clientId,
	}

	return nil
}

func (c *client) getToken() string {
	if time.Now().After(c.token.Expiry) {
		if err := c.authorization(); err != nil {
			c.logger.Error("failed authorize keycloak", zap.Error(err))
			return ""
		}
	}

	return c.token.AccessToken
}
