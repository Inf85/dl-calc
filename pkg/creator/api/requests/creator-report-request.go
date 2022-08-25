package creator

import (
	"fmt"
	creator "prevailing-calculator/pkg/creator/api/system"
	"strconv"
)

type ReportData struct {
	httpClient *creator.Single
	baseUrl    string
}

const DEFAULT_CRITERIA = "(ID != 0)"

func NewReportRequest(client *creator.Single, url string) ReportData {
	return ReportData{
		httpClient: client,
		baseUrl:    url,
	}
}

func (receiver *ReportData) GetById(recordId string, reportUrl string) ([]byte, error) {
	m := make(map[string]string)
	response, err := receiver.httpClient.Request("GET", fmt.Sprintf("%s", receiver.baseUrl+"/"+reportUrl+"/"+recordId), nil, m)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (receiver *ReportData) GetAll(reportUrl string, from int, limit int, criteria string) ([]byte, error) {
	m := make(map[string]string)
	m["from"] = strconv.Itoa(from)
	m["limit"] = strconv.Itoa(limit)
	//m["sort_order"] = "Effective_Date"
	//m["sort_by"] = "desc"
	if criteria != "" {
		m["criteria"] = criteria
	} else {
		m["criteria"] = DEFAULT_CRITERIA
	}

	response, err := receiver.httpClient.Request("GET", fmt.Sprintf("%s", receiver.baseUrl+"/"+reportUrl), nil, m)

	if err != nil {
		return nil, err
	}

	return response, nil

}
