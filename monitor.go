/* Monitor */
package main

import (
	"code.google.com/p/gcfg"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/jmoiron/jsonq"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	//"bytes"
)

type JsonType struct {
	Status       int    `json:"status"`
	ResponseTime int64  `json:"responsetime"`
	Timestamp    string `json:"@timestamp"`
	Url          string `json:"url"`
	Errmsg       string `json:"errmsg"`
	Rspstatus    string `json:rspstatus`
}

type JsonDnsType struct {
	CustomerName string `json:"customername"`
	Site         string `json:"site"`
	Status       int    `json:"status"`
	Timestamp    string `json:"@timestamp"`
	CurrentIP    string `json:"currentip"`
	Change       int    `json:"change"`
}

type JsonReportType struct {
	Timestamp     string `json:"@timestamp"`
	OnlineUser    int    `json:"onlineuser"`
	Pageviews     int    `json:"pageviews"`
	Visitors      int    `json:"visitors"`
	Threats       int    `json:"threats"`
	Bandwidth     int    `json:"bandwidth"`
	BandwidthPeak int    `json:"bandwithpeak"`
	TotalRequest  int    `json:"totalrequest"`
	CacheHit      int    `json:"cachehit"`
	Legitimated   int    `json:"legitimated"`
	CacheRatio    int    `json:"cacheratio"`
	Upstream      int    `json:"upstream"`
	SiteSpeed     int    `json:"sitespeed"`
}

var debug *bool
var syslogSender *SyslogSender
var filepath1 string = "./log/"
var filepath2 string = "./csmlog/"
var cfg cfgObject //at types.go
var dcobj DCObject
var configFile = flag.String("c", "api.gcfg", "config filename")
var customer *Customers
var allsite AllSite
var allCustomerSite []string
var g2Site []string
var dnsSite DnsSite
var SmtpServer string
var Port string
var From string

func timeoutDialer(cTimeout, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func HttpsGet(url string, funcName string) (rspstring string, err error) {
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //for https
		},
	}
	response, err := myClient.Get(url)
	if err != nil {
		fmt.Printf("[%s] http.Get => %v", funcName, err.Error())
		return "", err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("[%s] readall err: %s", funcName, err)
		return "", err
	}
	rspstring = string(contents)
	return rspstring, nil

}

func CheckVariation() (ReqRatio, LegRatio float64) {
	url := "https://g2api.nexusguard.com/API/Proxy?cust_id=C-a4c0f8fd-ccc9-4dbf-b2dd-76f466b03cdb&kind=60&length=24&site_id=S-44a17b93-b9b3-4356-ab21-ef0a97c8f67d&type=cddInfoData"
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //for https
		},
	}
	response, err := myClient.Get(url)
	if err != nil {
		fmt.Printf("[CheckVariation] http.Get => %v", err.Error())
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("[CheckVariation] readall err: %s", err)
		return
	}
	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(string(contents)))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)

	TotalRequest, _ := jq.Int("cddInfoData", "Reqs", "reqs")
	Threats, _ := jq.Int("cddInfoData", "Threats", "threats")
	Legitimated, _ := jq.Int("cddInfoData", "Legitimated", "Legitimated")
	Upstream, _ := jq.Int("cddInfoData", "Upstream", "Upstream")
	CacheHit, _ := jq.Int("cddInfoData", "CacheData", "CacheHit")
	ReqRatio = (float64(TotalRequest) - float64(Threats) - float64(Legitimated)) / float64(TotalRequest)
	LegRatio = (float64(Legitimated) - float64(CacheHit) - float64(Upstream)) / float64(Legitimated)
	return ReqRatio, LegRatio
}

