package ctx

import (
	"github.com/sedmess/go-ctx/ctx/autoctx"
	"github.com/sedmess/go-ctx/ctx/logger"
	"github.com/sedmess/go-ctx/u/nopanic"
)

type AppTask interface {
	Run()
}

func Run(task AppTask, servicePackage ...ServicePackage) {
	SetEnv(slogHandlerParam, slogHandlerLegacy)
	servicePackage = append(servicePackage, PackageOf(task))
	if err := nopanic.Run(func() {
		app := CreateContextualizedApplication(servicePackage...)
		defer app.Stop().Join()
		task.Run()
	}); err != nil {
		logger.Debug("app", err.Error())
		logger.Fatal("app", err.Reason())
	}
}

func RunAuto(task AppTask) {
	SetEnv(slogHandlerParam, slogHandlerLegacy)
	autoctx.S(task)
	if err := nopanic.Run(func() {
		app := CreateAutoContextualizedApplication()
		defer app.Stop().Join()
		task.Run()
	}); err != nil {
		logger.Debug("app", err.Error())
		logger.Fatal("app", err.Reason())
	}
}
