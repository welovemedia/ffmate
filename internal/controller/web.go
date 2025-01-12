package controller

import (
	"embed"
	"io"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/sev"
)

type WebController struct {
	sev.Controller
	sev *sev.Sev

	Frontend embed.FS
	Prefix   string
}

func (c *WebController) Setup(s *sev.Sev) {
	c.sev = s
	s.Gin().StaticFS(c.getEndpoint(), mustFS(c.Frontend))

	c.sev.Gin().NoRoute(func(gin *gin.Context) {
		f, _ := mustFS(c.Frontend).Open("index.html")
		b, _ := io.ReadAll(f)
		gin.Writer.Write(b)
	})
	s.Gin().GET("/", func(gin *gin.Context) {
		gin.Redirect(http.StatusMovedPermanently, c.getEndpoint())
	})
}

func mustFS(frontend embed.FS) http.FileSystem {
	sub, err := fs.Sub(frontend, "ui/.output/public")

	if err != nil {
		panic(err)
	}

	return http.FS(sub)
}

func (c *WebController) GetName() string {
	return "web"
}

func (c *WebController) getEndpoint() string {
	return "/ui"
}