func MonitorG2Server(Url []string, seconds int, Too []string) {
	var flag_arr = make([]bool, len(Url))
	var cnt int = 0
	var flag_idx int
	var errMsg string
	var To []string
	jj := JsonType{}

	//timeoutDialer (connect timeout, write timeout)
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(15)*time.Second, 15*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //for https
			MaxIdleConnsPerHost:   250,
		},
	}

	for {
		flag_idx = 0
		for _, url := range Url { //monitor all url at array
			url = strings.TrimSpace(url)
			if _, ok := cfg.G2Server[url]; ok {
				To = cfg.G2Server[url].To // Mail owner list depend on different url
			} else {
				To = Too
			}
			if cnt == 0 {
				WriteToLogFile(url, "START MONITORING", "", filepath1)
			}
			t1 := time.Now()
			nanoold := time.Now().UnixNano() / 1000000 //to ms
			rsptime := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04"))
			rsptime = strings.Replace(rsptime, " ", "T", 1)
			response, err := myClient.Get(url)
			nanonew := time.Now().UnixNano() / 1000000 //to ms
			responseTime := fmt.Sprintf("%s", time.Now().Sub(t1))
			jj.ResponseTime = nanonew - nanoold
			jj.Timestamp = rsptime
			jj.Url = url
			var rspStatus string
			var rspCode int
			if response != nil {
				rspStatus = response.Status   //ex: 302 Moved Temporarily
				rspCode = response.StatusCode // ex: 302
			} else {
				rspStatus = ""
			}

			if err != nil {
				WriteToLogFile(url, "DIE", responseTime, filepath1)
				errMsg = fmt.Sprintf("%s", err)
				if strings.Index(errMsg, "timeout") != -1 {
					To[0] = "jimmy.ko@nexusguard.com"
					Title := "[G2Monitor] Only Jimmy(io timeout)- " + "[G2] - " + url + " - Status"
					Body := Title + "<br>" + "STATUS CODE: " + rspStatus + "<br>" + "ERROR: " + errMsg
					MorningMail(SmtpServer, Port, From, To, Title, Body)
					continue
				}
				jj.Status = 1 //down
				if flag_arr[flag_idx] == false && rspCode != 302 {
					Title := "[G2Monitor] - " + "[G2] - " + url + " - Status"
					Body := Title + "<br>" + "STATUS CODE: " + rspStatus + "<br>" + "ERROR: " + errMsg
					if strings.Index(errMsg, "connection") != -1 {
						prd := "0"
						uat := "1"
						GMonitorAudio(prd, uat)
					}
					MorningMail(SmtpServer, Port, From, To, Title, Body)
					WriteToLogFile(url, "SENT MAIL", responseTime, filepath1)
					flag_arr[flag_idx] = true
				}
				//err != nil, response is nil, do response.Body.Close() will get
				//runtime error: invalid memory address or nil pointer dereference
			} else {
				errMsg = "None"
				//if rspCode == 200 || rspCode < 500  {
				if rspCode < 500 {
					if flag_arr[flag_idx] == true { //Revoery Mail, notify service is back
						Title := "[G2Monitor] - " + "[G2] - " + url + " is back"
						Body := Title + "<br>" + "STATUS CODE: " + rspStatus + "<br>" + "ERROR: " + errMsg
						MorningMail(SmtpServer, Port, From, To, Title, Body)
					}
					flag_arr[flag_idx] = false
					WriteToLogFile(url, "ALIVE", responseTime, filepath1)
					jj.Status = 100
				} else {
					jj.Status = 1
					if flag_arr[flag_idx] == false {
						Title := "[G2Monitor]x - " + "[G2] - " + url + " - Status"
						Body := Title + "<br>" + "STATUS CODE: " + rspStatus + "<br>" + "ERROR: " + errMsg
						MorningMail(SmtpServer, Port, From, To, Title, Body)
						WriteToLogFile(url, "SENT MAIL", responseTime, filepath1)
						flag_arr[flag_idx] = true
					}
				}
				//err = nil, response is not nil, need to Close()
				response.Body.Close()
			}
			jj.Errmsg = errMsg
			jj.Rspstatus = rspStatus
			url = strings.Replace(url, "//", "", -1)

			ElkInput("g_monitor", url, jj)
			flag_idx++
		}
		time.Sleep(time.Duration(seconds) * time.Second)
		cnt++
	}
}

func MonitorCustomerServer(Url []string, seconds int, To []string) {
	var flag_arr = make([]bool, len(Url))
	var cnt int = 0
	var flag_idx int
	jj := JsonType{}
	var errMsg string
	errMsg = "nn"

	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(15)*time.Second, 15*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //for https
			//MaxIdleConnsPerHost: 250,
		},
	}

	for {
		flag_idx = 0
		for _, url := range Url { //monitor all url at array
			url = strings.TrimSpace(url)
			if cnt == 0 {
				WriteToLogFile(url, "START MONITORING", "", filepath2)
			}
			t1 := time.Now()
			nanoold := time.Now().UnixNano() / 1000000 //to ms
			rsptime := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04"))
			rsptime = strings.Replace(rsptime, " ", "T", 1)

			response, err := myClient.Get(url)

			nanonew := time.Now().UnixNano() / 1000000 //to ms

			responseTime := fmt.Sprintf("%s", time.Now().Sub(t1))
			jj.ResponseTime = nanonew - nanoold
			//jj.Timestamp = fmt.Sprintf("%s", t1)
			jj.Timestamp = rsptime
			jj.Url = url
			var rspStatus string
			var rspCode int
			if response != nil {
				rspStatus = response.Status
				rspCode = response.StatusCode
			} else {
				rspStatus = ""
			}

			if err != nil {
				WriteToLogFile(url, "DIE", responseTime, filepath2)
				errMsg = fmt.Sprintf("%s", err)
				jj.Status = 1
				WriteToJsonFile(jj)
				if flag_arr[flag_idx] == false && rspCode != 302 {
					//fmt.Println("***********")
					//fmt.Println(rspCode)
					//SendMail(SmtpServer, Port, From, To, url, errMsg, rspStatus)
					fmt.Println(errMsg)
					//WriteToLogFile(url, "SENT MAIL", responseTime)
					flag_arr[flag_idx] = true
				}
				//err != nil, response is nil, do response.Body.Close() will get
				//runtime error: invalid memory address or nil pointer dereference
			} else {
				errMsg = "None"
				//if rspCode == 200 || rspCode < 500  {
				if rspCode < 500 {
					flag_arr[flag_idx] = false
					WriteToLogFile(url, "ALIVE", responseTime, filepath2)
					jj.Status = 100
					WriteToJsonFile(jj)
				} else {
					jj.Status = 1
					WriteToJsonFile(jj)
					if flag_arr[flag_idx] == false {
						fmt.Println(rspStatus)
						//SendMail(SmtpServer, Port, From, To, url, errMsg, rspStatus)
						WriteToLogFile(url, "SENT MAIL", responseTime, filepath2)
						flag_arr[flag_idx] = true
					}
				}
				//err = nil, response is not nil, need to Close()
				response.Body.Close()

			}
			jj.Errmsg = errMsg
			jj.Rspstatus = rspStatus
			//url = strings.Replace(url, "http://", "",-1)
			url = strings.Replace(url, "//", "", -1)
			ElkInput("g_monitor", url, jj)

			flag_idx++
		}
		time.Sleep(time.Duration(seconds) * time.Second)
		cnt++
	}
}

