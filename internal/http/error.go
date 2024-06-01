package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	er "github.com/mcorbin/corbierror"
)

func errorHandler(logger *slog.Logger) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		// can happen of ctx.Error() is called in a middleware
		// with nil passed, like for the rate limiter
		if err != nil {
			errLoggedMsg := err.Error() + " on " + c.Request().Method + " " + c.Request().URL.Path
			corbiError, ok := err.(*er.Error)
			if ok {
				if corbiError.Type == er.Forbidden {
					logger.Warn(errLoggedMsg)
				} else {
					logger.Error(errLoggedMsg)
				}
				finalErr, status := er.HTTPError(*corbiError)
				err := c.JSON(status, finalErr)
				if err != nil {
					logger.Error(err.Error())
					c.Response().Status = http.StatusInternalServerError
				}
				return
			} else {
				logger.Error(errLoggedMsg)
			}
			echoError, ok := err.(*echo.HTTPError)
			if ok {
				internal := echoError.Internal
				if internal != nil {
					jsonError, ok := internal.(*json.UnmarshalTypeError)
					if ok {
						msg := fmt.Sprintf("invalid JSON payload, field %s is incorrect", jsonError.Field)
						err := c.JSON(http.StatusBadRequest, er.Error{
							Messages: []string{msg},
						})
						if err != nil {
							logger.Error(err.Error())
							c.Response().Status = http.StatusInternalServerError
						}
						return
					}
				}
				if echoError.Code == http.StatusBadRequest && strings.Contains(echoError.Error(), "Field validation") {
					msg := strings.Split(fmt.Sprintf("%+v", echoError.Message), "\n")
					err := c.JSON(http.StatusBadRequest, er.Error{
						Messages: msg,
					})
					if err != nil {
						logger.Error(err.Error())
						c.Response().Status = http.StatusInternalServerError
					}
					return
				}
				if echoError.Code == http.StatusMethodNotAllowed {
					err := c.JSON(http.StatusMethodNotAllowed, er.Error{
						Messages: []string{"method not allowed"},
					})
					if err != nil {
						logger.Error(err.Error())
						c.Response().Status = http.StatusInternalServerError
					}
					return
				}
				if echoError.Code == http.StatusNotFound {
					err := c.JSON(http.StatusNotFound, er.Error{
						Messages: []string{"not found"},
					})
					if err != nil {
						logger.Warn(err.Error())
						c.Response().Status = http.StatusInternalServerError
					}
					return
				}
			}
			err = c.JSON(http.StatusInternalServerError, er.Error{
				Messages: []string{"internal server error"},
			})
			if err != nil {
				logger.Error(err.Error())
				c.Response().Status = http.StatusInternalServerError
			}
		}
	}
}
