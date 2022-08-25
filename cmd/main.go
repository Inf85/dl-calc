package main

import (
	"github.com/joho/godotenv"
	"log"
	"prevailing-calculator/internal/ratesCalculator"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("ERROR")
	}
	ratesCalculator.CalculateClassificationRates("4110017000007855072")
}
