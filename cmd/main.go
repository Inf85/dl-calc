package main

import (
	"context"
	"github.com/aquasecurity/lmdrouter"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"github.com/joho/godotenv"
	"log"
	"prevailing-calculator/cmd/routes"
)

var router *lmdrouter.Router
var ctx = context.Background()
var secretCache, _ = secretcache.New()
var dateLayout = "01/02/2006"

// Init
func init() {
	router = routes.Routing()
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("ERROR")
	}

	lambda.Start(router.Handler)
	//ratesCalculator.CalculateClassificationRates("4110017000008053215")

}
