package main

import (
	"github.com/abh/geoip"
	"sync"
	"fmt"
)



type DCObject struct {
	CenterName string
	CenterCount int
}
type AllSite struct{
	List map[string]map[string][]DCObject
}
type DnsSite struct{
	List map[string]string
}


type Command struct {
	Time  int64
	Cmd   string
	Param string
}

type cfgObject struct {
	GetReport struct {
		CheckTime string
		IntervalSeconds int 
		To []string
	}
	DnsCheck struct {
		IntervalSeconds int 
		To []string
		FilterCustomer []string
	}
	System struct {
		Key string
		Syslog string
	}
	CheckCacheRatio struct {
		IntervalSeconds int 
		CacheRatioBound int
		To []string
		MonitorList []string
	}
	CheckVariation struct {
		CheckTime string
		To []string
	}
	MonitorCustomerSite struct {
		IntervalSeconds int 
		To []string
	}
	Monitorg2 struct {
		IntervalSeconds int 
		SmtpServer string
		Port string
		From string
		Site []string
		To []string
	}
	Mail struct {
		SmtpServer string
		Port string
		From string
	}
	G2Server map[string]*struct {
		To []string
	}
	MonitorDC struct {
		MonitorList []string
		IntervalSeconds int
		To []string
	}
	MonitorBand struct {
		MonitorList []string
		IntervalSeconds int 
		SmtpServer string
		Port string 
		From string
		To []string
	}
	Monitor struct {
		FilterList string
	}
	Gen struct {		
		Http      int		
		Dev       bool
		Debug     bool
		GCenter   string
		GCenterPrd string
		Timeout int
		Sync	int64
	}
	Db struct {
		Username string
		Password string
		Hostname string
		Database string
		Cron     string		
		BackupTable []string
		ClearTable []string
		SaveTableDays []int
	}
	Tracker struct {		
		DayCron  string
		HourCron string
		MinCron string
		TwoMinCron string
		TenMinCron string
		SixHourCron string
		LiveDataStoreExpired int
		Collector []string
	}
	Waf struct {
		LiveDataStoreNum int
	}
}

type Customers struct {
	mu   *sync.Mutex
	List []customerObject
}

type customerObject struct {
	MoAlias	       string
	SiteAliasList  []string
	MoId           string
	CIdList        []int64
	DomainList     []string
	SiteList       []string
	ModuleList     []string
	ModuleMd5List  []string
	UpstreamList   []string
	SiteModuleList []map[string]map[string]string
	SiteHttpList   []string
	SiteHttpsList   []string
}

//Tracker
type AggregateList struct{	
	List map[string]map[string]map[string]int64
	UrlList map[string]map[string]map[string]int64
}

type JsonTrackerList struct {
	Single map[string]int64
	Mutli map[string]map[string]int64
}

type StringSort struct {
	Name string
	Value  int64
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type StringSortList []StringSort
func (p StringSort) String() string {
	return fmt.Sprintf("%s: %d", p.Name, p.Value)
}
func (a StringSortList) Len() int           { return len(a) }
func (a StringSortList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StringSortList) Less(i, j int) bool { return a[i].Value < a[j].Value }

type TrackerRecord struct{
	Time int64
	List map[string]int64
}

type TrackerObject struct {
	CId string
	SId string
	FirstViewTs int64//first view timestamp
	UserId string
	Bot string
	Search string
	Views int
	Url string
	Referer string
	Traffic string
	Ms int
	Sec int
	Min int
	Hour int
	Ts int64 // this time view timestamp
	Window string //borwser window size
	NewVisitor bool
	Browser string
	ClientIP string
	Exists bool
	Online bool
	Device string
	GeoInfo *geoip.GeoIPRecord
}


//WAF
type CheckCountData struct {
	SiteId	string
	Module	string
	Mode	string
	Info	*CountInfo
}
type CountInfo struct {
	Id	string
	Client	string
	GeoInfo	*geoip.GeoIPRecord
	Host	string
	Data	[]string
}
