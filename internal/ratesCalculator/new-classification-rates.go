package ratesCalculator

import (
	"encoding/json"
	"fmt"
	"log"
	"prevailing-calculator/helspers"
	"prevailing-calculator/internal/models"
	creator "prevailing-calculator/pkg/creator/api"
	"strconv"
	"sync"
	"time"
)

type ApiClassificationData struct {
	Code    float64
	Data    models.ClassificationRates `json:"data"`
	Message string                     `json:"message"`
}

type ApiNewPayRates struct {
	Code float64
	Data []models.NewPayRate `json:"data"`
}

type ApiSupplementalRates struct {
	Code float64
	Data []models.SupplementalRates `json:"data"`
}

type ApiVacRate struct {
	Code float64
	Data []models.PTORates `json:"data"`
}

type ApiHalconePTORate struct {
	Code float64
	Data []models.HalconPTORates `json:"data"`
}

type ApiEstimateData struct {
	Code float64
	Data models.Estimate `json:"data"`
}

type Rates struct {
	ID                  string  `json:"classification_id"`
	WeekNumber          int32   `json:"Week_Number"`
	WeekDate            string  `json:"Week_Date"`
	LengthOfService     int32   `json:"Length_of_Service"`
	PayRate             float64 `json:"Pay_Rate"`
	MarkUpRate          float64 `json:"Mark_Up_Rate"`
	SupervisionFee      float64 `json:"Supervision_Fee"`
	PayAmount           float64 `json:"Pay_Amount"`
	SupplementalRate    float64 `json:"Supplemental_Rate"`
	SupplementalAmount  float64 `json:"Supplemental_Amount"`
	PTOAmount           float64 `json:"PTO_Amount"`
	TotalCost           float64 `json:"Total_Cost1"`
	TotalCustomerCharge float64 `json:"Total_Customer_Charge"`
	TotalProfit         float64 `json:"Total_Profit"`
	Estimate            string  `json:"Estimate"`
}

type InsertData struct {
	Employees string  `json:"Employees"`
	Rates     []Rates `json:"Rates"`
}

type ApiInsertData struct {
	Data json.RawMessage `json:"data"`
}

var apiClassificationData ApiClassificationData
var apiEstimateData ApiEstimateData
var newPayRates ApiNewPayRates
var apiSupplementalRates ApiSupplementalRates
var apiVacRate ApiVacRate
var apiHalconePTORate ApiHalconePTORate
var insertData InsertData
var apiInsertData ApiInsertData

var dateLayout = "01/02/2006"
var vacAccrued float64
var sickAccrued float64
var prevsPers float64
var vacRate string
var sickRate string
var persR float64
var totalCost float64
var totalCustomerCharge float64
var totalProfit float64

var wg sync.WaitGroup
var lock sync.Mutex

