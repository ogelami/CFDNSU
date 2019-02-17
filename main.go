package main

import(
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"fmt"
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
	CheckRate int `json:"check_rate"`
}

/*
{
{"result":{"id":"4ed7c0c3f64f590405a877a23ec0427d",
"type":"A",
"name":"forwarddevelopment.se",
"content":"212.85.80.142",
"proxiable":true,
"proxied":true,
"ttl":1,
"locked":false,
"zone_id":"7d5a4e3e35b882ad128fa1952a1b7
e87",
"zone_name":"forwarddevelopment.se",
"modified_on":"2019-02-15T10:13:31.565456Z",
"created_on":"2019-02-15T10:13:31.565456Z",
"meta":{"auto_added":false,
"managed_by_apps":false,
"managed_by_argo_tunnel":false}},
"success":true,
"errors":[],
"messages":[]}
*/

type CFDNSRecordDetails struct {
	Result []struct {
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
	Errors []string `json:"errors"`
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
	resp, err := http.Get("https://icanhazip.com")

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

	log.Info(url)

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

	fmt.Printf("bf: %v", cFDNSRecordDetails)

	err = json.Unmarshal(body, &cFDNSRecordDetails)

	if err != nil {
		log.Error(err)
		log.Error(string(body))
	}

	fmt.Printf("ft: %v", cFDNSRecordDetails)

	return cFDNSRecordDetails
}

func main() {
	logging.SetFormatter(logging.MustStringFormatter(`%{color} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`))

/*	configurationRaw, err := ioutil.ReadFile(CONFIGURATION_PATH)

	if err != nil {
		log.Fatal(err)
		return
	}
*/

//	json.Unmarshal([]byte(configurationRaw), &bird)
//	fmt.Printf("Species: %s, Description: %s", bird.Species, bird.Description)

//	log.Info(string())
	c := getCFListDNSRecords(configuration.Records[0].ZoneIdentifier)

	for _, element := range c.Result {
    	fmt.Printf("\n=> %v\n", getCFDNSRecordDetails(configuration.Records[0].ZoneIdentifier, element.Id))
    	break
	}
//	log.Info(resolveIp())

/*
	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
	*/
}