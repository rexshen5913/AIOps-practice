package ioc

import (
	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/config"
	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/repository"
	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/service"
	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/web"
	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/pkg/k8s"
	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/pkg/utils"
	"go.uber.org/fx"
)

func ProvideDependencies() fx.Option {
	return fx.Options(
		fx.Provide(
			utils.NewLogger,
			config.NewConfig,
			k8s.NewK8sConfig,
			k8s.NewK8sClients,
			repository.NewAIOpsRepository,
			service.NewAIOpsService,
			// web.NewCLIHandler,
			NewCLIHandlerWithKind,
		),
		fx.Invoke(web.RunCLI),
	)
}

func NewCLIHandlerWithKind(service *service.AIOpsService, cfg *config.Config) *web.CLIHandler {
	return web.NewCLIHandler(service, cfg.Kind)
}
