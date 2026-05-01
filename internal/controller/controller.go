package controller

import (
	"dbms/internal/service"
	"dbms/pkg"
)

type Controller struct {
	Svc    *service.Service
	Config *pkg.Config
}