func CalculateClassificationRates(recordID string) []Rates {

	data, _ := getClassificationData(recordID)
	estimate, _ := getEstimateData(data.Estimate.ID)
	numberOfWeeks, _ := strconv.Atoi(estimate.NumberOfWeeks)
	employees, _ := strconv.ParseFloat(data.Employees, 64)
	weeklyHours, _ := strconv.ParseFloat(data.WeeklyHours, 64)
	serviceFee, _ := strconv.ParseFloat(data.ServiceFee, 64)
	markUp, _ := strconv.ParseFloat(data.MarkUp, 64)
	rates := make([]Rates, numberOfWeeks)
	wg.Add(numberOfWeeks)
	for i := 1; i <= 52; i++ {

		vacAccrued = 0.00
		sickAccrued = 0.00

		vacRate = ""
		sickRate = ""

		if i > numberOfWeeks {
			break
		}

		go func(i int) {
			rates[i-1].ID = recordID
			rates[i-1].WeekNumber = int32(i)
			startDate, _ := time.Parse(dateLayout, estimate.StartDate)
			dateOfHire, _ := time.Parse(dateLayout, data.DateOfHire)
			weekDate := startDate.AddDate(0, 0, 7*i).Format(dateLayout)
			weekDateFormatted, _ := time.Parse(dateLayout, weekDate)
			rates[i-1].WeekDate = weekDate
			_, month, _ := helspers.DatesDiff(dateOfHire, weekDateFormatted)
			rates[i-1].LengthOfService = int32(month)
			//le := int32(month)

			// Get Pay Rate
			payRate := fetchPayRate(weekDate, strconv.Itoa(month), data.WorkType, data.DateOfHire)

			if payRate == "" {
				newPayRates = ApiNewPayRates{}
				creatorApi := creator.NewCreatorApi()
				report := creatorApi.Report()
				criteriaString := "ID != 0 && From_Months <= " + strconv.Itoa(month) + " && To_Months >= " + strconv.Itoa(month) + " && Pay_Rate != null"
				criteriaString = "(" + criteriaString + ")"
				rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Pay_Rates", 0, 200, criteriaString)
				json.Unmarshal(rows, &newPayRates)

				if newPayRates.Code == 3000 {
					payRate = newPayRates.Data[len(newPayRates.Data)-1].PayRate
				}

			}

			pr, _ := strconv.ParseFloat(payRate, 64)
			rates[i-1].PayRate = pr
			rates[i-1].PayAmount = helspers.ToFixed(pr*weeklyHours, 2)

			//************************************************************************************

			// Get Sup Rate
			supRate := fetchSupRate(weekDate, strconv.Itoa(month), data.WorkType, data.DateOfHire)
			if supRate == "" {
				creatorApi := creator.NewCreatorApi()
				report := creatorApi.Report()
				apiSupplementalRates = ApiSupplementalRates{}
				criteriaString := "ID != 0 && From_Months <= " + strconv.Itoa(month) + " && To_Months >= " + strconv.Itoa(month) + " && Supplemental_Rate != null"
				criteriaString = "(" + criteriaString + ")"
				rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Supplemental_Rates", 0, 200, criteriaString)
				json.Unmarshal(rows, &apiSupplementalRates)
				if apiSupplementalRates.Code == 3000 {
					supRate = apiSupplementalRates.Data[len(apiSupplementalRates.Data)-1].SupplementalRate
				}

			}
			sr, _ := strconv.ParseFloat(supRate, 64)
			rates[i-1].SupplementalRate = sr
			perEmplHours := weeklyHours / employees
			supHours := 0.00
			if perEmplHours > 40 {
				supHours = 40 * employees
			} else {
				supHours = weeklyHours
			}

			rates[i-1].SupplementalAmount = supHours * sr

			//**************************************************************************************************

			// Standart PTO
			if data.PTOType == "Standard PTO" || data.PTOType == "" {
				vacRate = fetchPTOVACRate(weekDate, strconv.Itoa(month), data.WorkType)
				if vacRate == "" {
					creatorApi := creator.NewCreatorApi()
					report := creatorApi.Report()
					apiVacRate = ApiVacRate{}
					criteriaString := "ID != 0 && From_Months <= " + strconv.Itoa(month) + " && To_Months >= " + strconv.Itoa(month) + " && Vac_Rate != null"
					criteriaString = "(" + criteriaString + ")"
					rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_PTO_Rates", 0, 200, criteriaString)
					json.Unmarshal(rows, &apiVacRate)
					if apiVacRate.Code == 3000 {
						vacRate = apiVacRate.Data[len(apiVacRate.Data)-1].VacRate
					}

				}

				vr, _ := strconv.ParseFloat(vacRate, 64)
				vacAccrued = weeklyHours * vr

				sickRate = fetchPTOSickRate(weekDate, strconv.Itoa(month), data.WorkType)
				if sickRate == "" {
					creatorApi := creator.NewCreatorApi()
					report := creatorApi.Report()
					apiVacRate = ApiVacRate{}
					criteriaString := "ID != 0 && From_Months <= " + strconv.Itoa(month) + " && To_Months >= " + strconv.Itoa(month) + " && Sick_Rate != null"
					criteriaString = "(" + criteriaString + ")"
					rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_PTO_Rates", 0, 200, criteriaString)
					json.Unmarshal(rows, &apiVacRate)
					if apiVacRate.Code == 3000 {
						sickRate = apiVacRate.Data[len(apiVacRate.Data)-1].VacRate
					}

				}

				sickr, _ := strconv.ParseFloat(sickRate, 64)
				sickAccrued = weeklyHours * sickr
				persRate := fetchPTOPersonalRate(weekDate, strconv.Itoa(month), data.WorkType)
				persR, _ = strconv.ParseFloat(persRate, 64)
				if persRate != "" {
					prevsPers = persR
				} else {
					persR = prevsPers
				}
			} else if data.PTOType == "Halcyon PTO" {
				vacRate = fetchHalcyonPtoVACRate(weekDate, strconv.Itoa(month), data.WorkType)
				if vacRate == "" {
					creatorApi := creator.NewCreatorApi()
					report := creatorApi.Report()
					apiHalconePTORate = ApiHalconePTORate{}
					criteriaString := "ID != 0 && From_Months <= " + strconv.Itoa(month) + " && To_Months >= " + strconv.Itoa(month) + " && Vac_Rate != null"
					criteriaString = "(" + criteriaString + ")"
					rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/Halcyon_PTO_Rates_Report", 0, 200, criteriaString)
					json.Unmarshal(rows, &apiHalconePTORate)
					if apiVacRate.Code == 3000 {
						vacRate = apiHalconePTORate.Data[len(apiHalconePTORate.Data)-1].VacRate
					}
				}

				vr, _ := strconv.ParseFloat(vacRate, 64)
				vacAccrued = weeklyHours * vr

				sickRate = fetchHalcyonPtoSickRate(weekDate, strconv.Itoa(month), data.WorkType)
				if sickRate == "" {
					creatorApi := creator.NewCreatorApi()
					report := creatorApi.Report()
					apiHalconePTORate = ApiHalconePTORate{}
					criteriaString := "ID != 0 && From_Months <= " + strconv.Itoa(month) + " && To_Months >= " + strconv.Itoa(month) + " && Sick_Rate != null"
					criteriaString = "(" + criteriaString + ")"
					rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/Halcyon_PTO_Rates_Report", 0, 200, criteriaString)
					json.Unmarshal(rows, &apiHalconePTORate)
					if apiVacRate.Code == 3000 {
						sickRate = apiHalconePTORate.Data[len(apiHalconePTORate.Data)-1].SickRate
					}
				}

				sickr, _ := strconv.ParseFloat(sickRate, 64)
				sickAccrued = weeklyHours * sickr
				persRate := fetchHalcyonPtoPersonalRate(weekDate, strconv.Itoa(month), data.WorkType)
				persR, _ = strconv.ParseFloat(persRate, 64)
				if persRate != "" {
					prevsPers = persR
				} else {
					persR = prevsPers
				}
			}
			persAccrued := weeklyHours * persR
			ptoAccrued := vacAccrued + sickAccrued + persAccrued
			payRateFloat, _ := strconv.ParseFloat(payRate, 64)
			supRateFloat, _ := strconv.ParseFloat(supRate, 64)
			ptoPayAmount := payRateFloat * ptoAccrued
			ptoSupAmount := supRateFloat * ptoAccrued
			markUpRate := helspers.ToFixed(rates[i-1].PayAmount*markUp/100, 2)
			rates[i-1].PTOAmount = helspers.ToFixed(ptoPayAmount+ptoSupAmount, 2)
			rates[i-1].TotalCost = helspers.ToFixed(rates[i-1].PayAmount+rates[i-1].SupplementalAmount+rates[i-1].PTOAmount, 2)
			rates[i-1].SupervisionFee = weeklyHours * serviceFee
			rates[i-1].MarkUpRate = helspers.ToFixed(rates[i-1].PayRate*markUp/100, 2)
			rates[i-1].TotalCustomerCharge = helspers.ToFixed(rates[i-1].PayAmount+serviceFee+markUpRate+rates[i-1].SupplementalAmount+rates[i-1].PTOAmount, 2)
			rates[i-1].TotalProfit = helspers.ToFixed(rates[i-1].TotalCustomerCharge-rates[i-1].PayAmount-rates[i-1].SupplementalAmount-rates[i-1].PTOAmount, 2)
			rates[i-1].Estimate = data.Estimate.ID
			totalCost = totalCost + rates[i-1].TotalCost
			totalCustomerCharge = totalCustomerCharge + rates[i-1].TotalCustomerCharge
			totalProfit = totalProfit + rates[i-1].TotalProfit
			wg.Done()
		}(i)

	}
	wg.Wait()
	fmt.Println(rates)
	//fmt.Println(estimate)
	return rates
}

