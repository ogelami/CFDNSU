package main

import(
	"github.com/op/go-logging"
	"github.com/alecthomas/kingpin"
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
	"CFDNSU/cloudflare"
	"syscall"
	"os/signal"
)

/*
GOOS=linux GOARCH=arm go build main.go && scp {main,config.json,CFDNSU.service,install.sh} charon:
*/

type Auth struct{
	Email string `json:"email"`
	Key string `json:"key"`
}

type Configuration struct {
	Auth Auth `json:"auth"`
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

const (
	CONFIGURATION_PATH = "/etc/CFDNSU/config.json"
//	CONFIGURATION_PATH = "config.json"
)

func loadConfiguration() (error, Configuration) {
	configurationRaw, err := ioutil.ReadFile(CONFIGURATION_PATH)

	if err != nil {
		log.Critical(err)
		return err, configuration
	}

	err = json.Unmarshal(configurationRaw, &configuration)

	if err != nil {
		log.Critical(err)
		return err, configuration
	}

	return nil, configuration
}

func resolveIp() (error, string) {
	url := configuration.Check.Targets[rand.Intn(len(configuration.Check.Targets))]
	resp, err := http.Get(url)

	if err != nil {
		return err, ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err, ""
	}

	if resp.StatusCode > 200 {
		return fmt.Errorf("Wrong response code %d", resp.StatusCode), ""
	}

	return nil, string(body)
}

func getCFListZones(auth Auth) (error, cloudflare.CFListZones) {
	var cFListZones cloudflare.CFListZones
	url := "https://api.cloudflare.com/client/v4/zones?per_page=50"
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err, cFListZones
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Auth-Email", auth.Email)
	request.Header.Add("X-Auth-Key", auth.Key)

	resp, err := client.Do(request)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err, cFListZones
	}

	err = json.Unmarshal(body, &cFListZones)

	if err != nil {
		log.Error(string(body))
		return err, cFListZones
	}

	return nil, cFListZones
}

func getCFListDNSRecords(zoneIdentifier string) (error, cloudflare.CFListDNSRecords) {
	var cFListDNSRecords cloudflare.CFListDNSRecords
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100&type=A", zoneIdentifier)
	client := &http.Client{}
	
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err, cFListDNSRecords
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Auth-Email", configuration.Auth.Email)
	request.Header.Add("X-Auth-Key", configuration.Auth.Key)

	resp, err := client.Do(request)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err, cFListDNSRecords
	}

	err = json.Unmarshal(body, &cFListDNSRecords)

	if err != nil {
		log.Error(string(body))
		return err, cFListDNSRecords
	}

	return nil, cFListDNSRecords
}

func getCFDNSRecordDetails(zoneIdentifier string, identifier string) cloudflare.CFDNSRecordDetails {
	var cFDNSRecordDetails cloudflare.CFDNSRecordDetails
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
	var cFUpdateDNSRecord cloudflare.CFUpdateDNSRecord
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
	ip, port, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		log.Error(err)
	}

	log.Infof("%s:%s made an ip request", ip, port)
	w.Write([]byte(ip))
}

func host() (error, net.Listener) {
	var (
		err error
		listen net.Listener
	)

	if configuration.FCGI.Protocol == "unix" {
		//cleanup if unix sockfile already exists
		if _, err = os.Stat(configuration.FCGI.Listen); err == nil {
			err = os.Remove(configuration.FCGI.Listen)

			if err != nil {
				log.Error(err)

				return err, nil
			}
		}

		listen, err = net.Listen(configuration.FCGI.Protocol, configuration.FCGI.Listen)

		if err != nil {
			log.Error(err)

			return err, nil
		}

		err = os.Chmod(configuration.FCGI.Listen, 0666)
	} else {
		listen, err = net.Listen(configuration.FCGI.Protocol, configuration.FCGI.Listen)
	}

	if err != nil {
		log.Error(err)

		return err, nil
	}

	fastCGIServer := new(FastCGIServer)

	log.Infof("Serving %s", configuration.FCGI.Listen)

	go fcgi.Serve(listen, fastCGIServer)

	return nil, listen
}

var (
	kingpinApp = kingpin.New("CFDNSU", "Cloudflare DNS updater")
	kingpinDump = kingpinApp.Command("dump", "Dump zone_identifiers and identifiers")
	kingpinRun = kingpinApp.Command("run", "run").Default()

	log = logging.MustGetLogger("logger")
	configuration Configuration
)

func dump() {
	err, cFListZones := getCFListZones(configuration.Auth)

	if err != nil {
		log.Error(err)
	}

	for _, zone := range cFListZones.Result {
		err, zoneRecord := getCFListDNSRecords(zone.Id)

		if err != nil {
			log.Error(err)
		}

		if zoneRecord.Success == false {
			log.Errorf("%+v", zone)
		}

		zoneMax := len(zoneRecord.Result) - 1

		modifier := "┌"

		if zoneMax == -1 {
			modifier = " "
		}

		log.Infof("%s[%s][%s]", modifier, zone.Id, zone.Name)

		for iterator, element := range zoneRecord.Result {
			zoneDetails := getCFDNSRecordDetails(zone.Id, element.Id)

			modifier = "├"

			if iterator == zoneMax {
				modifier = "└"
			}

			log.Infof("%s─%s - %s", modifier, zoneDetails.Result.Id, zoneDetails.Result.Name)
		}
	}
}

func run() {
	if configuration.FCGI.Listen != "" && configuration.FCGI.Protocol != "" {
		err, listener := host()

		if err != nil {
			log.Errorf("%s", err)
		}

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)

		go func(c chan os.Signal) {
			sig := <-c
			log.Infof("Caught signal %s: shutting down.", sig)
			listener.Close()
			os.Exit(0)
		}(sigc)
	}

	oldIp := ""

	for true {
		err, currentIp := resolveIp()

		if err != nil {
			if oldIp == "" {
				log.Fatal(err)
				return
			} else {
				log.Warning(err)
			}
		} else {
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
		}

		time.Sleep(time.Second * time.Duration(configuration.Check.Rate))
	}
}

func main() {
	var err error
	rand.Seed(time.Now().Unix())
	logging.SetFormatter(logging.MustStringFormatter(`%{color} %{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`))

	err, configuration = loadConfiguration()

	if err != nil {
		log.Criticalf("%s", err)
		return
	}

	switch kingpin.MustParse(kingpinApp.Parse(os.Args[1:])) {
	case kingpinDump.FullCommand():
		dump()
	case kingpinRun.FullCommand():
		run()
	}
}