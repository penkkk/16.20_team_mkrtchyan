package apierror

import "github.com/gin-gonic/gin"

type Response struct {
	Error Error `json:"error"`
}

type Error struct {
	Code    string   `json:"code" example:"invalid_request"`
	Message string   `json:"message" example:"Request is invalid."`
	Details []Detail `json:"details"`
}

type Detail struct {
	Field   string `json:"field,omitempty" example:"username"`
	Message string `json:"message" example:"User with this username already exists."`
}

func Respond(c *gin.Context, status int, code string, message string, details ...Detail) {
	if details == nil {
		details = []Detail{}
	}

	c.JSON(status, Response{
		Error: Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func Abort(c *gin.Context, status int, code string, message string, details ...Detail) {
	if details == nil {
		details = []Detail{}
	}

	c.AbortWithStatusJSON(status, Response{
		Error: Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
