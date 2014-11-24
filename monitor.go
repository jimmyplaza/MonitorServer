/* Monitor */
package main 

import (
	"fmt"
	"net/url"
	"net"
	"net/http" 
	"strings"
	"os"
	"time"
	"log"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"flag"
	"code.google.com/p/gcfg"
	"sync"
	"strconv"
    "github.com/jmoiron/jsonq"
     "os/exec"
	//"bytes"
)


type JsonType struct {
	Status int `json:"status"` 
	ResponseTime int64 `json:"responsetime"`
	Timestamp string `json:"@timestamp"`
	Url string `json:"url"`
	Errmsg string `json:"errmsg"`
	Rspstatus string `json:rspstatus`
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

func GetCacheRatio()(ReqRatio, LegRatio float32){
    url := "https://g2api.nexusguard.com/API/Proxy?cust_id=C-a4c0f8fd-ccc9-4dbf-b2dd-76f466b03cdb&kind=60&length=24&site_id=S-44a17b93-b9b3-4356-ab21-ef0a97c8f67d&type=cddInfoData"
    var myClient = &http.Client{
            Transport: &http.Transport{
                    Dial: timeoutDialer(time.Duration(10)*time.Second,
                            time.Duration(10)*time.Second),
                    ResponseHeaderTimeout: time.Second * 10,
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //for https
            },
    }
    response, err := myClient.Get(url) 
    if err != nil {
        //log.Fatalf("http.Get => %v", err.Error())
        fmt.Printf("[GetCacheRatio] http.Get => %v", err.Error())
        return
    }
    defer response.Body.Close()
    contents, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Printf("[GetCacheRatio] readall err: %s", err)
        return
    }
    data := map[string]interface{}{}
    dec := json.NewDecoder(strings.NewReader(string(contents)))
    dec.Decode(&data)
    jq := jsonq.NewQuery(data)

    TotalRequest , _ := jq.Int("cddInfoData", "Reqs", "reqs")
    Threats, _ := jq.Int("cddInfoData", "Threats", "threats")
    Legitimated, _ := jq.Int("cddInfoData", "Legitimated", "Legitimated")
    Upstream , _ := jq.Int("cddInfoData", "Upstream", "Upstream")
    CacheHit, _ := jq.Int("cddInfoData", "CacheData", "CacheHit")
    //ReqRatio = ( TotalRequest - Threats - Legitimated ) / TotalRequest
    ReqRatio = ( float32(TotalRequest) - float32(Threats) - float32(Legitimated) ) / float32(TotalRequest)
    //( Legitimated  - cache hit - Served by origin(Upsream) ) / Legitimated
    LegRatio = ( float32(Legitimated) - float32(CacheHit) - float32(Upstream) ) / float32(Legitimated)
    return ReqRatio, LegRatio
    /*
    fmt.Println("TotalRequest: ",TotalRequest)
    fmt.Println("Threats: ",Threats)
    fmt.Println("Legitimated: ",Legitimated)
    fmt.Println("Upstream: ", Upstream)
    fmt.Println("CacheHit: ", CacheHit)
    */

    //var m interface{}
    //var str map[string]interface{}
    //TotalRequest := str["cddInfoData"].(map[string]interface{})["Reqs"].(map[string]interface{})["reqs"]
    //Threats := str["cddInfoData"].(map[string]interface{})["Threats"].(map[string]interface{})["threats"]
    //Legitimated := str["cddInfoData"].(map[string]interface{})["Legitimated"].(map[string]interface{})["Legitimated"]
    //var to int
    //to = strconv.Atoi(TotalRequest)
    //fmt.Println(reflect.TypeOf(TotalRequest))
    //fmt.Println(reflect.TypeOf(Threats))
    //fmt.Println(reflect.TypeOf(Legitimated))
}


