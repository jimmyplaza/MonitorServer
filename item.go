package main

import (
	"encoding/json"
	"fmt"
)

type Item interface {
	Do(chl chan bool)
	GetJSON() ([]byte, error)
	GetChartPath() (string, error)
}

type CustomerObj struct {
	//CoAlias       string
	//SiteAliasList string
	CId    string
	SId    string
	Length string
	Kind   string
}

type PieChart struct {
	Value      float64 `json:"value"`
	Color      string  `json:"color"`
	Highlight  string  `json:"highlight"`
	Lable      string  `json:"label"`
	Percentage string  `json:"percentage"`
}

type TopThreatsCountryDataSource struct {
	CddWAFCountryRangeList [][]interface{}
}

type TopThreatsCountryItem struct {
	csmobj                      CustomerObj
	pie                         [6]PieChart //for Chart.js API, generate pie chart jpg at server side
	TopThreatsCountryDataSource             //store CW API result structure
}

/*
	Genterate Top 5 Country Pie Chart data
*/
func (ttc *TopThreatsCountryItem) Do(c chan bool) {
	//Get CW API ------------------------------------------------
	var dataSource = HTTPSDataSource{}
	CId := ttc.csmobj.CId
	SId := ttc.csmobj.SId
	length := ttc.csmobj.Length
	tmpurl := "https://g2api.nexusguard.com/API/Proxy?cust_id=%s&site_id=%s&length=%s&type=cddWAFCountryRangeList"
	url := fmt.Sprintf(tmpurl, CId, SId, length)
	chl := make(chan []byte)
	go dataSource.Get(url, chl)
	rspcontent := <-chl
	//Get CW API ------------------------------------------------
	fmt.Println(url)
	//fmt.Println(string(rspcontent))
	err := json.Unmarshal(rspcontent, &ttc.TopThreatsCountryDataSource) //store CW API result structure at ttc.TopThreatsCountryDataSource
	if err != nil {
		fmt.Println("[TopThreatsCountryItem] Do() Unmarshal Error: ", err)
		return
	}

	topCountryList := ttc.CddWAFCountryRangeList
	if len(topCountryList) == 0 {
		fmt.Println("[TopThreatsCountryItem] Do() Empty API content")
		c <- false
		return
	}
	var val_sum float64
	var rest_sum float64
	color_arr := []string{"#BED693", "#8FB751", "#4EA19B", "#26645E", "#052F33"}

	for i, val := range topCountryList {
		val_sum = val_sum + val[1].(float64)
		if i >= 5 {
			rest_sum = rest_sum + val[1].(float64)
		}
	}
	totalcountry := len(topCountryList)
	if totalcountry >= 5 {
		topCountryList = topCountryList[:5]
	}
	for i, val := range topCountryList { //Top 5 country
		ttc.pie[i].Value = val[1].(float64) //Threats Num
		ttc.pie[i].Color = color_arr[i]
		ttc.pie[i].Highlight = ""
		ttc.pie[i].Lable = val[0].(string) //Country
		tmp := (val[1].(float64) / val_sum) * 100
		p := fmt.Sprintf("%.2f", tmp)
		ttc.pie[i].Percentage = p
	}
	/*Others Country*/
	if totalcountry >= 5 {
		ttc.pie[5].Value = rest_sum
		ttc.pie[5].Color = "#928858" //Other Color
		ttc.pie[5].Highlight = ""
		ttc.pie[5].Lable = "Others"
		tmp := (rest_sum / val_sum) * 100
		ttc.pie[5].Percentage = fmt.Sprintf("%.2f", tmp)
	}
	//-------------//
	c <- true
}

func (ttc *TopThreatsCountryItem) GetJSON() ([]byte, error) {
	return json.Marshal(ttc.pie)
}

func (ttc *TopThreatsCountryItem) GetChartPath() (string, error) {
	root_url := "http://gcptools.nexusguard.com:3000"
	url := root_url + "/chart/pd/"

	var (
		err      error
		postdata []byte
		contents []byte
	)

	if postdata, err = ttc.GetJSON(); err != nil {
		return "", err
	}

	if contents, err = HttpPost(url, postdata); err != nil {
		return "", err
	}
	return root_url + string(contents), nil
}

type DCObject struct {
	CenterName  string
	CenterCount int
}

type dataCenterSource struct {
	DataCenter []DCObject
}

type DataCenterItem struct {
	csmobj CustomerObj
	pie    [6]PieChart //for Chart.js API, generate pie chart jpg at server side
	dataCenterSource
}

func (dc *DataCenterItem) Do(c chan bool) {
	//Get CW API ------------------------------------------------
	var dataSource = HTTPSDataSource{}
	CId := dc.csmobj.CId
	SId := dc.csmobj.SId
	length := dc.csmobj.Length
	tmpurl := "https://g2api.nexusguard.com/API/Proxy?cust_id=%s&site_id=%s&length=%s&type=dataCenter"
	url := fmt.Sprintf(tmpurl, CId, SId, length)
	chl := make(chan []byte)
	go dataSource.Get(url, chl)
	rspcontent := <-chl
	//Get CW API ------------------------------------------------
	err := json.Unmarshal(rspcontent, &dc.dataCenterSource) //store CW API result structure at dc.dataCenterSource
	if err != nil {
		fmt.Println("[DataCenterItem] Do() Unmarshal Error: ", err)
		return
	}
	/*
		fmt.Println(dc.dataCenterSource.DataCenter[0].CenterName)
		fmt.Println(dc.dataCenterSource.DataCenter[0].CenterCount)
		fmt.Println(dc.dataCenterSource.DataCenter[1].CenterName)
		fmt.Println(dc.dataCenterSource.DataCenter[1].CenterCount)
	*/
	var val_sum float64
	for _, val := range dc.dataCenterSource.DataCenter {
		val_sum = val_sum + float64(val.CenterCount)
	}
	color_arr := []string{"#BED693", "#8FB751", "#4EA19B", "#26645E", "#052F33", "#928858"}
	for i, val := range dc.dataCenterSource.DataCenter[:3] {
		dc.pie[i].Value = float64(val.CenterCount) //DC request
		dc.pie[i].Color = color_arr[i]
		dc.pie[i].Highlight = ""
		dc.pie[i].Lable = val.CenterName //Data Center Name
		fmt.Println(float64(val.CenterCount) / val_sum)
		tmp := (float64(val.CenterCount) / val_sum) * 100
		p := fmt.Sprintf("%.2f", tmp)
		dc.pie[i].Percentage = p
	}
	fmt.Println("dc Do() end, before c<-true")
	c <- true
	/*if len(dcList) == 0 {
		fmt.Println("[DataCenterItem] Do() Empty API content")
		c <- false
		return
	}*/

}

func (dc *DataCenterItem) GetChartPath() (string, error) {
	root_url := "http://gcptools.nexusguard.com:3000"
	url := root_url + "/chart/pd/"

	var (
		err      error
		postdata []byte
		contents []byte
	)

	fmt.Println("@0000000000")
	if postdata, err = dc.GetJSON(); err != nil {
		return "", err
	}
	fmt.Println("^1111111111111")

	if contents, err = HttpPost(url, postdata); err != nil {
		return "", err
	}
	fmt.Println("%22222222222222")
	return root_url + string(contents), nil
}

func (dc *DataCenterItem) GetJSON() ([]byte, error) {
	return json.Marshal(dc.pie)
}

/**********************/
type TopReqCountryItem struct {
	csmobj CustomerObj
	pie    [6]PieChart //for Chart.js API, generate pie chart jpg at server side
}

func (trc *TopReqCountryItem) Do() { // not ready
}
