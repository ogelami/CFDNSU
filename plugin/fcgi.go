package main

import (
	"../cfdnsu"
	"net"
	"os"
	"net/http"
	"net/http/fcgi"
)

func Startup() {
	if cfdnsu.SharedInformation.Configuration.FCGI.Listen != "" && cfdnsu.SharedInformation.Configuration.FCGI.Protocol != "" {
		err, _ := host()

		if err != nil {
			cfdnsu.SharedInformation.Logger.Errorf("%s", err)
		}
	}
}

func host() (error, net.Listener) {
	var (
		err error
		listen net.Listener
	)

	if cfdnsu.SharedInformation.Configuration.FCGI.Protocol == "unix" {
		//cleanup if unix sockfile already exists
		if _, err = os.Stat(cfdnsu.SharedInformation.Configuration.FCGI.Listen); err == nil {
			err = os.Remove(cfdnsu.SharedInformation.Configuration.FCGI.Listen)

			if err != nil {
				cfdnsu.SharedInformation.Logger.Error(err)

				return err, nil
			}
		}

		listen, err = net.Listen(cfdnsu.SharedInformation.Configuration.FCGI.Protocol, cfdnsu.SharedInformation.Configuration.FCGI.Listen)

		if err != nil {
			cfdnsu.SharedInformation.Logger.Error(err)

			return err, nil
		}

		err = os.Chmod(cfdnsu.SharedInformation.Configuration.FCGI.Listen, 0666)
	} else {
		listen, err = net.Listen(cfdnsu.SharedInformation.Configuration.FCGI.Protocol, cfdnsu.SharedInformation.Configuration.FCGI.Listen)
	}

	if err != nil {
		cfdnsu.SharedInformation.Logger.Error(err)

		return err, nil
	}

	fastCGIServer := new(FastCGIServer)

	cfdnsu.SharedInformation.Logger.Infof("Serving %s", cfdnsu.SharedInformation.Configuration.FCGI.Listen)

	go fcgi.Serve(listen, fastCGIServer)

	return nil, listen
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