func WriteToLogFile(remote string, msg, responseTime, filepath string) {
	logMsg := "[" + remote + "] , " + msg + " , " + responseTime + " , "
	log.Println(logMsg)
	t := time.Now()
	var trimStr string
	if strings.Index(remote, "http") != -1 {
		trimStr = "http://"
	}
	if strings.Index(remote, "https") != -1 {
		trimStr = "https://"
	}
	trm := strings.Trim(remote, trimStr)
	trimIndex := strings.Index(trm, "?")
	if trimIndex != -1 {
		trm = trm[:trimIndex]
	}

	fileName := t.Format("20060102") + "_" + strings.Replace(trm, "/", ".", -1) + ".log"
	fmt.Println(fileName)
	fmt.Println(logMsg, fileName)
	f, err := os.OpenFile(filepath+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	//defer f.Close()
	if err != nil {
		fmt.Printf("[WriteToLogFile] error opening file: %v", err)
		f.Close()
		//log.Printf("[WriteToLogFile] error opening file: %v", err)
	}
	log.SetOutput(f)
	log.Println(logMsg)
	f.Close()
}

func WriteToJsonFile(jj JsonType) {
	out, _ := json.Marshal(jj)
	logMsg := string(out)
	t := time.Now()
	fileName := "allSite.json"
	fmt.Println(logMsg, fileName)
	f, err := os.OpenFile("./resources/assets/"+t.Format("20060102")+"_"+fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	//defer f.Close()
	//f, err := os.OpenFile("./resources/assets/" + fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		fmt.Printf("error opening file: %v", err)
		f.Close()
	}

	_, err = f.WriteString(logMsg + ", ")
	if err != nil {
		fmt.Printf("error write file: %v", err)
		f.Close()
	}
	f.Close()

}

//func MonitorBandwidth(seconds int, SmtpServer string, Port uint, From string, To []string){
func MonitorBandwidth() {
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(15)*time.Second, 15*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //for https
		},
	}
	/*
		SmtpServer:= cfg.Mail.SmtpServer
		Port := cfg.Mail.Port
		From := cfg.Mail.From
	*/
	seconds := cfg.MonitorBand.IntervalSeconds
	To := cfg.MonitorBand.To

	var m []int
	var url_arr []string
	var errMsg []string
	//var rspStatus string

	//var b interface{}
	//var rspCode int
	tmpurl := "https://g2api.nexusguard.com/API/NetflowBandwidth/2?cust_id="
	tmperr := " has zero Bandwidth recent 10 minutes"
	for i, _ := range customer.List {
		CId := customer.List[i].MoId
		MoAlias := customer.List[i].MoAlias
		for j, _ := range cfg.MonitorBand.MonitorList {
			if MoAlias == cfg.MonitorBand.MonitorList[j] {
				fmt.Println(MoAlias)
				urlstr := tmpurl + CId + "&length=5"
				url_arr = append(url_arr, urlstr)
				errstr := "[Monitor Bandwidth] - " + MoAlias + tmperr
				errMsg = append(errMsg, errstr)
			}
		}
	}

	//errMsg[0] = "AAH has zero Bandwidth recent 10 minutes"
	//errMsg[1] = "HKP has zero Bandwidth recent 10 minutes"
	//errMsg[2] = "EDB has zero Bandwidth recent 10 minutes"
	//url_arr[0] = "https://g2api.nexusguard.com/API/NetflowBandwidth/2?cust_id=C-a4c0f8fd-ccc9-4dbf-b2dd-76f466b03cdb&length=5"
	//url_arr[1] = "https://g2api.nexusguard.com/API/NetflowBandwidth/2?cust_id=C-326318f5-0f1b-4c6e-bb3b-6e69e091c35f&length=5"
	//url_arr[2] = "https://g2api.nexusguard.com/API/NetflowBandwidth/2?cust_id=C-6a44172e-3b5b-4981-a9d8-2f1b853b4c90&length=5"

	for {
		for u, url := range url_arr { //monitor all url at array
			fmt.Println("url: " + url)
			response, err := myClient.Get(url)
			/*
						if response != nil {
				 		   //rspStatus = response.Status
							rspStatus = ""
				    	 	//rspCode = response.StatusCode
						} else{
							rspStatus = ""
						}
			*/
			if err != nil {
				fmt.Printf("%s", err)
				continue
				//os.Exit(1)
			} else {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("%s", err)
					//os.Exit(1)
					continue
				}
				err = json.Unmarshal(contents, &m)
				if err != nil {
					fmt.Println(err)
				}
				//m := b.([]interface{})
				m = m[:3] //the last two value must be zero, trim
				fmt.Println(m)
				for i := 0; i < len(m); i++ {
					if m[i] == 0 {
						MorningMail(SmtpServer, Port, From, To, errMsg[u], errMsg[u])
						//WriteToSyslog(0,"Monitor",errMsg[u])
						//SendMail(SmtpServer, Port, From, To, errMsg[u], errMsg[u], rspStatus)
					}
				}
			}
		}
		time.Sleep(time.Duration(seconds) * time.Second)
	}
}

