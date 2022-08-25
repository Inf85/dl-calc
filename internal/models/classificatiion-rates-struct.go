package models

type estimate struct {
	ID    string `json:"ID"`
	Value string `json:"display_value"`
}

type ClassificationRates struct {
	ID                  string `json:"ID"`
	ServiceFee          string `json:"Service_Fee"`
	Estimate            estimate
	DateOfHire          string `json:"Date_of_Hire"`
	Employees           string `json:"Employees"`
	EmployeeName        string `json:"Employee_Name"`
	TotalProfit         string `json:"Total_Profit"`
	WorkType            string `json:"Work_Type"`
	MarkUp              string `json:"Mark_Up"`
	WeeklyHours         string `json:"Weekly_Hours"`
	PTOType             string `json:"PTO_Type"`
	TotalCost           string `json:"Total_Cost"`
	TotalCustomerCharge string `json:"Total_Customer_Charge"`
}
