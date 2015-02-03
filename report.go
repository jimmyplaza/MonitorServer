package main

import (
	"fmt"
	//"runtime"
)

type Report struct {
	items []Item
}

func (r *Report) registerItem(i Item) {
	r.items = append(r.items, i)
}

func (r *Report) start() (arr []string) {
	var it Item
	for i, l := 0, len(r.items); i < l; i++ {
		it = r.items[i]
		chl := make(chan bool)
		go it.Do(chl)
		result := <-chl
		fmt.Println(result)
		if result {
			if data, err := it.GetChartPath(); err != nil {
				fmt.Println("JJJJJJJJJJJJJIMMMMMMMMM")
				fmt.Println(err)
			} else {
				//fmt.Println(string(data))
				arr = append(arr, string(data))
			}
		}
	}
	return arr
}

/*
func main() {
	cpunum := runtime.NumCPU()
	runtime.GOMAXPROCS(cpunum)

	report := Report{}
	var topThreatsCountry TopThreatsCountryItem
	var dcItem DataCenterItem
	topThreatsCountry.csmobj.CId = "C-a4c0f8fd-ccc9-4dbf-b2dd-76f466b03cdb"
	topThreatsCountry.csmobj.SId = "S-44a17b93-b9b3-4356-ab21-ef0a97c8f67d"
	topThreatsCountry.csmobj.Length = "30"

	dcItem.csmobj.CId = "C-a4c0f8fd-ccc9-4dbf-b2dd-76f466b03cdb"
	dcItem.csmobj.SId = "S-44a17b93-b9b3-4356-ab21-ef0a97c8f67d"
	dcItem.csmobj.Length = "30"

	report.registerItem(&topThreatsCountry)
	report.registerItem(&dcItem)
	arr := report.start()
	for _, val := range arr {
		fmt.Println(val)
	}

}
*/