func MonitorDataCenter(seconds int, To []string) {
	var Timeout time.Duration
	Timeout = 10
	allsite.List = make(map[string]map[string][]DCObject)
	var m map[string][]DCObject
	var monitorListObj = make(map[string][]string)
	var customerFilterList []string
	for i, _ := range cfg.MonitorDC.MonitorList {
		filterarr := strings.Split(cfg.MonitorDC.MonitorList[i], " ")
		customerFilterList = append(customerFilterList, filterarr[0])
		monitorListObj[filterarr[0]] = filterarr[1:]
	}
	/*
		    for i, _ := range customer.List{
			    fmt.Println(customer.List[i].MoAlias)
			    fmt.Println(customer.List[i].SiteAliasList)
			    //fmt.Println(customer.List[i].MoId)
			    //fmt.Println(customer.List[i].SiteList)
		    }
	*/

	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(Timeout)*time.Second, Timeout*time.Second),
			ResponseHeaderTimeout: Timeout * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //for https
		},
	}

	for {
		passFlag := false
		for i, _ := range customer.List {
			CId := customer.List[i].MoId
			MoAlias := customer.List[i].MoAlias
			for j, _ := range customerFilterList {
				if customer.List[i].MoAlias == customerFilterList[j] {
					passFlag = true
					break
				} else {
					passFlag = false
				}
			}
			if passFlag {
				//fmt.Println("MoAlias: ", MoAlias)
				//fmt.Println("CId: ", CId)
				for s, SId := range customer.List[i].SiteList {
					//fmt.Println("SId: ", SId)
					url := fmt.Sprintf("https://g2api.nexusguard.com/API/Proxy?cust_id=%s&site_id=%s&type=dataCenter", CId, SId)
					t1 := time.Now()
					response, err := myClient.Get(url)
					responseTime := fmt.Sprintf("%s", time.Now().Sub(t1))
					if err != nil {
						if *debug {
							fmt.Printf("%s", err)
						}
						WriteToLogFile("DCenter", fmt.Sprintf("%s", err), responseTime, filepath1)
					} else {
						defer response.Body.Close()
						contents, err := ioutil.ReadAll(response.Body)
						if err != nil {
							if *debug {
								fmt.Printf("%s", err)
							}
							WriteToLogFile("DCenter", fmt.Sprintf("%s", err), responseTime, filepath1)
							continue
						}
						err = json.Unmarshal(contents, &m)
						if err != nil {
							if *debug {
								fmt.Printf("%s", err)
							}
							WriteToLogFile("DCenter", fmt.Sprintf("%s", err), responseTime, filepath1)
							fmt.Println(err)
						}
						if _, ok := allsite.List[CId]; !ok {
							allsite.List[CId] = make(map[string][]DCObject)
						}

						if allsite.List[CId][SId] == nil { //First time, store value at mem
							for n, _ := range m["DataCenter"] {
								if n < 3 { //Only monitor HK, MI, SJ datacenter
									dcobj = DCObject{m["DataCenter"][n].CenterName, m["DataCenter"][n].CenterCount}
									allsite.List[CId][SId] = append(allsite.List[CId][SId], dcobj)
								}
							}
						} else {
							for t, _ := range allsite.List[CId][SId] {
								/*fmt.Println("Current record: ============")
								        		fmt.Println( m["DataCenter"][t].CenterCount)
								        		fmt.Println("last record: ============")
										        fmt.Println(allsite.List[CId][SId][t])
										        fmt.Println("t: ", t)
								*/
								if monitorListObj[MoAlias][t] == "1" { //"1" means need to monitor
									//fmt.Println("need to monitor")

									//fmt.Println("old: ",allsite.List[CId][SId][t].CenterCount)
									//fmt.Println("new: ", m["DataCenter"][t].CenterCount )
									if allsite.List[CId][SId][t].CenterCount == m["DataCenter"][t].CenterCount || m["DataCenter"][t].CenterCount == 0 {
										var url string
										var errMsg string
										if m["DataCenter"][t].CenterCount == 0 {
											url = " [" + customer.List[i].MoAlias + "]" + " -  " + "[" + customer.List[i].SiteAliasList[s] + "]" + " - " + allsite.List[CId][SId][t].CenterName + " DC" + " is zero!"
											errMsg = " [Error][" + customer.List[i].MoAlias + "]" + " -  " + "[" + customer.List[i].SiteAliasList[s] + "]" + " - " + allsite.List[CId][SId][t].CenterName + " DC" + " - [" + strconv.Itoa(m["DataCenter"][t].CenterCount) + "]" + " is zero!"
										} else {
											url = " [" + customer.List[i].MoAlias + "]" + " -  " + "[" + customer.List[i].SiteAliasList[s] + "]" + " - " + allsite.List[CId][SId][t].CenterName + " DC"
											errMsg = " [Error][" + customer.List[i].MoAlias + "]" + " -  " + "[" + customer.List[i].SiteAliasList[s] + "]" + " - " + allsite.List[CId][SId][t].CenterName + " DC" + " - Request Value: [" + strconv.Itoa(m["DataCenter"][t].CenterCount) + "]"
										}
										WriteToLogFile("DCenter", errMsg, responseTime, filepath1)
										//WriteToSyslog(0,"Monitor-DCenter",errMsg)
										//SendMail(SmtpServer, Port, From, To, url, errMsg, rspStatus)
										Title := "[G2Monitor] - " + "[Data Center]: " + url
										Body := Title + "<br>" + errMsg
										MorningMail(SmtpServer, Port, From, To, Title, Body)
									} else {
										allsite.List[CId][SId][t].CenterCount = m["DataCenter"][t].CenterCount
										Msg := "[Normal][" + customer.List[i].MoAlias + "]" + " -  " + "[" + customer.List[i].SiteAliasList[s] + "]" + " - " + allsite.List[CId][SId][t].CenterName + " DC" + " - [" + strconv.Itoa(m["DataCenter"][t].CenterCount) + "]"
										WriteToLogFile("DCenter", Msg, responseTime, filepath1)

									}
								} // determine DC center need to monitor, HK, MI, SJ [1 1 1] means all need to monito:479

							}
						} //if
					} // if myClient.Get
				} // range customer site
			} // passFlag
		} //range all customer
		time.Sleep(time.Duration(seconds) * time.Second) //240 sec
	} // forever loop
}

