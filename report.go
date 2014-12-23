package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

type JsonLiveReportType struct {
	Timestamp            string `json:"@timestamp"`
	Threats_min          int    `json:"threats_min"`
	Threats_max          int    `json:"threats_max"`
	Threats_avg          int    `json:"threats_avg"`
	NetflowBandwidth_min int    `json:"netflowBandwidth_min"`
	NetflowBandwidth_max int    `json:"netflowBandwidth_max"`
	NetflowBandwidth_avg int    `json:"netflowBandwidth_avg"`
	LiveReqsChart_min    int    `json:"liveReqsChart_min"`
	LiveReqsChart_max    int    `json:"liveReqsChart_max"`
	LiveReqsChart_avg    int    `json:"liveReqsChart_avg"`
	CacheHit_min         int    `json:"cacheHit_min"`
	CacheHit_max         int    `json:"cacheHit_max"`
	CacheHit_avg         int    `json:"cacheHit_avg"`
	Legitimated_min      int    `json:"legitimated_min"`
	Legitimated_max      int    `json:"legitimated_min"`
	Legitimated_avg      int    `json:"legitimated_min"`
	Upstream_min         int    `json:"upstream_min"`
	Upstream_max         int    `json:"upstream_max"`
	Upstream_avg         int    `json:"upstream_avg"`
}

func (n Nodes) GetMinMaxAvg() (min, max, avg int) {
	intArray := []int{}
	var sum int
	for _, val := range n {
		v := val[1]
		sum = sum + v
		intArray = append(intArray, v)
	}
	sort.Ints(intArray)
	arr_len := len(intArray)
	avg = sum / arr_len
	min = intArray[0]
	max = intArray[arr_len-1]
	return min, max, avg
}

func GetLiveReport(cid, sid, length string) (out JsonLiveReportType) {
	var report Report
	funcname := "GetReport"
	tmp_url := "https://g2api.nexusguard.com/API/Proxy?cust_id=%s&length=%s&site_id=%s&type=Pageviews2,Visitors2,NetflowBandwidth,liveThreatsChart,liveReqsChart,liveCacheChart,liveLegitimatedChart,liveUpstreamChart"
	url := fmt.Sprintf(tmp_url, cid, length, sid)

	content, err := HttpsGet(url, funcname)
	if err != nil {
		fmt.Println("ERROR: [%s]: HttpsGet-> %v", funcname, err.Error())
		return
	}
	err = json.Unmarshal(content, &report)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	Threats_min, Threats_max, Threats_avg := report.LiveThreatsChart["Threats"].GetMinMaxAvg()
	NetflowBandwidth_min, NetflowBandwidth_max, NetflowBandwidth_avg := report.NetflowBandwidth[:len(report.NetflowBandwidth)-2].GetMinMaxAvg()
	LiveReqsChart_min, LiveReqsChart_max, LiveReqsChart_avg := report.LiveReqsChart["Reqs"].GetMinMaxAvg()
	CacheHit_min, CacheHit_max, CacheHit_avg := report.LiveCacheChart["CacheHit"].GetMinMaxAvg()
	Legitimated_min, Legitimated_max, Legitimated_avg := report.LiveLegitimatedChart["Legitimated"].GetMinMaxAvg()
	Upstream_min, Upstream_max, Upstream_avg := report.LiveUpstreamChart["Upstream"].GetMinMaxAvg()

	out.Threats_min = Threats_min
	out.Threats_max = Threats_max
	out.Threats_avg = Threats_avg
	out.NetflowBandwidth_min = NetflowBandwidth_min
	out.NetflowBandwidth_max = NetflowBandwidth_max
	out.NetflowBandwidth_avg = NetflowBandwidth_avg
	out.LiveReqsChart_min = LiveReqsChart_min
	out.LiveReqsChart_max = LiveReqsChart_max
	out.LiveReqsChart_avg = LiveReqsChart_avg
	out.CacheHit_min = CacheHit_min
	out.CacheHit_max = CacheHit_max
	out.CacheHit_avg = CacheHit_avg
	out.Legitimated_min = Legitimated_min
	out.Legitimated_max = Legitimated_max
	out.Legitimated_avg = Legitimated_avg
	out.Upstream_min = Upstream_min
	out.Upstream_max = Upstream_max
	out.Upstream_avg = Upstream_avg
	return out

}
