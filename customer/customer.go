package customer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	Reader = iota
	Updater
)

var (

	// Gcenter Url
	//  https://gc.nexusguard.com/ssdbapi.php?cmd=proxyvip&param=Proxy-VIP

	gCenterUrl = "https://gc.nexusguard.com"

	// default sleep second to update with G center
	pollSeconds = 600 * time.Second

	// DEBUG
	//pollSeconds = 10 * time.Second

	tcpProxyConfig = "./tcpproxy.json"

	URLAllVips = "%s/ssdbapi.php?cmd=proxyvip&param=Proxy-VIP"
	URLVipData = "%s/ssdbapi.php?cmd=proxyvip&param=Proxy-%s&from=%d"
	URLAllSite = "%s/ssdbapi.php?cmd=list"

	// later will use this https://gc.nexusguard.com/ssdbapi.php?cmd=listvip

	webServiceType      = "web"
	tcpproxyServiceType = "tcp"

	// allcustomer key this is the very only one key for
	//store customer in redis
	allcustomerKey = "allcustomer"

	// our own httpclient
	HttpClient = http.Client{Timeout: time.Second * 8}

	idcMap = map[string]string{
		"27.126.193.126": "HongKong",
		"27.126.197.126": "Miami",
		"27.126.198.126": "SanJose",
		"113.21.223.28":  "Taipei",
		"192.1.1.2":      "London",
		"27.126.253.126": "LosAngeles",
		"27.126.254.126": "Washington",
		"27.126.255.126": "Singapore",
	}
)

//AllSite  is struct for gcenter api
type AllSite struct {
	Idc         string `json:"idc"`
	Sites       string `json:"sites"`
	Sitesconfig string `json:"sitesconfig"`
}

// Site struct for g center api
type Customer struct {
	CId   string `json:"CId"`
	CName string `json:"CName"`
	Sites []Site `json:"Sites"`
}

type Site struct {
	BP       string `json:"BP"`
	SerId    int    `json:"SerId"`
	Site     string `json:"Site"`
	SiteName string `json:"SiteName"`
	Status   string `json:"Status"`
	VIP      string `json:"VIP"`
}

func (all *AllSite) parseCustomers() []Customer {
	var customers []Customer
	json.Unmarshal([]byte(all.Sitesconfig), &customers)
	return customers
}

// GetVips split the vips string with comma split
func (s *Site) GetVips() []string {
	return strings.Split(s.VIP, ",")
}

// CWCustomer struct for clearwatch
// TODO should we put this to schema ?
type CWCustomer struct {
	Id    string    `json:"id"`
	Name  string    `json:"name"`
	Sites []*CWSite `json:"sites"`
}

// CWSite struct for clearwatch
type CWSite struct {
	CID         string   `json:"cid"`
	SiteID      string   `json:"siteid"`
	SiteName    string   `json:"sitename"`
	BP          bool     `json:"bp"`
	VIPS        []string `json:"vip"`
	ServiceType string   `json:"servicetype"`
}

func (s *CWSite) IsTCPProxy() bool {
	if s.ServiceType == "tcp" {
		return true
	}
	return false
}

type CustomerService struct {
	Customers   map[string]*CWCustomer
	Sites       map[string]*CWSite
	VipMap      map[string]string
	IDCMap      map[string]string
	gcenterUrl  string
	servicetype int
	store       *redis.Pool
	Timestamp   time.Time
	sync.RWMutex
}

////////////////////////////////////////
// FIXME THIS MUST REMOVE for TCPProxy

func loadTCPProxyConfig() map[string][]string {
	var proxyconfig = make(map[string][]string)
	tcpproxybytes, err := ioutil.ReadFile(tcpProxyConfig)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Going to load ./tcpproxy.json file to config TCPProxy site")
		err = json.Unmarshal(tcpproxybytes, &proxyconfig)
	}
	return proxyconfig
}

// End of FIXME
////////////////////////////////////////