//func MonitorG2Server(Url []string, seconds int, SmtpServer string, Port string, From string, Too []string){
func MonitorG2Server(Url []string, seconds int, Too []string){
	//url_arr := strings.Split(Url,"@")
	//var flag bool
	var flag_arr = make([]bool,len(Url))
	var cnt int = 0
	var flag_idx int
	var errMsg string
	var To []string
	jj := JsonType{}

	//timeoutDialer (connect timeout, write timeout)
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(15)*time.Second, 15 * time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //for https
			MaxIdleConnsPerHost: 250,
		},
	}
	
	for{
		flag_idx = 0
		for _, url := range Url{	//monitor all url at array 
			url  = strings.TrimSpace(url)
			if _, ok := cfg.G2Server[url]; ok{
			    To = cfg.G2Server[url].To  // Mail owner list depend on different url
			}else {
		    	To = Too
			}

			if cnt == 0 {
				WriteToLogFile(url, "START MONITORING","",filepath1)
			    //WriteToSyslog(5,"Monitor","START MONITORING")
			}
			t1 := time.Now()
			nanoold := time.Now().UnixNano()/1000000 //to ms
			rsptime := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04") )
        	rsptime = strings.Replace(rsptime," ", "T", 1)

			response, err := myClient.Get(url) 

			nanonew := time.Now().UnixNano()/1000000 //to ms
			responseTime := fmt.Sprintf("%s", time.Now().Sub(t1)) 

			jj.ResponseTime = nanonew - nanoold

			//jj.Timestamp = fmt.Sprintf("%s", t1)
			jj.Timestamp = rsptime 
			jj.Url = url
			var rspStatus string
			var rspCode int
			if response != nil {
		        rspStatus = response.Status   //ex: 302 Moved Temporarily
		        rspCode = response.StatusCode  // ex: 302
			} else{
				rspStatus = ""
			}

			if err != nil {
				WriteToLogFile(url, "DIE", responseTime, filepath1)
				errMsg = fmt.Sprintf("%s",err)
				jj.Status = 1
				WriteToJsonFile(jj)
				if flag_arr[flag_idx] == false && rspCode != 302{
					//fmt.Println("***********")
					//fmt.Println(rspCode)
					//SendMail(SmtpServer, Port, From, To, url, errMsg, rspStatus)
					MorningMail(SmtpServer, Port, From, To,  url , errMsg, rspStatus)
					WriteToLogFile(url, "SENT MAIL", responseTime,filepath1)
				    //WriteToSyslog(0,"Monitor",errMsg)
					flag_arr[flag_idx] = true
				}
				//err != nil, response is nil, do response.Body.Close() will get
				//runtime error: invalid memory address or nil pointer dereference
			} else {
				errMsg = "None"
				//if rspCode == 200 || rspCode < 500  {
				if rspCode < 500  {
					flag_arr[flag_idx] = false
					WriteToLogFile(url, "ALIVE", responseTime, filepath1)
					jj.Status = 100
					WriteToJsonFile(jj)
				} else {
					jj.Status = 1
					WriteToJsonFile(jj)
					if flag_arr[flag_idx] == false{
						fmt.Println(rspStatus)
						//SendMail(SmtpServer, Port, From, To, url, errMsg, rspStatus)
						MorningMail(SmtpServer, Port, From, To,  url , errMsg, rspStatus)
						WriteToLogFile(url, "SENT MAIL", responseTime,filepath1)
					    //WriteToSyslog(0,"Monitor",errMsg)
						flag_arr[flag_idx] = true
					}
				}
				//err = nil, response is not nil, need to Close()

				response.Body.Close()
				//ElkInput(index, table, "", obj interface{}){
			}
			jj.Errmsg = errMsg
			jj.Rspstatus = rspStatus
			url = strings.Replace(url, "https://", "",-1)
			url = strings.Replace(url, "http://", "",-1)

			ElkInput("g_monitor", url, jj)
			flag_idx++
		}
     	time.Sleep(time.Duration(seconds) * time.Second)
		cnt++
	}
}


