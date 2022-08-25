package creator

import (
	"os"
	creatorrequest "prevailing-calculator/pkg/creator/api/requests"
	creator "prevailing-calculator/pkg/creator/api/system"
)

type creatorApi struct {
	baseUrl string
	request *creator.Single
}

func NewCreatorApi() *creatorApi {
	r := creator.GetInstance()
	url := os.Getenv("BASE_URL") + "/"
	return &creatorApi{
		baseUrl: url,
		request: r,
	}
}

func (ca *creatorApi) Report() creatorrequest.ReportData {
	data := creatorrequest.NewReportRequest(ca.request, ca.baseUrl)
	return data
}
