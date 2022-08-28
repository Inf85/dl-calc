package routes

import (
	"github.com/aquasecurity/lmdrouter"
	"prevailing-calculator/cmd/handlers"
)

func Routing() *lmdrouter.Router {
	router := lmdrouter.NewRouter("/prevailing-calculator")
	router.Route("GET", "/calculate", handlers.HandleRequest)
	return router
}
