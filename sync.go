package main

import (	
	"encoding/json"
	"fmt"	
	"io/ioutil"
	"net/http"
	"time"
	"crypto/tls"

)

func getCustomers(url string) []customerObject {	
	var myClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(cfg.Gen.Timeout)*time.Second,
				time.Duration(cfg.Gen.Timeout)*time.Second),
			ResponseHeaderTimeout: time.Second * 2,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	res, err := myClient.Get(url)
	if err != nil {
		fmt.Printf("Get Error:%s\n", err)
		return nil
	}
	if res.StatusCode != 200 {
		fmt.Printf("getCustomers error code: %d,url:%s\n", res.StatusCode, url)
		res.Body.Close()
		return nil
	}
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Read Body Error:%s\n", err)
		res.Body.Close()
		return nil
	}
	res.Body.Close()

	var obj []customerObject
	err = json.Unmarshal(contents, &obj)
	if err != nil {
		fmt.Printf("getCustomers JSON Error:%s => %s\n", url, err)
		return nil
	}
	return obj
}