/*Read api.gcfg config*/
func ConfigInit() {
	url := fmt.Sprintf("https://%s/api/customer/list/%s", cfg.Gen.GCenter, GetToken())
	customer.mu.Lock()
	customer.List = getCustomers(url)
	//fmt.Println(customer.List)
	//SiteHttpList
	//SiteHttpsList
	for i, _ := range customer.List {
		SId := customer.List[i].SiteAliasList
		Https := customer.List[i].SiteHttpsList
		Http := customer.List[i].SiteHttpList
		//fmt.Println(SId)
		for i, site := range SId {
			if Https[i] == "443" {
				site_https := "https://" + site
				allCustomerSite = append(allCustomerSite, site_https)
				//fmt.Println(site_https)
			}
			if Http[i] == "80" {
				site_http := "http://" + site
				allCustomerSite = append(allCustomerSite, site_http)
				//fmt.Println(site_http)
			}
		}
	}
	allCustomerSite = removeDuplicates(allCustomerSite)

	//customer.mu.Unlock()
	//url = fmt.Sprintf("http://%s/api/customer/list/%s", cfg.Gen.GCenterPrd, GetToken())
	//prdList := getCustomers(url)
	//customer.List = append(customer.List,prdList... )
	//customer.mu.Unlock()
}

func exe_cmd(cmd string, wg *sync.WaitGroup) (output string) {
	fmt.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	output = fmt.Sprintf("%s", out)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return output
}

/*
visitedURL := map[string]bool {
    "http://www.google.com": true,
    "https://paypal.com": true,
}
if visitedURL[thisSite] {
    fmt.Println("Already been here.")
}
*/

func CheckCacheRatio() {

	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(15)*time.Second, 15*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //for https
		},
	}
	seconds := cfg.CheckCacheRatio.IntervalSeconds
	To := cfg.CheckCacheRatio.To
	CacheRatioBound := cfg.CheckCacheRatio.CacheRatioBound

	MonitorList := cfg.CheckCacheRatio.MonitorList
	MonitorArray := make(map[string]bool)
	for _, MoAlias := range MonitorList {
		MonitorArray[MoAlias] = true
	}

	var url_arr []string
	var errMsg []string

	tmpurl := "https://g2api.nexusguard.com/API/Proxy?cust_id=%s&kind=60&length=24&site_id=%s&type=cddInfoData"
	for i, _ := range customer.List {
		CId := customer.List[i].MoId
		MoAlias := customer.List[i].MoAlias
		if MonitorArray[MoAlias] == true {
			//fmt.Println(MoAlias)
			for s, SId := range customer.List[i].SiteList {
				urlstr := fmt.Sprintf(tmpurl, CId, SId)
				//fmt.Println(urlstr)
				url_arr = append(url_arr, urlstr)
				errstr := "[" + MoAlias + "] -" + customer.List[i].SiteAliasList[s]
				//fmt.Println(errstr)
				errMsg = append(errMsg, errstr)
			}
		}
	}

	for {
		for u, url := range url_arr { //monitor all url at array
			//fmt.Println("url: " + url)
			response, err := myClient.Get(url)
			if err != nil {
				fmt.Printf("%s", err)
				continue
			} else {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("%s", err)
					continue
				}
				data := map[string]interface{}{}
				dec := json.NewDecoder(strings.NewReader(string(contents)))
				dec.Decode(&data)
				jq := jsonq.NewQuery(data)

				TotalRequest, _ := jq.Int("cddInfoData", "Reqs", "reqs")
				Upstream, _ := jq.Int("cddInfoData", "Upstream", "Upstream")
				CacheRatio, _ := jq.Int("cddInfoData", "CacheData", "CachePercent")
				ratio := strconv.Itoa(CacheRatio)
				if CacheRatio < CacheRatioBound {
					Title := "[G2Monitor] - " + "[Cache Ratio]" + errMsg[u] + " Cache rate abnormal!"
					Body := errMsg[u] + "<br>Current ratio: " + ratio + "%"
					MorningMail(SmtpServer, Port, From, To, Title, Body)
				}
				a := (float64(Upstream) / float64(TotalRequest)) * 100
				if a > 80 {
					Title := "[G2Monitor] - " + "[Upstream & TotalRequest Ratio]" + errMsg[u] + " - Upstream & Total Request variation more than 80%"
					Body := errMsg[u] + "<br>Upstream/Total Request ratio: " + strconv.FormatFloat(a, 'g', 2, 64) + "%"
					MorningMail(SmtpServer, Port, From, To, Title, Body)
				}
			}
		}
		time.Sleep(time.Duration(seconds) * time.Second)
	}
}

