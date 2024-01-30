package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
)

func CorrelationIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()

		id := req.Header.Get(echo.HeaderXCorrelationID)
		if id == "" {
			id = uuid.NewV4().String()
		}

		c.Response().Header().Set(echo.HeaderXCorrelationID, id)
		req = req.WithContext(context.WithValue(req.Context(), echo.HeaderXCorrelationID, id))
		c.SetRequest(req)

		return next(c)
	}
}
