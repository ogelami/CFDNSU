package main

import(
	"github.com/op/go-logging"
	"github.com/alecthomas/kingpin"
	"io/ioutil"
	"net/http"
	"net"
	"encoding/json"
	"fmt"
	"os"
	"plugin"
	"strings"
	"time"
	"math/rand"
	"./cloudflare"
	"./cfdnsu"
	"syscall"
	"os/signal"
)

/**

make dep
go run -ldflags "-X main.CONFIGURATION_PATH=bin/cfdnsu.conf" main.go

* lets keep the CF api calls down by only calling getCFDNSRecordDetails on startup
* if no records are specify still make it possible to run the server as a fcgi only
* upon failure of retrieving the servers ip address, retry in the next cycle
* evaluate if its worth moving fcgi to its own package
* Shutdown event not firing
* ^C interrupt => runtimer error

* working test check "https://api.ipify.org/"
* ip resolution does not validate the ip addr comming back, if its html or ipv6 it will just be passed to cloudflare which will break.

*/

var CONFIGURATION_PATH string

func loadConfiguration() (error) {
	configurationRaw, err := ioutil.ReadFile(CONFIGURATION_PATH)

	if err != nil {
		log.Critical(err)
		return err
	}

	err = json.Unmarshal(configurationRaw, &cfdnsu.SharedInformation.Configuration)

	if err != nil {
		log.Critical(err)
		return err
	}

	cloudflare.Auth = cfdnsu.SharedInformation.Configuration.Auth
	cloudflare.Records = cfdnsu.SharedInformation.Configuration.Records

	return nil
}

func loadPlugins() {
	var fullPath string
	var symbol plugin.Symbol

	eventMap = map[string][]plugin.Symbol{}

	for _, record := range cfdnsu.SharedInformation.Configuration.Plugin.Load {
		fullPath = cfdnsu.SharedInformation.Configuration.Plugin.Path + "/" + record

		hotPlug, err := plugin.Open(fullPath)

		if err != nil {
			log.Critical(err)
			continue
		}

		for _, eventName := range []string{"Startup", "Shutdown", "IpChanged"} {
			symbol, err = hotPlug.Lookup(eventName)

			if err != nil {
				log.Critical(err)
				continue
			}

			eventMap[eventName] = append(eventMap[eventName], symbol)
		}
	}
}

func triggerEvent(eventName string) {
	if val, ok := eventMap[eventName]; ok {
		for _, element := range val {
			element.(func())()
		}
	}
}

func resolveIp() (error, string) {
	url := cfdnsu.SharedInformation.Configuration.Check.Targets[rand.Intn(len(cfdnsu.SharedInformation.Configuration.Check.Targets))]
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

	ip := strings.TrimSpace(string(body))

	if net.ParseIP(ip).To4() == nil {
		return fmt.Errorf("Not ipv4 coming from %s", url), ""
	}

	return nil, ip
}

var (
	kingpinApp = kingpin.New("CFDNSU", "Cloudflare DNS updater")
	kingpinDump = kingpinApp.Command("dump", "Dump zone_identifiers and identifiers")
	kingpinRun = kingpinApp.Command("run", "Run CFDNSU in foreground").Default()

	log = logging.MustGetLogger("logger")
	pluginMap map[string]*plugin.Plugin
	eventMap map[string][]plugin.Symbol
)

func dump() {
	err, cFListZones := cloudflare.GetCFListZones()

	if err != nil {
		log.Critical(err)

		return
	}

	if !cFListZones.Success {
		log.Critical(cFListZones.Errors)

		return
	}

	for _, zone := range cFListZones.Result {
		err, zoneRecord := cloudflare.GetCFListDNSRecords(zone.Id)

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
			err, zoneDetails := cloudflare.GetCFDNSRecordDetails(zone.Id, element.Id)

			if err != nil {
				log.Error(err)
			}

			modifier = "├"

			if iterator == zoneMax {
				modifier = "└"
			}

			log.Infof("%s─%s - %s", modifier, zoneDetails.Result.Id, zoneDetails.Result.Name)
		}
	}
}

func run() {
	cfdnsu.SharedInformation.Logger = log

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func(c chan os.Signal) {
		sig := <-c
		log.Infof("Caught signal %s: shutting down.", sig)
		os.Exit(0)
	}(sigc)

	oldIp := ""

	triggerEvent("Startup")

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

				for recordIndex, record := range cloudflare.Records {
					err, cFDNSRecordDetails := cloudflare.GetCFDNSRecordDetails(record.ZoneIdentifier, record.Identifier)

					if err != nil {
						log.Error(err)
						continue
					}

					if !cFDNSRecordDetails.Success {
						log.Error(cFDNSRecordDetails.Errors)
						continue
					}

					cloudflare.Records[recordIndex].Name = cFDNSRecordDetails.Result.Name

					if cFDNSRecordDetails.Result.Ip != currentIp {
						err, cCFDNSRecord := cloudflare.SetCFDNSRecord(recordIndex, currentIp)

						if err != nil {
							log.Error(err)
							log.Error(cCFDNSRecord)
							log.Error("Failed to update record")
							continue
						}

						if !cCFDNSRecord.Success {
							log.Error(cCFDNSRecord)
							continue
						}

						cfdnsu.SharedInformation.CurrentIp = currentIp
						triggerEvent("IpChanged")

						log.Infof("Server ip has changed to %s previously %s updating, updated %t", currentIp, cFDNSRecordDetails.Result.Ip, cCFDNSRecord.Success)
					}
				}
			}

			oldIp = currentIp
		}

		time.Sleep(time.Second * time.Duration(cfdnsu.SharedInformation.Configuration.Check.Rate))
	}
}

func main() {
	var err error
	rand.Seed(time.Now().Unix())
	logging.SetFormatter(logging.MustStringFormatter(`%{color} %{shortfunc} ▶ %{level:.4s} %{color:reset} %{message}`))

	err = loadConfiguration()
	loadPlugins()

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
