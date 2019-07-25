package main

import (
	"../cfdnsu"
	"net"
	"os"
	"net/http"
	"net/http/fcgi"
	"encoding/json"
)

type s_configuration struct {
	FCGI struct {
		Protocol string `json:"protocol"`
		Listen string `json:"listen"`
	} `json:"fcgi"`
}

var (
	listen net.Listener
	configuration s_configuration
)

func Startup() error {
	err := json.Unmarshal(cfdnsu.SharedInformation.Configuration, &configuration)

	if err != nil {
		cfdnsu.SharedInformation.Logger.Error(err)
		return err
	}

	if configuration.FCGI.Protocol == "unix" {
		//cleanup if unix sockfile already exists
		if _, err = os.Stat(configuration.FCGI.Listen); err == nil {
			err = os.Remove(configuration.FCGI.Listen)

			if err != nil {
				cfdnsu.SharedInformation.Logger.Error(err)

				return err
			}
		}

		listen, err = net.Listen(configuration.FCGI.Protocol, configuration.FCGI.Listen)

		if err != nil {
			cfdnsu.SharedInformation.Logger.Error(err)

			return err
		}

		err = os.Chmod(configuration.FCGI.Listen, 0666)
	} else {
		listen, err = net.Listen(configuration.FCGI.Protocol, configuration.FCGI.Listen)
	}

	if err != nil {
		cfdnsu.SharedInformation.Logger.Error(err)

		return err
	}

	fastCGIServer := new(FastCGIServer)

	cfdnsu.SharedInformation.Logger.Infof("Serving %s", configuration.FCGI.Listen)

	go fcgi.Serve(listen, fastCGIServer)

	if err != nil {
		cfdnsu.SharedInformation.Logger.Errorf("%s", err)
	}

	return nil
}

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ip, port, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		cfdnsu.SharedInformation.Logger.Error(err)
	}

	cfdnsu.SharedInformation.Logger.Infof("%s:%s made an ip request", ip, port)
	w.Write([]byte(ip))
}