func MonitorCustomerServer(Url []string, seconds int, To []string){
	var flag_arr = make([]bool,len(Url))
	var cnt int = 0
	var flag_idx int
	jj := JsonType{}
	var errMsg string
	errMsg = "nn"

	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(15)*time.Second, 15 * time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //for https
			//MaxIdleConnsPerHost: 250,
		},
	}
	
	for{
		flag_idx = 0
		for _, url := range Url {	//monitor all url at array 
			url  = strings.TrimSpace(url)
			//url = "http://" + url
			if cnt == 0 {
				WriteToLogFile(url, "START MONITORING","", filepath2)
			}
			t1 := time.Now()
			nanoold := time.Now().UnixNano()/1000000 //to ms
			rsptime := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04") )
        	rsptime = strings.Replace(rsptime," ", "T", 1)

			response, err := myClient.Get(url) 

			nanonew := time.Now().UnixNano()/1000000 //to ms

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
			} else{
				rspStatus = ""
			}

			if err != nil {
				WriteToLogFile(url, "DIE", responseTime, filepath2)
				errMsg = fmt.Sprintf("%s",err)
				jj.Status = 1
				WriteToJsonFile(jj)
				if flag_arr[flag_idx] == false && rspCode != 302{
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
				if rspCode < 500  {
					flag_arr[flag_idx] = false
					WriteToLogFile(url, "ALIVE", responseTime, filepath2)
					jj.Status = 100
					WriteToJsonFile(jj)
				} else {
					jj.Status = 1
					WriteToJsonFile(jj)
					if flag_arr[flag_idx] == false{
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
			url = strings.Replace(url, "https://", "",-1)
			url = strings.Replace(url, "http://", "",-1)
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
	if strings.Index(remote,"http") != -1 {
		trimStr = "http://"
	}
	if strings.Index(remote,"https") != -1 {
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
	f, err := os.OpenFile(filepath + fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
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
	f, err := os.OpenFile("./resources/assets/" + t.Format("20060102") + "_" + fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	//defer f.Close()
	//f, err := os.OpenFile("./resources/assets/" + fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		fmt.Printf("error opening file: %v", err)
		f.Close()
	}
	
	_, err = f.WriteString(logMsg+", ")
	if err != nil {
		fmt.Printf("error write file: %v", err)
		f.Close()
	}
	f.Close()
	
}

//func MonitorBandwidth(seconds int, SmtpServer string, Port uint, From string, To []string){
func MonitorBandwidth(){
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(15)*time.Second, 15 * time.Second),
			ResponseHeaderTimeout: time.Second * 10,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //for https
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
	for i, _ := range customer.List{
		CId := customer.List[i].MoId
		MoAlias := customer.List[i].MoAlias
		for j, _ := range cfg.MonitorBand.MonitorList{
			if MoAlias == cfg.MonitorBand.MonitorList[j]{
				fmt.Println(MoAlias)
				urlstr := tmpurl + CId + "&length=5" 
				url_arr = append(url_arr, urlstr)
				errstr :=  MoAlias + tmperr
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
		for u, url := range url_arr {	//monitor all url at array 
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
		        if err != nil{
		        	fmt.Println(err)
		        }
			    //m := b.([]interface{})
		        m = m[:3] //the last two value must be zero, trim 
		        fmt.Println(m)
		        for i:=0; i < len(m); i++{
		        	if m[i] == 0 {
						MorningMail(SmtpServer, Port, From, To,  errMsg[u] , errMsg[u], "")
					    //WriteToSyslog(0,"Monitor",errMsg[u])
		        		//SendMail(SmtpServer, Port, From, To, errMsg[u], errMsg[u], rspStatus)
		        	}
		        } 
		    }
	    }	
     	time.Sleep(time.Duration(seconds) * time.Second)
	}
}

func MonitorDataCenter(seconds int, To []string){
	var Timeout time.Duration
	Timeout =  10	
	allsite.List = make(map[string]map[string][]DCObject)
	var m map[string][]DCObject
    var monitorListObj = make(map[string][]string)
    var customerFilterList []string
    for i,_ := range cfg.MonitorDC.MonitorList{
    	filterarr := strings.Split(cfg.MonitorDC.MonitorList[i]," ")
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
			Dial: timeoutDialer(time.Duration(Timeout)*time.Second, Timeout* time.Second),
			ResponseHeaderTimeout: Timeout * time.Second,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //for https
		},
	}

	for {
		passFlag := false 
		for i, _ := range customer.List{
			CId := customer.List[i].MoId
			MoAlias := customer.List[i].MoAlias
			for j, _ := range customerFilterList{
				if customer.List[i].MoAlias == customerFilterList[j]{
					passFlag = true
					break
				}else{
					passFlag = false
				}
			}
			if passFlag {
				//fmt.Println("MoAlias: ", MoAlias) 
				//fmt.Println("CId: ", CId) 
				for s, SId := range customer.List[i].SiteList{
					//fmt.Println("SId: ", SId)
					url := fmt.Sprintf("https://g2api.nexusguard.com/API/Proxy?cust_id=%s&site_id=%s&type=dataCenter",CId,SId)
					t1 := time.Now()
					response, err := myClient.Get(url) 
					responseTime := fmt.Sprintf("%s", time.Now().Sub(t1)) 
					if err != nil {
						if *debug{
							fmt.Printf("%s", err)
						}
						WriteToLogFile("DCenter", fmt.Sprintf("%s", err), responseTime, filepath1)
					} else {
				        defer response.Body.Close()
				        contents, err := ioutil.ReadAll(response.Body)
				        if err != nil {
				        	if *debug{
					            fmt.Printf("%s", err)
				        	}
							WriteToLogFile("DCenter", fmt.Sprintf("%s", err), responseTime, filepath1)
				            continue
				        }
				        err = json.Unmarshal(contents, &m)
				        if err != nil{
				        	if *debug{
					            fmt.Printf("%s", err)
				        	}
							WriteToLogFile("DCenter", fmt.Sprintf("%s", err), responseTime, filepath1)
				        	fmt.Println(err)
				        }
						if _,ok := allsite.List[CId];!ok{
							allsite.List[CId] = make(map[string][]DCObject)	
						}

						if allsite.List[CId][SId] == nil{ //First time, store value at mem
				        	for n, _ := range m["DataCenter"]{
				        		if n < 3 { //Only monitor HK, MI, SJ datacenter
									dcobj = DCObject{m["DataCenter"][n].CenterName, m["DataCenter"][n].CenterCount}
									allsite.List[CId][SId] = append(allsite.List[CId][SId],dcobj)
				        		}
				        	}
				        }else {
				        	for t, _ := range allsite.List[CId][SId]{
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
							        if allsite.List[CId][SId][t].CenterCount == m["DataCenter"][t].CenterCount || m["DataCenter"][t].CenterCount == 0{
							        	rspStatus := ""
							        	var url string
							        	var errMsg string
							        	if m["DataCenter"][t].CenterCount == 0{
								        	url =  " [" + customer.List[i].MoAlias + "]" + " -  " + "[" +customer.List[i].SiteAliasList[s] + "]"+ " - " + allsite.List[CId][SId][t].CenterName + " DC" + " is zero!"
								        	errMsg =  " [Error][" + customer.List[i].MoAlias + "]" + " -  " + "[" +customer.List[i].SiteAliasList[s] + "]"+ " - " + allsite.List[CId][SId][t].CenterName + " DC" + " - [" + strconv.Itoa(m["DataCenter"][t].CenterCount) + "]" + " is zero!"
							        	} else {
								        	url =  " [" + customer.List[i].MoAlias + "]" + " -  " + "[" +customer.List[i].SiteAliasList[s] + "]"+ " - " + allsite.List[CId][SId][t].CenterName + " DC" 
								        	errMsg =  " [Error][" + customer.List[i].MoAlias + "]" + " -  " + "[" +customer.List[i].SiteAliasList[s] + "]"+ " - " + allsite.List[CId][SId][t].CenterName + " DC" + " - [" + strconv.Itoa(m["DataCenter"][t].CenterCount) + "]"
							        	}
										WriteToLogFile("DCenter", errMsg, responseTime, filepath1)
									    //WriteToSyslog(0,"Monitor-DCenter",errMsg)
										//SendMail(SmtpServer, Port, From, To, url, errMsg, rspStatus)
										MorningMail(SmtpServer, Port, From, To,  url, errMsg, rspStatus)
							        }else{
							        	allsite.List[CId][SId][t].CenterCount = m["DataCenter"][t].CenterCount 
							        	Msg :=  "[Normal][" + customer.List[i].MoAlias + "]" + " -  " + "[" +customer.List[i].SiteAliasList[s] + "]"+ " - " + allsite.List[CId][SId][t].CenterName + " DC" + " - [" + strconv.Itoa(m["DataCenter"][t].CenterCount) + "]" 
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
		for i, _ := range customer.List{
			SId := customer.List[i].SiteAliasList
			//fmt.Println(SId)
			for _, site := range SId{
				site = "http://" + site
				allCustomerSite = append(allCustomerSite, site)
			}
		}
		allCustomerSite = removeDuplicates(allCustomerSite)

        //customer.mu.Unlock()
        //url = fmt.Sprintf("http://%s/api/customer/list/%s", cfg.Gen.GCenterPrd, GetToken())
        //prdList := getCustomers(url)
        //customer.List = append(customer.List,prdList... )
        //customer.mu.Unlock()
}


func exe_cmd(cmd string, wg *sync.WaitGroup)(output string) {
  fmt.Println("command is ",cmd)
  // splitting head => g++ parts => rest of the command
  parts := strings.Fields(cmd)
  head := parts[0]
  parts = parts[1:len(parts)]

  out, err := exec.Command(head,parts...).Output()
  if err != nil {
    fmt.Printf("%s", err)
  }
  output = fmt.Sprintf("%s", out)
  wg.Done() // Need to signal to waitgroup that this goroutine is done
  return output 
}



/*
CustomerName
@timestamp
Status: (G2/NotG2)
CurrentIP:
Change: (Change/NotChange)
*/

type JsonDnsType struct {
	CustomerName string `json:"customername"`
	Status string `json:"status"` 
	Timestamp string `json:"@timestamp"`
	CurrentIP string `json:"currentip"`
	Change string `json:Change`
} 


func DnsCheck(){
	IntervalSeconds := cfg.DnsCheck.IntervalSeconds
	wg := new(sync.WaitGroup)
	dnsSite.List = make(map[string]string)
	jj := JsonDnsType{}
	for {
		for i, _ := range customer.List{
			CustomerName := customer.List[i].MoAlias 
			SiteAlias := customer.List[i].SiteAliasList

			DomainList := customer.List[i].DomainList
			fmt.Println(DomainList)
			for i, site := range SiteAlias {
				wg.Add(1)
	        	//go exe_cmd(str, wg)
	        	cmdstr := "dig +short " + site 
	        	fmt.Println("Domain: ", DomainList[i])
	        	currentip := exe_cmd(cmdstr, wg)
	        	if currentip != "" {
		        	fmt.Println(currentip)
		        	curtime := fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04") )
		        	curtime = strings.Replace(curtime," ", "T", 1)
		        	jj.Timestamp = curtime
		        	jj.CustomerName = CustomerName

		        	if dnsSite.List[site] == "" {//First time, store value at mem
		        		dnsSite.List[site] = currentip
		        		jj.Change = "NOTCHANGE"
		        	} else{
		        		if dnsSite.List[site] == currentip{
		        			jj.Change = "NOTCHANGE"
		        		}else{
		        			jj.Change = "CHANGE"
		        		}
		        	}
		        	jj.CurrentIP = currentip
		        	if currentip[:6] == "27.126"{
			        	jj.Status = "G2"
		        	} else{
		        		jj.Status = "NOTG2"
		        	}
		        	//fmt.Println(jj)
	        		ElkInput("ggtest", "dnscheck", jj)
	        	}
			} //for Site Alias
		} // for customer.List
     	time.Sleep(time.Duration(IntervalSeconds) * time.Second) //300 sec
	}
}


func WriteToSyslog(level int,remote string, msg string) {
	logMsg := "[" + remote + "]:" + msg	
	fmt.Println("logMsg: ", logMsg)
	if *debug {
		fmt.Println(logMsg)
	}
	syslogSender.Write("udp", cfg.System.Syslog, level,"MonitorSys",logMsg)
}

func MorningMail(SmtpServer, Port, From string, Too []string, Title, BodyMsg, rspStatus string){
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
    BodyMsg = "[" + Title + "]" + " is down\n" + "STATUS CODE: " + rspStatus + "\nERROR: " + BodyMsg 
    Title = "[G2Monitor]" + " - [" + Title + "]" + " status: stop "

    v := url.Values{}
    v.Set("to", To)
    v.Set("from", "g2.service@nexusguard.com")
    v.Set("subject",Title)
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
    if *debug{
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

var SmtpServer string
var Port string
var From string

func main() {
	debug = flag.Bool("debug", false, "Show debug information.")
	flag.Parse()
	err := gcfg.ReadFileInto(&cfg, *configFile)
    if err != nil {
        log.Fatalf("Fail to load config file: %s\n", err)
    }
    customer = &Customers{mu: &sync.Mutex{}}
    ConfigInit()   //Read api.gcfg config, get customer.List & allCustomerSite
    syslogSender = &SyslogSender{key:[]byte(cfg.System.Key)}
    go DnsCheck()
    


    SmtpServer= cfg.Mail.SmtpServer
	Port = cfg.Mail.Port
	From = cfg.Mail.From
	To1 := cfg.Monitorg2.To
	


	// ===================== G2 component Site ===================
	
	Url := cfg.Monitorg2.Site
	IntervalSeconds := cfg.Monitorg2.IntervalSeconds
		// ===================== Customer Site ===================
	IntervalSeconds2 := cfg.MonitorCustomerSite.IntervalSeconds
	go MonitorCustomerServer(allCustomerSite, IntervalSeconds2, To1)


	go MonitorG2Server(Url, IntervalSeconds, To1)

	

	//===================== Portal Customer Bandwidth ===================
	go MonitorBandwidth()

	// ==================== Portal DataCenter =======================
	IntervalSeconds0 := cfg.MonitorDC.IntervalSeconds 
	go MonitorDataCenter(IntervalSeconds0, To1)

	//==================== Start Server Service ==================
	//go httpService()
	CheckTime := cfg.Cacheratio.CheckTime
	for{
		Now := fmt.Sprintf("%s", time.Now().Format("15:04"))

		if Now == CheckTime {
			a , b := GetCacheRatio()
			ReqRatio := fmt.Sprintf("%s", a)
			LegRatio := fmt.Sprintf("%s", b) 
			To4 := cfg.Cacheratio.To
			url := "[G2Monitor] - " + "Legitimate variation &  " + "Served by origin variation"
			errMsg := "[G2Monitor] - " + "(AAH)Legitimate variation: " + ReqRatio + "<br>Served by origin variation: " +  LegRatio
			MorningMail(SmtpServer, Port, From, To4, url, errMsg, "")
		}
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