func NewCustomerService(redisPool *redis.Pool, servicetype int, gcurl string) (*CustomerService, error) {
	cs := new(CustomerService)

	cs.store = redisPool

	// check service type
	if servicetype != Reader && servicetype != Updater {
		return nil, errors.New("service must be customer.Reader or customer.Updater")
	}

	if gcurl == "" {
		cs.gcenterUrl = gCenterUrl
	} else {
		cs.gcenterUrl = gcurl
	}

	cs.servicetype = servicetype

	cs.IDCMap = idcMap

	// Reader
	if servicetype == Reader {
		go func() {
			for {
				cs.loadAllCustomer()
				time.Sleep(pollSeconds)
			}
		}()
	}

	// Updater
	if servicetype == Updater {
		go func() {
			cs.loadAllCustomer()
			for {
				cs.syncWithGCenter()
				time.Sleep(pollSeconds)
				//time.Sleep(10 * time.Second)
			}
		}()
	}

	return cs, nil
}

// syncWithGCenter will request all site infomation from gcenter then update current config
func (s *CustomerService) syncWithGCenter() {
	var body []byte
	var allsite AllSite

	// transform to ClearWatch Customer
	var allcustomer = make(map[string]*CWCustomer)
	log.Println(s.gcenterUrl)

	url := fmt.Sprintf(URLAllSite, s.gcenterUrl)

	response, err := HttpClient.Get(url)

	if err == nil {
		defer response.Body.Close()
		body, err = ioutil.ReadAll(response.Body)
	} else {
		log.Println(err)
		// G center fail stop here, just use data from redis
		return
	}

	json.Unmarshal(body, &allsite)
	customers := allsite.parseCustomers()

	for _, val := range customers {
		// new ?
		u := &CWCustomer{val.CId, val.CName, []*CWSite{}}

		for _, site := range val.Sites {
			var bp bool
			var serviceType = webServiceType

			if site.BP == "on" {
				bp = true
			}

			// this site is active
			if site.Status == "1" {
				u.Sites = append(u.Sites, &CWSite{val.CId, site.Site, site.SiteName, bp, site.GetVips(), serviceType})
			}
		}

		// having at least one site active
		if len(u.Sites) > 0 {
			allcustomer[u.Id] = u
		}

		// compare with Context
	}

	// FIXME, add hard code TCP Proxy site
	// compare from allcustomer
	proxyconfig := loadTCPProxyConfig()

	for _, val := range allcustomer {

		lines, ok := proxyconfig[val.Id]

		if ok {
			for _, line := range lines {
				vipSlice := strings.Split(line, ",")
				// line like "27.126.219.21,S-669eba1a-11d5-4382-8b1a-e59b86d6bd20"
				// vipSlice like ["27.126.219.21", "S-669eba1a-11d5-4382-8b1a-e59b86d6bd20"]
				if len(vipSlice) == 2 {
					for _, site := range val.Sites {
						if site.SiteID == vipSlice[1] {
							// servicetype overwrite
							site.ServiceType = tcpproxyServiceType
						}
					}
				}

			}
		}

	}

	// compare from proxyconfig
	// not found add manual from proxyconfig

	var cidSorted []string
	for k := range proxyconfig {
		cidSorted = append(cidSorted, k)
	}

	// keep Sorted
	sort.Strings(cidSorted)

	for _, custId := range cidSorted {

		if _, ok := allcustomer[custId]; !ok {
			// manual add customer
			// line like "27.126.219.21,S-669eba1a-11d5-4382-8b1a-e59b86d6bd20"
			// vipSlice like ["27.126.219.21", "S-669eba1a-11d5-4382-8b1a-e59b86d6bd20"]

			siteMap := make(map[string][]string)
			lines := proxyconfig[custId]

			for _, line := range lines {
				vipSlice := strings.Split(line, ",")
				if len(vipSlice) == 2 {
					vip, siteid := vipSlice[0], vipSlice[1]
					siteMap[siteid] = append(siteMap[siteid], vip)
				}
			}

			// fill the site
			var siteSorted []string

			for key, _ := range siteMap {
				siteSorted = append(siteSorted, key)
			}

			sort.Strings(siteSorted)

			if len(siteSorted) > 0 {
				u := &CWCustomer{custId, "NoName", []*CWSite{}}
				for _, siteid := range siteSorted {
					u.Sites = append(u.Sites, &CWSite{custId, siteid, "NoName", false, siteMap[siteid], tcpproxyServiceType})
				}

				allcustomer[custId] = u
			}
		}
	}
	// end of FIXME

	// compare with original customer infomation
	if !reflect.DeepEqual(s.Customers, allcustomer) {
		log.Println("Update Config")
		s.Lock()
		defer s.Unlock()

		s.Customers = allcustomer
		s.Timestamp = time.Now()
		s.updateSitesAndVipMap()
		s.Sites, s.VipMap = s.updateSitesAndVipMap()
		// update customers, write this update to redis
		s.updateAllCustomer()
	} else {
		log.Println("Same Config")
	}
}

