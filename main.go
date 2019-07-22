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
	"strings"
	"time"
	"math/rand"
	"./cloudflare"
	"syscall"
	"os/signal"
)

/**

make dep
go run -ldflags "-X main.CONFIGURATION_PATH=bin/cfdnsu.conf" main.go

* lets keep the CF api calls down by only calling getCFDNSRecordDetails on startup
* if no records are specify still make it possible to run the server as a fcgi only
* upon failure of retrieving the servers ip address, retry in the next cycle
* fcgi doesn't seem to respond properly on ?windows?
* evaluate if its worth moving fcgi to its own package

* working test check "https://api.ipify.org/"
* ip resolution does not validate the ip addr comming back, if its html or ipv6 it will just be passed to cloudflare which will break.

*/

type Configuration struct {
	Auth cloudflare.Authentication `json:"auth"`
	Records []cloudflare.Record `json:"records"`
	Check struct {
		Rate int `json:"rate"`
		Targets []string `json:"targets"`
	} `json:"check"`
	FCGI struct {
		Protocol string `json:"protocol"`
		Listen string `json:"listen"`
	} `json:"fcgi"`
}

var CONFIGURATION_PATH string

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

	cloudflare.Auth = configuration.Auth
	cloudflare.Records = configuration.Records

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

	ip := strings.TrimSpace(string(body))

	if net.ParseIP(ip).To4() == nil {
		return fmt.Errorf("Not ipv4 coming from %s", url), ""
	}

	return nil, ip
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
	err, cFListZones := cloudflare.GetCFListZones(configuration.Auth)

	if err != nil {
		log.Error(err)
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

						log.Infof("Server ip has changed to %s previously %s updating, updated %t", currentIp, cFDNSRecordDetails.Result.Ip, cCFDNSRecord.Success)
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
