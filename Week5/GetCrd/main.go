package main

import (
	"fmt"
	"runtime"

	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/ioc"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
)

func main() {

	defer Recovery()
	app := fx.New(
		ioc.ProvideDependencies(),
	)
	app.Run()
}
func Recovery() {
	if r := recover(); r != nil {
		// unknown error
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("unknown error: %v", r)
		}
		trace := make([]byte, 4096)
		runtime.Stack(trace, true)
		log.Error().Fields(map[string]interface{}{
			"stack_trace": string(trace),
		}).Msg(err.Error())
	}
}
