package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/welovemedia/ffmate/sev"
	"github.com/yosev/debugo"
)

type DebugController struct {
	sev.Controller
	sev    *sev.Sev
	Prefix string
}

func (v *DebugController) Setup(s *sev.Sev) {
	v.sev = s
	s.Gin().PATCH(v.Prefix+v.getEndpoint()+"/namespace/:namespaces", v.setDebug)
	s.Gin().DELETE(v.Prefix+v.getEndpoint()+"/namespace", v.disableDebug)
}

// @Summary Set debug namespace(s)
// @Description Set debug namespace(s)
// @Tags debug
// @Success 204
// @Router /debug/namespace/{namespace} [patch]
// @Router /debug/namespace [delete]
func (v *DebugController) setDebug(gin *gin.Context) {
	ns := gin.Param("namespaces")
	debugo.SetDebug(ns)
	if gin.Request.Method == "DELETE" {
		v.sev.Logger().Info("disabled debug logging")
	} else {
		v.sev.Logger().Infof("changed debug logging to '%s'", ns)
	}

	gin.Status(204)
}

// @Summary Turn debugging off
// @Description Turn debugging off
// @Tags debug
// @Success 204
// @Router /debug/namespace [delete]
func (v *DebugController) disableDebug(gin *gin.Context) {
	ns := gin.Param("namespaces")
	debugo.SetDebug(ns)
	if gin.Request.Method == "DELETE" {
		v.sev.Logger().Info("disabled debug logging")
	} else {
		v.sev.Logger().Infof("changed debug logging to '%s'", ns)
	}

	gin.Status(204)
}

func (v *DebugController) GetName() string {
	return "debug"
}

func (v *DebugController) getEndpoint() string {
	return "/v1/debug"
}
