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
	"bytes"
)

/*
GOOS=linux GOARCH=arm go build main.go && scp {main,config.json,CFDNSU.service,install.sh} charon:
*/

/*
,
		{
			"zone_identifier" : "704c351a5412d4144779f57c789c5b1a",
			"identifier" : "4141e2721e41beed61097578301696e4",
			"name" : "charon.fwdev.se"
		}
*/

/*
"http://charon.fwdev.se/test"
*/

/*
@todo: graceful .sock shutdown on sigterm
	- https://stackoverflow.com/questions/16681944/how-to-reliably-unlink-a-unix-domain-socket-in-go-programming-language
*/

const (
	CONFIGURATION_PATH = "/etc/CFDNSU/config.json"
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
		Name string `json:"name"`
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
		Ip string `json:"content"`
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
	Errors []struct {
		Code int `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
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

type CFUpdateDNSRecord struct {
	Success bool `json:"success"`
	Errors []struct {
		Code int `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Messages []string `json:"messages"`
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
		Created_on string `json:"created_on"`
		Modified_on string `json:"modified_on"`
	} `json:"result"`
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

func resolveIp() (error, string) {
	url := configuration.Check.Targets[rand.Intn(len(configuration.Check.Targets))]
	resp, err := http.Get(url)

	if err != nil {
		log.Error(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error(err)
		log.Errorf("%s", body)
	}

	if resp.StatusCode > 200 {
		return fmt.Errorf("Wrong response code %d", resp.StatusCode), ""
	}

	return nil, string(body)
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
		log.Errorf("%s", body)
	}

	return cFDNSRecordDetails
}

func setCFDNSRecord(recordId int, ip string) bool {
	var cFUpdateDNSRecord CFUpdateDNSRecord
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", configuration.Records[recordId].ZoneIdentifier, configuration.Records[recordId].Identifier)
	client := &http.Client{}

	cFUpdateDNSRecordRequest := map[string]string{"type": "A", "name": configuration.Records[recordId].Name, "content": ip}

	jsonData, err := json.Marshal(cFUpdateDNSRecordRequest)

	if err != nil {
		log.Error(err)
	}

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

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

	err = json.Unmarshal(body, &cFUpdateDNSRecord)

	if err != nil {
		log.Error(err)
		log.Errorf("%s", body)
	}

	return cFUpdateDNSRecord.Success
}

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

	/* @todo: check if configuration.FCGI.Listen is socket */
	if configuration.FCGI.Protocol == "unix" {
		err = os.Chmod(configuration.FCGI.Listen, 0666)

		if err != nil {
			log.Error(err)
		}
	}

	log.Infof("Serving %s", configuration.FCGI.Listen)

	fcgi.Serve(listen, fastCGIServer)
}

func main() {
	rand.Seed(time.Now().Unix())
	logging.SetFormatter(logging.MustStringFormatter(`%{color} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`))

	if configuration.FCGI.Listen != "" && configuration.FCGI.Protocol != "" {
		go host()
	}

	var oldIp string

	for true {
		err, currentIp := resolveIp()

		if err != nil {
			log.Fatal(err)
		}

		if currentIp != oldIp {
			log.Infof("Current ip %s", currentIp)

			for recordIndex, record := range configuration.Records {
				cFDNSRecordDetails := getCFDNSRecordDetails(record.ZoneIdentifier, record.Identifier)

				if cFDNSRecordDetails.Result.Ip != currentIp {
					log.Infof("Server ip has changed to %s previously %s updating, updated %t", currentIp, cFDNSRecordDetails.Result.Ip, setCFDNSRecord(recordIndex, currentIp))	
				}
			}
		}

		oldIp = currentIp
		time.Sleep(time.Second * time.Duration(configuration.Check.Rate))
	}

//	json.Unmarshal([]byte(configurationRaw), &bird)
//	fmt.Printf("Species: %s, Description: %s", bird.Species, bird.Description)

//	log.Info(string())
/*	c := getCFListDNSRecords(configuration.Records[2].ZoneIdentifier)

	for _, element := range c.Result {
    	fmt.Printf("\n=> %+v\n", getCFDNSRecordDetails(configuration.Records[2].ZoneIdentifier, element.Id))
    	break;
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