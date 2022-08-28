package handlers

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"prevailing-calculator/internal/ratesCalculator"
	"strings"
)

type outData struct {
	calcData []ratesCalculator.Rates `json:"calc_data"`
}

// Main Handle
func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	out := outData{}
	idsList := request.QueryStringParameters["ids"]
	if idsList != "" {
		ids := strings.Split(idsList, ",")
		for _, val := range ids {
			out.calcData = append(out.calcData, ratesCalculator.CalculateClassificationRates(val)...)
		}
	}
	//data := ratesCalculator.CalculateClassificationRates("4110017000008053215")

	e, err := json.Marshal(&out.calcData)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			MultiValueHeaders: nil,
			Body:              string(err.Error()),
			IsBase64Encoded:   false,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		MultiValueHeaders: nil,
		Body:              string(e),
		IsBase64Encoded:   false,
	}, nil
}