func DnsCheck() {
	To := cfg.DnsCheck.To
	IntervalSeconds := cfg.DnsCheck.IntervalSeconds
	FilterCustomerList := cfg.DnsCheck.FilterCustomer
	visitedURL := make(map[string]bool)
	for _, site := range FilterCustomerList {
		visitedURL[site] = true
	}

	wg := new(sync.WaitGroup)
	dnsSite.List = make(map[string]string)
	jj := JsonDnsType{}
	for {
		for i, _ := range customer.List {
			CustomerName := customer.List[i].MoAlias
			if visitedURL[CustomerName] {
				continue
			}
			SiteAlias := customer.List[i].SiteAliasList
			//DomainList := customer.List[i].DomainList
			//fmt.Println(DomainList)
			for _, site := range SiteAlias {
				wg.Add(1)
				//go exe_cmd(str, wg)
				cmdstr := "dig +short " + site
				//fmt.Println("Domain: ", DomainList[i])
				currentip := exe_cmd(cmdstr, wg)
				jj.Site = site
				if currentip != "" {
					fmt.Println(currentip)
					curtime := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05"))
					curtime = strings.Replace(curtime, " ", "T", 1)
					jj.Timestamp = curtime
					jj.CustomerName = CustomerName

					if dnsSite.List[site] == "" { //First time, store value at mem
						/*if site == "www.rocars.gov.hk" {
							fmt.Println("*******************")
							fmt.Println("currentip: ", currentip)
						}*/
						dnsSite.List[site] = currentip
						jj.Change = 0 //NOTCHANGE
					} else {
						if dnsSite.List[site] == currentip {
							jj.Change = 0 //NOTCHANGE
							/*
								if site == "www.rocars.gov.hk" {
									fmt.Println("not change!!!!!*******************")
									fmt.Println("currentip: ", currentip)
								}*/
						} else {
							jj.Change = 1 //CHANGE
							Title := "[G2Monitor] - " + "[DNS CHANGE]" + "[" + CustomerName + "] -" + site + " DNS IP change!"
							Body := "[" + CustomerName + "] -" + site + " from " + dnsSite.List[site] + " change to " + currentip
							MorningMail(SmtpServer, Port, From, To, Title, Body)
						}
						dnsSite.List[site] = currentip
					}
					jj.CurrentIP = currentip
					if currentip[:6] == "27.126" {
						jj.Status = 0 //"G2"
					} else {
						jj.Status = 1 //"NOTG2"
					}
					//fmt.Println(jj)
					ElkInput("dnscheck", "dnschange", jj)
				}
			} //for Site Alias
		} // for customer.List
		time.Sleep(time.Duration(IntervalSeconds) * time.Second) //300 sec
	}
}

func WriteToSyslog(level int, remote string, msg string) {
	logMsg := "[" + remote + "]:" + msg
	fmt.Println("logMsg: ", logMsg)
	if *debug {
		fmt.Println(logMsg)
	}
	syslogSender.Write("udp", cfg.System.Syslog, level, "MonitorSys", logMsg)
}

func GMonitorAudio(prd, uat string) {
	uurl := "http://107.167.183.111:5487/gmonitor/issue?prd=%s&uat=%s"
	final_url := fmt.Sprintf(uurl, prd, uat)
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
		},
	}
	_, _ = myClient.Get(final_url)
}

func MorningMail(SmtpServer, Port, From string, Too []string, Title, BodyMsg string) {
	var To string
	for _, t := range Too {
		To = To + t + " , "
	}
	To = To[:len(To)-3]
	uurl := "http://g2tool.cloudapp.net:445/morningbird"
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
		},
	}
	Title = Title
	BodyMsg = Title + "<br>" + BodyMsg
	//BodyMsg = Title + "]" + "<br>STATUS CODE: " + rspStatus + "<br>ERROR: " + BodyMsg

	v := url.Values{}
	v.Set("to", To)
	v.Set("from", "g2.service@nexusguard.com")
	v.Set("subject", Title)
	v.Set("content", BodyMsg)
	v.Set("publickey", "cba2eb")
	v.Set("privatekey", "c3e12e")

	//out, _ := json.Marshal(m)
	//outReader := bytes.NewReader([]byte(out))
	//res, err := myClient.Post(uurl, "application/x-www-form-urlencoded", outReader)
	//res, err := myClient.Post(url, "application/json", outReader)
	//res, err := myClient.PostForm(uurl, url.Values{ "from" : { "g2.service@nexusguard.com" }, "to" : { "jimmy.ko@nexusguard.com, stickbob@gmail.com"  }, "subject" : {"aaaa"}, "content":{"ttttt"}, "publickey":{"cba2eb"}, "privatekey":{"c3e12e"}  })
	res, err := myClient.PostForm(uurl, v)
	if err != nil {
		fmt.Printf("MorningBird Mail Error:%s\n", err)
		return
	}
	if res.StatusCode != 200 {
		fmt.Printf("MorningBird Mail Error code: %d,url:%s\n", res.StatusCode, uurl)
	}
	//err = json.Unmarshal([]byte(res.Body), &obj)
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Read Body Error:%s\n", err)
		//errMsg := fmt.Sprintf("%s",err)
		//WriteToSyslog(3,"Monitor-MorningMail",errMsg)
		res.Body.Close()
	}
	var obj interface{}
	err = json.Unmarshal(contents, &obj)
	if *debug {
		fmt.Println(obj)
	}
	if err != nil {
		//errMsg := fmt.Sprintf("%s",err)
		fmt.Printf("MorningMail JSON Error:%s => %s\n", uurl, err)
		//WriteToSyslog(3,"Monitor-MorningMail",errMsg)
	}
	res.Body.Close()
}

