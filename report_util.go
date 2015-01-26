package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

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

func HttpPost(url string, postdata []byte) (contents []byte, err error) {
	var myClient = &http.Client{
		Transport: &http.Transport{
			//Dial: timeoutDialer(time.Duration(10)*time.Second,
			//	time.Duration(10)*time.Second),
			ResponseHeaderTimeout: time.Second * 10,
		},
	}
	outReader := bytes.NewReader(postdata)
	res, err := myClient.Post(url, "application/json", outReader)
	//defer res.Body.Close()
	if err != nil {
		fmt.Printf("\n[HttpPost] ERROR:%s\n", err)
		return nil, err
	}
	if res.StatusCode == 200 {
		fmt.Printf("\n[HttpPost] Post Seccuess.\n\n")
	} else {
		errmsg := fmt.Sprintf("\nError code: %d,url:%s\n", res.StatusCode, url)

		return nil, errors.New(errmsg)
	}
	contents, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	return contents, nil
}
