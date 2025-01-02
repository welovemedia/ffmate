package validate

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/welovemedia/ffmate/sev/exceptions"
)

const (
	URL       string = `^https?:\/\/(?:www\.)?[a-zA-Z0-9\-._~:\/?#[\]@!$&'()*+,;=.]+$`
	UUID      string = `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	NON_EMPTY string = `^.+$`
)

var val = validator.New()

type Validate struct {
}

func (s *Validate) Query(name string, regex string) gin.HandlerFunc {
	return func(gin *gin.Context) {
		r, err := regexp.Compile(regex)
		if err != nil {
			gin.AbortWithStatusJSON(500, exceptions.InternalServerError(err))
			return
		}

		q := gin.Query(name)

		if !r.MatchString(q) {
			gin.AbortWithStatusJSON(400, exceptions.HttpInvalidParam(name))
			return
		}

		gin.Next()
	}
}

func (s *Validate) Param(name string, regex string) gin.HandlerFunc {
	return func(gin *gin.Context) {
		r, err := regexp.Compile(regex)
		if err != nil {
			gin.AbortWithStatusJSON(500, exceptions.InternalServerError(err))
			return
		}

		q := gin.Param(name)

		if !r.MatchString(q) {
			gin.AbortWithStatusJSON(400, exceptions.HttpInvalidQuery(name))
			return
		}
		gin.Next()
	}
}

func (s *Validate) Bind(gin *gin.Context, v interface{}) {
	err := gin.BindJSON(v)
	if err != nil {
		e := exceptions.HttpInvalidBody(err)
		gin.AbortWithStatusJSON(e.HttpCode, e)
		return
	}

	err = val.Struct(v)
	if err != nil {
		e := exceptions.HttpInvalidBody(err)
		gin.AbortWithStatusJSON(e.HttpCode, e)
		return
	}

	gin.Next()
}
