package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	//"net"
)

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
func ElkInput(index, table string, obj interface{}) {
	url := "http://g2tool.cloudapp.net:9200/" + index + "/" + table + "/"
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
		},
	}
	out, _ := json.Marshal(obj)
	outReader := bytes.NewReader(out)
	res, err := myClient.Post(url, "application/x-www-form-urlencoded", outReader)
	if err != nil {
		fmt.Printf("\nELK Get Error:%s\n", err)
		return
	}
	if res.StatusCode == 200 || res.StatusCode == 201 {
		fmt.Printf("\nELK input seccuessful. _Source: %s\n", out)
	} else {
		fmt.Printf("\nELK Get Error code: %d,url:%s\n", res.StatusCode, url)
	}
	res.Body.Close()
}

func ElkGet(index, table string, int_id int) {
	id := strconv.Itoa(int_id)
	url := "http://g2tool.cloudapp.net:9200/" + index + "/" + table + "/" + id + "/"
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
		},
	}
	response, err := myClient.Get(url)
	if err != nil {
		log.Fatalf("http.Get => %v", err.Error())
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("\n%v\n\n", string(body))
}

func ElkGetAll(index, table string) {
	url := "http://g2tool.cloudapp.net:9200/" + index + "/" + table + "/_search?pretty"
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(10)*time.Second,
				time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
		},
	}
	response, err := myClient.Get(url)
	if err != nil {
		log.Fatalf("http.Get => %v", err.Error())
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("\n%v\n\n", string(body))
}

/*
func main(){
   type Log struct{
        Health string
        Rsptime int
        Timestamp string
    }
    var log Log
    log.Health = "ALIVE"
    //log.Rsptime = 99
    //log.Timestamp = "2014-11-11T10:00:00+00:00"
    index := "index_t2"
    table := "table_t2"
    //ElkInput(index, table,  strconv.Itoa(100), log)

   // for i:=1; i<5 ; i++{
   //     responseTime := fmt.Sprintf("%s", time.Now())
   //     log.Timestamp = responseTime
   //     log.Rsptime = i
   //     ElkInput(index, table,  strconv.Itoa(i), log)
   // }

    //ElkGet(index, table, 14)
    ElkGetAll(index, table)
}


*/
