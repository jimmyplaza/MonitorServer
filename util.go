package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"code.google.com/p/gcfg"
	"log"
	"time"
	//"net"
	//"net/http"
	//"io/ioutil"
	//"crypto/tls"
	"encoding/json"
)

type Config struct {
    Server struct {
		    Url string
    		IntervalSeconds int 
    		SmtpServer string
    		Port uint
    		From string
    		To string
    }
}



func WriteToLogFile(remote string, msg, responseTime, filepath string) {
	logMsg := "[" + remote + "] , " + msg + " , " + responseTime + " , "
	//log.Println(logMsg)
	t := time.Now()
	//var trimStr string

	remote = strings.Replace(remote, "https://", "", -1)
	remote = strings.Replace(remote, "http://", "", -1)
	remote = strings.Replace(remote, "?", "", -1)

	/*
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
	*/
	//fileName := t.Format("20060102") + "_" + strings.Replace(trm, "/", ".", -1) + ".log"
	fileName := t.Format("20060102") + "_" + strings.Replace(remote, "/", ".", -1) + ".log"
	fmt.Println(fileName)         //20141217_ortal.nexusguard.com.log
	fmt.Println(logMsg, fileName) // [https://portal.nexusguard.com] , DIE , 15.062934514s ,  20141217_ortal.nexusguard.com.log
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



func WriteToSyslog(level int, remote string, msg string) {
	logMsg := "[" + remote + "]:" + msg
	fmt.Println("logMsg: ", logMsg)
	if *debug {
		fmt.Println(logMsg)
	}
	syslogSender.Write("udp", cfg.System.Syslog, level, "MonitorSys", logMsg)
}



/*
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
*/

/*
func HttpsGet(url string, funcName string) (rspcontent []byte, err error) {
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
		return nil, err
	}
	defer response.Body.Close()
	rspcontents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("[%s] readall err: %s", funcName, err)
		return nil, err
	}
	//rspstring = string(contents)
	return rspcontents, nil
}
*/

func CheckDir() {
        _, err := os.Stat("./log")
        if err != nil {
                fmt.Println("Directory log not exist, create log dir")
                err := os.Mkdir("./log", 0777)
                if err != nil {
                        os.Exit(1)
                }
        }
        _, err = os.Stat("./csmlog")
        if err != nil {
                fmt.Println("Directory csmlog not exist, create log dir")
                err := os.Mkdir("./csmlog", 0777)
                if err != nil {
                        os.Exit(1)
                }
        }
}

func GetToken() string {
        t1 := &Token{User: "APIServer", Action: "API", Value: "Token", DF: "G-Center"}
        t1.Make()
        return t1.String()
}

/*
func CheckToken(token string) bool {
        if !cfg.Gen.Dev {
                t1 := &Token{User: "G-Center", Action: "API", Value: "Token", DF: "APIServer"}
                t1.Make()
                if cfg.Gen.Debug {
                        fmt.Println(t1.String())
                }
                return t1.VerifyString(token)
        }
        return true
}
*/
func LogMessage(msg string,level int){
        timeStr := time.Now().Format("2006-01-02 15:04:05")
        lvStr := "info"
        switch level {
                case 1:
                        lvStr = "info"
                case 2:
                        lvStr = "warning"
                case 3:
                        lvStr = "error"
                case 4:
                        lvStr = "crit"
                case 5:
                        lvStr = "panic"
        }
        fmt.Printf("LogMessage [%s][%s][msg] %s\n",timeStr,lvStr,msg)

}



func LoadConfiguration(cfgFile string) Config {
    var err error
    var cfg Config

    if cfgFile!= "" {
        err = gcfg.ReadFileInto(&cfg, cfgFile)
    }
    if err != nil {
        fmt.Println(err)
        log.Printf("Failed to parse gcfg data: %s", err)
        os.Exit(2)
    } 
    return cfg
}

func removeDuplicates(a []string) []string { 
        result := []string{} 
        seen := map[string]string{} 
        for _, val := range a { 
                if _, ok := seen[val]; !ok { 
                        result = append(result, val) 
                        seen[val] = val 
                } 
        } 
        return result 
} 



func readCsv(filename string, separateSymbol rune, combineSymbol string, colNum int)(retStr string){
	file, err := os.Open(filename)
	if err != nil {
		// err is printable
		// elements passed are separated by space automatically
		fmt.Println("Error:", err)
		return
	}
	// automatically call Close() at the end of current method
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = separateSymbol
	//reader.Comma = '\t' 
	lineCount := 0
	url_arr := []string{} 
	for {
		// read just one record, but we could ReadAll() as well
		record, err := reader.Read()
		// end-of-file is fitted into err
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// record is an array of string so is directly printable
		//fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")
		// and we can iterate on top of that
		if lineCount != 0 {
			for i := 0; i < len(record); i++ {
				if i == colNum {
					if record[i] != ""{
						//retStr = retStr + strings.TrimSpace(string(record[i])) + combineSymbol
						url_arr = append(url_arr, strings.TrimSpace(string(record[i])))
					}
				}
			}
		}
		lineCount += 1
	}	
	url_arr = removeDuplicates(url_arr)
	for _, url := range url_arr {
		retStr = retStr + url + combineSymbol
	}	

	retStr = retStr[:len(retStr) -2] 
	return retStr
}

/*
func main() {
	filename := "assets2.tsv"
	//separateSymbol := "," 
	separateSymbol := '\t' 
	combineSymbol := "@ "
	colNum := 1
	retStr := readCsv(filename, separateSymbol, combineSymbol, colNum)
	//retStr := readCsv(filename, combineSymbol, colNum)
	fmt.Println(retStr)
}
*/
