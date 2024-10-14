package web

import (
	"context"
	"fmt"
	"time"

	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/service"
	"github.com/rs/zerolog/log"
)

type CLIHandler struct {
	service *service.AIOpsService
	kind    string
}

func NewCLIHandler(service *service.AIOpsService, kind string) *CLIHandler {
	return &CLIHandler{
		service: service,
		kind:    kind,
	}
}

func (h *CLIHandler) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resources, err := h.service.ListAIOpsResource(ctx, h.kind)
	if err != nil {
		log.Error().Err(err).Msg("Error listing AIops resource")
		return
	}

	for _, r := range resources {
		fmt.Printf("Name: %s, Namespace: %s, UID: %s\n", r.Name, r.Namespace, r.UID)
	}
}

func RunCLI(handler *CLIHandler) {
	handler.Run()
}