// loadAllCustomer load customer from redis
func (s *CustomerService) loadAllCustomer() {
	var customerForLoad = make(map[string]*CWCustomer)
	var err error

	conn := s.store.Get()
	defer conn.Close()
	values, err := redis.Bytes(conn.Do("GET", allcustomerKey))
	// handle successful return
	if err == nil {
		//json.Unmarshal(values, &s.Customers)
		err = json.Unmarshal(values, &customerForLoad)

		if err != nil {
			// stop here
			log.Println(err)
			return
		}

		if !reflect.DeepEqual(s.Customers, customerForLoad) {
			s.Lock()
			defer s.Unlock()
			log.Println("Update Config from last read")
			s.Customers = customerForLoad
			s.Sites, s.VipMap = s.updateSitesAndVipMap()
			s.Timestamp = time.Now()
		} else {
			log.Println("Same Config from last read")
		}
		log.Println("Loaded AllCustomer")
	} else {
		log.Println(err)
	}

}

// dump curreent PorterContext.customerSlice to redis
func (s *CustomerService) updateAllCustomer() {
	conn := s.store.Get()
	defer conn.Close()
	bytes, err := json.Marshal(s.Customers)

	if err == nil {
		log.Println("updateAllCustomer")
		conn.Do("SET", allcustomerKey, bytes)
	} else {
		log.Println("updateAllCustomer error", err)
	}
}

// cook Sites && VipMap this action must protect via mutex Lock
func (s *CustomerService) updateSitesAndVipMap() (map[string]*CWSite, map[string]string) {

	// create new variables
	sites := make(map[string]*CWSite)
	vipmap := make(map[string]string)
	for _, c := range s.Customers {
		for _, site := range c.Sites {
			sites[site.SiteID] = site
			for _, vip := range site.VIPS {
				vipmap[vip] = site.SiteID
			}
		}
	}
	log.Println("update Sites and VipMap")
	return sites, vipmap
}

// help function for CustomerService
// SiteIds return a slice with all siteids
func (s *CustomerService) SiteIds() []string {
	var siteids []string
	for k, _ := range s.Sites {
		siteids = append(siteids, k)
	}
	return siteids
}

// Vips return a slice with all vips
func (s *CustomerService) Vips() []string {
	var vips []string
	for k, _ := range s.VipMap {
		vips = append(vips, k)
	}
	return vips
}

// GetVipMap return a copy of VipMap
func (s *CustomerService) GetVipMap() map[string]string {
	s.RLock()
	defer s.RUnlock()
	return s.VipMap
}

// GetSites return a copy of Sites
func (s *CustomerService) GetSites() map[string]*CWSite {
	s.RLock()
	defer s.RUnlock()
	return s.Sites
}

func (s *CustomerService) GetIDCMap() map[string]string {
	return s.IDCMap
}

// InfoHandler http handler for CustomerServiee
func (s *CustomerService) InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(s.Customers)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(result)
}

// AllInfoHandler http handler for CustomerServiee
func (s *CustomerService) AllInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(s)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(result)
}

// VipsHandler http handler for CustomerServiee
func (s *CustomerService) VipsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(s.GetVipMap())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(result)
}

// SitesHandler http handler for CustomerServiee
func (s *CustomerService) SitesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	result, err := json.Marshal(s.GetSites())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(result)
}

// TimestampHandler http handler for CustomerServiee
func (s *CustomerService) TimestampHandler(w http.ResponseWriter, r *http.Request) {
	result := make(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")
	result["time"] = s.Timestamp
	body, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(body)
}
