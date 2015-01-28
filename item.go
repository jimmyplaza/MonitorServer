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
	//root_url := "http://130.211.243.7:3000"
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

type TopReqCountryItem struct {
	csmobj CustomerObj
	pie    [6]PieChart //for Chart.js API, generate pie chart jpg at server side
}

func (trc *TopReqCountryItem) Do() { // not ready
}
