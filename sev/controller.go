package sev

type ControllerImpl interface {
	Setup(*Sev)
	GetName() string
	getEndpoint() string
}

type Controller struct {
	ControllerImpl
}

func (s *Sev) RegisterController(controller ControllerImpl) {
	controller.Setup(s)
	s.Logger().Debugf("registered controller '%s'", controller.GetName())
}