func getEstimateData(estimateID string) (models.Estimate, error) {
	apiEstimateData = ApiEstimateData{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()
	estimate, _ := report.GetById(estimateID, "pbsportal639/prevailing-utility/report/All_Calculators")

	err := json.Unmarshal(estimate, &apiEstimateData)

	if err != nil {
		return models.Estimate{}, err
	}
	fmt.Println(apiEstimateData.Data)
	return apiEstimateData.Data, nil
}

func getClassificationData(recordID string) (models.ClassificationRates, error) {
	apiClassificationData = ApiClassificationData{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()
	/*	insertData = InsertData{
			Employees: "3",
			Rates:     []Rates{{WeekNumber: 53, WeekDate: "08/08/2022", LengthOfService: 7, PayRate: 44.00, MarkUpRate: 8.14, SupervisionFee: 50.00, PayAmount: 20.00, SupplementalRate: 14.00, SupplementalAmount: 45.00, PTOAmount: 45.00, TotalCost: 55.00, TotalCustomerCharge: 44.00, TotalProfit: 65.00}},
		}

		s, err1 := json.Marshal(&insertData)
		if err1 != nil {
			log.Fatal(err1)
		}
		apiInsertData = ApiInsertData{Data: json.RawMessage(s)}
		s1, _ := json.Marshal(&apiInsertData)
		reader := bytes.NewBuffer(s1)
		resp, _ := report.UpdateById(recordID, "pbsportal639/prevailing-utility/report/All_Classification_Rates", reader)
		json.Unmarshal(resp, &apiClassificationData)
		fmt.Println(apiClassificationData)*/
	classificationRate, _ := report.GetById(recordID, "pbsportal639/prevailing-utility/report/All_Classification_Rates")

	err := json.Unmarshal(classificationRate, &apiClassificationData)

	if err != nil {
		return models.ClassificationRates{}, err
	}
	return apiClassificationData.Data, nil
}

func fetchPayRate(weekDate string, lof string, workType string, dateOfHire string) string {
	payRate := ""
	newPayRates = ApiNewPayRates{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	weekDateFormatted, _ := time.Parse(dateLayout, weekDate)
	weekDateCheck, _ := time.Parse(dateLayout, "07/01/2020")
	dateOfHireFormatted, _ := time.Parse(dateLayout, dateOfHire)
	dateOfHireCheck, _ := time.Parse(dateLayout, "01/01/2016")

	if workType == "Guard" && weekDateFormatted.Unix() < weekDateCheck.Unix() && dateOfHireFormatted.Unix() < dateOfHireCheck.Unix() {
		criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\" && Hire_Date_Criteria == \"Prior 1/1/16\""
		criteriaString = "(" + criteriaString + ")"

		rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Pay_Rates", 0, 200, criteriaString)
		err := json.Unmarshal(rows, &newPayRates)

		if err != nil {
			log.Fatal(err.Error())
		}
		if newPayRates.Code == 3000 {
			payRate = newPayRates.Data[len(newPayRates.Data)-1].PayRate
		}
	} else if workType == "Guard" && weekDateFormatted.Unix() < weekDateCheck.Unix() && dateOfHireFormatted.Unix() >= dateOfHireCheck.Unix() {
		criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\" && Hire_Date_Criteria == \"After 1/1/16\""
		criteriaString = "(" + criteriaString + ")"

		rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Pay_Rates", 0, 200, criteriaString)
		err := json.Unmarshal(rows, &newPayRates)

		if err != nil {
			log.Fatal(err.Error())
		}
		if newPayRates.Code == 3000 {
			if len(newPayRates.Data) > 0 {
				payRate = newPayRates.Data[len(newPayRates.Data)-1].PayRate
			}
		}
	} else {

		criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
		criteriaString = "(" + criteriaString + ")"
		rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Pay_Rates", 0, 200, criteriaString)
		err := json.Unmarshal(rows, &newPayRates)

		if err != nil {
			log.Fatal(err.Error())
		}
		if newPayRates.Code == 3000 {
			if len(newPayRates.Data) > 0 {
				payRate = newPayRates.Data[len(newPayRates.Data)-1].PayRate
			}
		}
	}
	return payRate
}

func fetchSupRate(weekDate string, lof string, workType string, dateOfHire string) string {
	supRate := ""
	apiSupplementalRates = ApiSupplementalRates{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	weekDateFormatted, _ := time.Parse(dateLayout, weekDate)
	weekDateCheck, _ := time.Parse(dateLayout, "07/01/2020")
	dateOfHireFormatted, _ := time.Parse(dateLayout, dateOfHire)
	dateOfHireCheck, _ := time.Parse(dateLayout, "01/01/2016")

	if workType == "Guard" && weekDateFormatted.Unix() < weekDateCheck.Unix() && dateOfHireFormatted.Unix() < dateOfHireCheck.Unix() {
		criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\" && Hire_Date_Criteria == \"Prior 1/1/16\""
		criteriaString = "(" + criteriaString + ")"

		rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Supplemental_Rates", 0, 200, criteriaString)
		err := json.Unmarshal(rows, &apiSupplementalRates)

		if err != nil {
			log.Fatal(err.Error())
		}
		if newPayRates.Code == 3000 {
			if len(apiSupplementalRates.Data) > 0 {
				supRate = apiSupplementalRates.Data[len(apiSupplementalRates.Data)-1].SupplementalRate
			}
		}
	} else if workType == "Guard" && weekDateFormatted.Unix() < weekDateCheck.Unix() && dateOfHireFormatted.Unix() >= dateOfHireCheck.Unix() {
		criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\" && Hire_Date_Criteria == \"After 1/1/16\""
		criteriaString = "(" + criteriaString + ")"

		rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Supplemental_Rates", 0, 200, criteriaString)
		err := json.Unmarshal(rows, &apiSupplementalRates)

		if err != nil {
			log.Fatal(err.Error())
		}
		if newPayRates.Code == 3000 {
			if len(apiSupplementalRates.Data) > 0 {
				supRate = apiSupplementalRates.Data[len(apiSupplementalRates.Data)-1].SupplementalRate
			}
		}
	} else {

		criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
		criteriaString = "(" + criteriaString + ")"
		rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_Supplemental_Rates", 0, 200, criteriaString)
		err := json.Unmarshal(rows, &apiSupplementalRates)

		if err != nil {
			log.Fatal(err.Error())
		}
		if newPayRates.Code == 3000 {
			if len(apiSupplementalRates.Data) > 0 {
				supRate = apiSupplementalRates.Data[len(apiSupplementalRates.Data)-1].SupplementalRate
			}
		}
	}
	return supRate
}

func fetchPTOVACRate(weekDate string, lof string, workType string) string {
	ptoVacRate := ""
	apiVacRate = ApiVacRate{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
	criteriaString = "(" + criteriaString + ")"

	rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_PTO_Rates", 0, 200, criteriaString)
	err := json.Unmarshal(rows, &apiVacRate)

	if err != nil {
		log.Fatal(err.Error())
	}
	if newPayRates.Code == 3000 {
		ptoVacRate = apiVacRate.Data[len(apiVacRate.Data)-1].VacRate
	}

	return ptoVacRate
}

func fetchPTOSickRate(weekDate string, lof string, workType string) string {
	sickRateValue := ""
	apiVacRate = ApiVacRate{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
	criteriaString = "(" + criteriaString + ")"

	rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_PTO_Rates", 0, 200, criteriaString)
	err := json.Unmarshal(rows, &apiVacRate)

	if err != nil {
		log.Fatal(err.Error())
	}
	if newPayRates.Code == 3000 {
		sickRateValue = apiVacRate.Data[len(apiVacRate.Data)-1].SickRate
	}
	return sickRateValue
}

func fetchPTOPersonalRate(weekDate string, lof string, workType string) string {
	personalRate := ""
	apiVacRate = ApiVacRate{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
	criteriaString = "(" + criteriaString + ")"

	rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/All_PTO_Rates", 0, 200, criteriaString)
	err := json.Unmarshal(rows, &apiVacRate)

	if err != nil {
		log.Fatal(err.Error())
	}
	if newPayRates.Code == 3000 {
		personalRate = apiVacRate.Data[len(apiVacRate.Data)-1].PersonalRate
	}
	return personalRate
}

func fetchHalcyonPtoVACRate(weekDate string, lof string, workType string) string {
	ptoVACRate := ""
	apiHalconePTORate = ApiHalconePTORate{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
	criteriaString = "(" + criteriaString + ")"

	rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/Halcyon_PTO_Rates_Report", 0, 200, criteriaString)
	err := json.Unmarshal(rows, &apiHalconePTORate)

	if err != nil {
		log.Fatal(err.Error())
	}
	if newPayRates.Code == 3000 {
		if len(apiHalconePTORate.Data) > 0 {
			ptoVACRate = apiHalconePTORate.Data[len(apiHalconePTORate.Data)-1].VacRate
		}
	}
	return ptoVACRate
}

func fetchHalcyonPtoSickRate(weekDate string, lof string, workType string) string {
	sickRateValue := ""
	apiHalconePTORate = ApiHalconePTORate{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
	criteriaString = "(" + criteriaString + ")"

	rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/Halcyon_PTO_Rates_Report", 0, 200, criteriaString)
	err := json.Unmarshal(rows, &apiHalconePTORate)

	if err != nil {
		log.Fatal(err.Error())
	}
	if newPayRates.Code == 3000 {
		if len(apiHalconePTORate.Data) > 0 {
			sickRateValue = apiHalconePTORate.Data[len(apiHalconePTORate.Data)-1].SickRate
		}
	}
	return sickRateValue
}

func fetchHalcyonPtoPersonalRate(weekDate string, lof string, workType string) string {
	personalRateValue := ""
	apiHalconePTORate = ApiHalconePTORate{}
	creatorApi := creator.NewCreatorApi()
	report := creatorApi.Report()

	criteriaString := "From_Months <= " + lof + " && To_Months >= " + lof + " && Effective_Date <= '" + weekDate + "' && End_Date >= '" + weekDate + "' && Work_Types ==\"" + workType + "\""
	criteriaString = "(" + criteriaString + ")"

	rows, _ := report.GetAll("pbsportal639/prevailing-utility/report/Halcyon_PTO_Rates_Report", 0, 200, criteriaString)
	err := json.Unmarshal(rows, &apiHalconePTORate)

	if err != nil {
		log.Fatal(err.Error())
	}
	if newPayRates.Code == 3000 {
		if len(apiHalconePTORate.Data) > 0 {
			personalRateValue = apiHalconePTORate.Data[len(apiHalconePTORate.Data)-1].PersonalRate
		}
	}
	return personalRateValue
}