/*
Four Parts:
1. Monitor G2 component
2. Monitor Customer Server
3. Monitor BandWidth(portal value)
4. Monitor DataCenter(portal value)
*/

/*
1.
/G-Monitor/www.nexusguard.com/id..
2. /G-Monitor/www.aaa.bbb/id..
3. /G-Monitor/
4. /G-Monitor/DCenter_AAH_site
              DCenter_
Format: Site:    @timestamp, health, rsptime
DCenter:         @timestamp, normal/error,  [MoAlias], Site, DC

*/

func MonitorVariation(CheckTime string) {
	Now := fmt.Sprintf("%s", time.Now().Format("15:04"))
	if Now == CheckTime {
		a, b := CheckVariation()
		ReqRatio := strconv.FormatFloat(a, 'g', 2, 64)
		LegRatio := strconv.FormatFloat(b, 'g', 2, 64)
		To4 := cfg.CheckVariation.To
		Title := "[G2Monitor] - " + "[AAH] - Legitimate & Served by origin variation"
		Body := "(AAH)Legitimate variation: " + ReqRatio + "<br>Served by origin variation: " + LegRatio
		MorningMail(SmtpServer, Port, From, To4, Title, Body)
	}
}

func GetReport() {
	funcname := "GetReport"
	CheckTime := cfg.GetReport.CheckTime
	To := cfg.GetReport.To
	IntervalSeconds := cfg.GetReport.IntervalSeconds
	jj := JsonReportType{}

	//AAH only
	url := "https://g2api.nexusguard.com/API/Proxy?cust_id=C-a4c0f8fd-ccc9-4dbf-b2dd-76f466b03cdb&site_id=S-44a17b93-b9b3-4356-ab21-ef0a97c8f67d&length=30&type=OnlineUser,AvgPage,cddInfoData,Netflow,SiteSpeed"
	for {
		Now := fmt.Sprintf("%s", time.Now().Format("15:04"))
		content_str, err := HttpsGet(url, "GetReport")
		if err != nil {
			fmt.Println("ERROR: [%s]: HttpsGet-> %v", funcname, err.Error())
			continue
		}
		data := map[string]interface{}{}
		dec := json.NewDecoder(strings.NewReader(content_str))
		dec.Decode(&data)
		jq := jsonq.NewQuery(data)
		OnlineUser, _ := jq.Int("OnlineUser", "S-76a919a5-a247-4728-9860-817b644bfe85")
		Pageviews, _ := jq.Int("AvgPage", "Pageviews")
		Visitors, _ := jq.Int("AvgPage", "Visitors")
		Threats, _ := jq.Int("cddInfoData", "Threats", "threats")
		Bandwidth, _ := jq.Int("Netflow", "BandwidthIn")
		BandwidthPeak, _ := jq.Int("Netflow", "BandwidthMaxIn")
		TotalRequest, _ := jq.Int("cddInfoData", "Reqs", "reqs")
		CacheHit, _ := jq.Int("cddInfoData", "CacheData", "CacheHit")
		Legitimated, _ := jq.Int("cddInfoData", "Legitimated", "Legitimated")
		CacheRatio, _ := jq.Int("cddInfoData", "CacheData", "CachePercent")
		Upstream, _ := jq.Int("cddInfoData", "Upstream", "Upstream")
		SiteSpeed, _ := jq.Int("SiteSpeed", "count")

		if Now == CheckTime {
			content := "[AAH]Report: " + "<br>OnlineUser: " + humanize.Comma(int64(OnlineUser)) + "<br>" +
				"Pageviews: " + humanize.Comma(int64(Pageviews)) + "<br>" +
				"Visitors: " + humanize.Comma(int64(Visitors)) + "<br>" +
				"Threats: " + humanize.Comma(int64(Threats)) + "<br>" +
				"Bandwidth: " + humanize.Bytes(uint64(Bandwidth)) + "<br>" +
				"BandwidthPeak: " + humanize.Bytes(uint64(BandwidthPeak)) + "<br>" +
				"TotalRequest: " + humanize.Comma(int64(TotalRequest)) + "<br>" +
				"CacheHit: " + humanize.Comma(int64(CacheHit)) + "<br>" +
				"Legitimated: " + humanize.Comma(int64(Legitimated)) + "<br>" +
				"CacheRatio: " + humanize.Comma(int64(CacheRatio)) + "%<br>" +
				"Serve by origin: " + humanize.Comma(int64(Upstream)) + "<br>" +
				"SiteSpeed: " + humanize.Comma(int64(SiteSpeed)) + " ms"
			Title := "[G2 Report] - " + "[AAH]"
			Body := content
			MorningMail(SmtpServer, Port, From, To, Title, Body)
		}
		curtime := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05"))
		curtime = strings.Replace(curtime, " ", "T", 1)
		jj.Timestamp = curtime
		jj.OnlineUser = OnlineUser
		jj.Pageviews = Pageviews
		jj.Visitors = Visitors
		jj.Threats = Threats
		jj.Bandwidth = Bandwidth
		jj.BandwidthPeak = BandwidthPeak
		jj.TotalRequest = TotalRequest
		jj.CacheHit = CacheHit
		jj.Legitimated = Legitimated
		jj.CacheRatio = CacheRatio
		jj.Upstream = Upstream
		jj.SiteSpeed = SiteSpeed

		ElkInput("report_idx", "report", jj)
		time.Sleep(time.Duration(IntervalSeconds) * time.Second) //60 sec
	} //Forever loop
}

