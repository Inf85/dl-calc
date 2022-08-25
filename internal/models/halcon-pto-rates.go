package models

type HalconPTORates struct {
	VacRate      string `json:"Vac_Rate"`
	SickRate     string `json:"Sick_Rate"`
	PersonalRate string `json:"Personal_Rate"`
}
