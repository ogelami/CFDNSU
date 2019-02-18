package main

import(
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
    "net/http/fcgi"
    "net"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"math/rand"
)

const (
	CONFIGURATION_PATH = "config.json"
)

var log = logging.MustGetLogger("logger")
var configuration = loadConfiguration()

type Configuration struct {
	Auth struct {
		Email string `json:"email"`
		Key string `json:"key"`
	} `json:"auth"`
	Records []struct {
		ZoneIdentifier string `json:"zone_identifier"`
		Identifier string `json:"identifier"`
	} `json:"records"`
	Check struct {
		Rate int `json:"rate"`
		Targets []string `json:"targets"`
	} `json:"check"`
	FCGI struct {
		Protocol string `json:"protocol"`
		Listen string `json:"listen"`
	} `json:"fcgi"`
}

type CFDNSRecordDetails struct {
	Result struct {
		Id string `json:"id"`
		Type string `json:"type"`
		Name string `json:"name"`
		Content string `json:"content"`
		Proxiable bool `json:"proxiable"`
		Proxied bool `json:"proxied"`
		Ttl int `json:"ttl"`
		Locked bool `json:"locked"`
		Zone_id string `json:"zone_id"`
		Zone_name string `json:"zone_name"`
		Modified_on string `json:"modified_on"`
		Created_on string `json:"created_on"`
		Meta struct {
			Auto_added bool `json:"auto_added"`
			Managed_by_apps bool `json:"managed_by_apps"`
			Managed_by_argo_tunnel bool `json:"managed_by_argo_tunnel"`
		} `json:"meta"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors []struct {
		Code int `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Messages []string `json:"messages"`
}

type CFListDNSRecords struct {
	Success bool `json:"success"`
	Errors []string `json:"errors"`
	Messages []string `json:"messages"`
	Result []struct {
		Id string `json:"id"`
		Type string `json:"type"`
		Content string `json:"content"`
		Proxiable bool `json:"proxiable"`
		Proxied bool `json:"proxied"`
		Ttl int `json:"ttl"`
		Locked bool `json:"locked"`
		Zone_id string `json:"zone_id"`
		Zone_name string `json:"zone_name"`
		Created_on string `json:"created_on"`
		Modified_on string `json:"modified_on"`
		Data string `json:"data"`
	} `json:"result"`
	ResultInfo struct {
		Page int `json:"page"`
		Per_page int `json:"per_page"`
		Count int `json:"count"`
		Total_count int `json:"total_count"`
	} `json:"result_info"`
}

func loadConfiguration() Configuration {
	var configuration Configuration

	configurationRaw, err := ioutil.ReadFile(CONFIGURATION_PATH)

	if err != nil {
		log.Critical(err)
		return configuration
	}

	err = json.Unmarshal(configurationRaw, &configuration)

	if err != nil {
		log.Critical(err)
		return configuration
	}

	return configuration
}

func resolveIp() string {
	url := configuration.Check.Targets[rand.Intn(len(configuration.Check.Targets))]
	resp, err := http.Get(url)

	if err != nil {
		log.Error(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error(err)
		log.Error(string(body))
	}

	return string(body)
}

func getCFListDNSRecords(zoneIdentifier string) CFListDNSRecords {
	var cFListDNSRecords CFListDNSRecords
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100&type=A", zoneIdentifier)
	client := &http.Client{}
	
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Error(err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Auth-Email", configuration.Auth.Email)
	request.Header.Add("X-Auth-Key", configuration.Auth.Key)

	resp, err := client.Do(request)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error(err)
	}

	err = json.Unmarshal(body, &cFListDNSRecords)

	if err != nil {
		log.Error(err)
		log.Error(string(body))
	}

	return cFListDNSRecords
}

func getCFDNSRecordDetails(zoneIdentifier string, identifier string) CFDNSRecordDetails {
	var cFDNSRecordDetails CFDNSRecordDetails
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneIdentifier, identifier)
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Error(err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Auth-Email", configuration.Auth.Email)
	request.Header.Add("X-Auth-Key", configuration.Auth.Key)

	resp, err := client.Do(request)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error(err)
	}

	err = json.Unmarshal(body, &cFDNSRecordDetails)

	if err != nil {
		log.Error(err)
		log.Error(string(body))
	}

	return cFDNSRecordDetails
}

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
/*	fmt.Printf("\n=> %+v\n", s)
	fmt.Printf("\n=> %+v\n", w)
	fmt.Printf("\n=> %+v\n", req)

	fmt.Printf("\n=> %+v\n", req.RemoteAddr)*/

	ip, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		log.Error(err)
	}

	w.Write([]byte(ip))
}

func host() {
	listen, err := net.Listen(configuration.FCGI.Protocol, configuration.FCGI.Listen)

	if err != nil {
		log.Error(err)
	}

	fastCGIServer := new(FastCGIServer)

	/*@todo: check if configuration.FCGI.Listen is socket*/
	if configuration.FCGI.Protocol == "unix" {
		err = os.Chmod(configuration.FCGI.Listen, 0666)

		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Serving")

	fcgi.Serve(listen, fastCGIServer)
}

func main() {
	rand.Seed(time.Now().Unix())
	logging.SetFormatter(logging.MustStringFormatter(`%{color} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`))

	go host()
	time.Sleep(time.Second * 10)
	fmt.Printf("%+v", resolveIp())


/*	configurationRaw, err := ioutil.ReadFile(CONFIGURATION_PATH)

	if err != nil {
		log.Fatal(err)
		return
	}
*/

//	json.Unmarshal([]byte(configurationRaw), &bird)
//	fmt.Printf("Species: %s, Description: %s", bird.Species, bird.Description)

//	log.Info(string())
/*	c := getCFListDNSRecords(configuration.Records[2].ZoneIdentifier)

	for _, element := range c.Result {
    	fmt.Printf("\n=> %+v\n", getCFDNSRecordDetails(configuration.Records[2].ZoneIdentifier, element.Id))
    	break;
	}*/

/*	for _, record := range configuration.Records {
		fmt.Printf("\n=> %+v\n", getCFDNSRecordDetails(record.ZoneIdentifier, record.Identifier))
	}*/



//	log.Info(resolveIp())

/*
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
	*/
}