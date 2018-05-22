package server

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/graceful"
	"github.com/spf13/cobra"
	"gitlab.com/omnijar/arusha/middleware"
	"gitlab.com/omnijar/arusha/util"
)

// RunHost runs the server environment for Arusha, opening up access via HTTP/2.
func RunHost() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		router := httprouter.New()

		if err := util.InitializeClients(); err != nil {
			log.Fatalln(err.Error())
		}

		handler := &RouteHandler{}
		handler.registerRoutes(router)

		m := middleware.New(router)

		server := graceful.WithDefaults(&http.Server{
			Addr:    ":54932",
			Handler: m,
		})

		log.Println("main: Starting the server")
		if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
			log.Fatalln("main: Failed to gracefully shutdown")
		}
		log.Println("main: Server was shutdown gracefully")
	}
}
