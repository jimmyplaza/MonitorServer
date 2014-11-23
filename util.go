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


