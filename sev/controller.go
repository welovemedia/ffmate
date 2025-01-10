package sev

type ControllerImpl interface {
	Setup(*Sev)
	GetName() string
	getEndpoint() string
}

type Controller struct {
	ControllerImpl
}

var debugController = debug.Extend("controller")

func (s *Sev) RegisterController(controller ControllerImpl) {
	controller.Setup(s)
	debugController.Debugf("registered controller '%s'", controller.GetName())
}
