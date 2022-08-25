package models

type Estimate struct {
	ID                          string `json:"ID"`
	TotalEstimateCustomerCharge string `json:"Total_Estimate_Customer_Charge"`
	NumberOfWeeks               string `json:"Number_of_Weeks"`
	TotalCost                   string `json:"Total_Cost"`
	StartDate                   string `json:"Start_Date"`
	EstimateName                string `json:"Estimate_Name"`
	TotalEstimateProfit         string `json:"Total_Estimate_Profit"`
}