func main() {
	debug = flag.Bool("debug", false, "Show debug information.")
	flag.Parse()
	err := gcfg.ReadFileInto(&cfg, *configFile)
	if err != nil {
		log.Fatalf("Fail to load config file: %s\n", err)
	}

	customer = &Customers{mu: &sync.Mutex{}}
	ConfigInit() //Read api.gcfg config, get customer.List & allCustomerSite
	syslogSender = &SyslogSender{key: []byte(cfg.System.Key)}

	SmtpServer = cfg.Mail.SmtpServer
	Port = cfg.Mail.Port
	From = cfg.Mail.From
	To1 := cfg.Monitorg2.To

	go CheckCacheRatio()

	go DnsCheck()

	// ===================== Customer Site ===================
	IntervalSeconds2 := cfg.MonitorCustomerSite.IntervalSeconds
	go MonitorCustomerServer(allCustomerSite, IntervalSeconds2, To1)

	// ===================== G2 component Site ===================
	Url := cfg.Monitorg2.Site
	IntervalSeconds := cfg.Monitorg2.IntervalSeconds
	go MonitorG2Server(Url, IntervalSeconds, To1)

	//===================== Portal Customer Bandwidth ===================
	go MonitorBandwidth()

	// ==================== Portal DataCenter =======================
	IntervalSeconds0 := cfg.MonitorDC.IntervalSeconds
	go MonitorDataCenter(IntervalSeconds0, To1)

	CheckTime := cfg.CheckVariation.CheckTime
	for {
		MonitorVariation(CheckTime)
		GetReport()
		time.Sleep(60 * time.Second)
	}

	//var cfg Config  //for main internal use only, has another Global cfg object
	//var Url string
	//var IntervalSeconds int
	//var filename string
	//var separateSymbol rune
	//var combineSymbol string
	//var colNum int
	//cfgFile := "./etc/ini.gcfg"
	/*
		_, err = os.Stat("./log")
		if err != nil {
			log.Println("Directory log not exist, create log dir")
			err := os.Mkdir("./log",0777)
			if err != nil{ os.Exit(1)}
		}
	*/
	//cfg = LoadConfiguration(cfgFile)
	//SmtpServer1 := cfg.Monitorg2.SmtpServer
	//Port1 := cfg.Monitorg2.Port
	//From1 := cfg.Monitorg2.From

	//filename = "./assetsList/assets.tsv"
	//separateSymbol = '\t'
	//combineSymbol = "@ "
	//colNum = 1
	//Url = readCsv(filename, separateSymbol, combineSymbol, colNum)

	//var cfg2 Config
	//var IntervalSeconds2 int
	//cfgFile2 := "./etc/customer_ini.gcfg"
	//_, err = os.Stat("./csmlog")
	//if err != nil {
	//	log.Println("Directory log not exist, create log dir")
	//	err := os.Mkdir("./csmlog",0777)
	//	if err != nil{
	//		os.Exit(1)
	//	}
	//}
	//cfg2 = LoadConfiguration(cfgFile2)
	//SmtpServer2 := cfg2.Server.SmtpServer
	//Port2 := cfg2.Server.Port
	//From2 := cfg2.Server.From
	//To2 := cfg2.Server.To

	//filename = "./assetsList/csm_assets.tsv"
	//separateSymbol = '\t'
	//combineSymbol = "@ "
	//colNum = 6
	//Url2 = readCsv(filename, separateSymbol, combineSymbol, colNum)
	//Url2 = allCustomerSite

	/*	cfgFile2 := "./etc/customer_ini.gcfg"
		cfg2 := LoadConfiguration(cfgFile2)
		SmtpServer := cfg2.Server.SmtpServer
		Port := cfg2.Server.Port
		From := cfg2.Server.From
		To := cfg2.Server.To
	*/

	//===================== api server bandwidth ===================
	/*
		IntervalSeconds3 := cfg.MonitorBand.IntervalSeconds
		SmtpServer3 := cfg.MonitorBand.SmtpServer
		Port3 := cfg.MonitorBand.Port
		From3 := cfg.MonitorBand.From
		To3 := cfg.MonitorBand.To
		//IntervalSeconds3 := 120
		go MonitorBandwidth(IntervalSeconds3, SmtpServer3, Port3, From3, To3)
	*/

}
