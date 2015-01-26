package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type DataSource interface {
	Get() error
}

/*
type DataSource struct {
	HTTPDataSource
	HTTPSDataSource
}
*/

type HTTPDataSource struct{}
type HTTPSDataSource struct{}

func (h HTTPSDataSource) Get(url string, c chan []byte) error {
	var myClient = &http.Client{
		Timeout: time.Duration(10) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	response, err := myClient.Get(url)
	if err != nil {
		fmt.Printf("HTTPSDataSource Get() => %v", err.Error())
		return err
	}
	defer response.Body.Close()
	rspcontents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("HTTPSDataSource ioutil.ReadAll() err: %s", err)
		return err
	}
	c <- rspcontents
	return nil
}

func (h HTTPDataSource) Get(url string, c chan []byte) (err error) {
	var myClient = &http.Client{
		Timeout:   time.Duration(10) * time.Second,
		Transport: &http.Transport{},
	}
	response, err := myClient.Get(url)
	if err != nil {
		fmt.Printf("HTTPDataSource Get() => %v", err.Error())
		return err
	}
	defer response.Body.Close()
	rspcontents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("HTTPDataSource ioutil.ReadAll() err: %s", err)
		return err
	}
	c <- rspcontents
	return nil
}

type RequestHandling struct {
	HTTPSDataSource
	CacheRatio    float64
	ServeByOrigin int
	Blocked       int
}

type Acceleration struct {
	HTTPSDataSource
	CachedHits   int
	TotalRequest int
}

type TopRequestCountryDataSource struct {
	HTTPSDataSource
}

// Every day summary, week trend, need 7 day data
type TrendHistogramDataSource struct {
	HTTPSDataSource
}

type DataCenterDataSource struct {
	HTTPSDataSource
}
type Nodes [][]int

type TopStoryReport struct {
	HTTPSDataSource
	Pageviews2           map[string]Nodes
	Visitors2            map[string]Nodes
	NetflowBandwidth     Nodes
	LiveThreatsChart     map[string]Nodes
	LiveReqsChart        map[string]Nodes
	LiveCacheChart       map[string]Nodes
	LiveLegitimatedChart map[string]Nodes
	LiveUpstreamChart    map[string]Nodes
}
