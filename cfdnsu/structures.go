package cfdnsu

import(
	"github.com/op/go-logging"
)

type sharedInformation struct {
	Logger *logging.Logger
	CurrentIp string
	Configuration []byte
}

/*type configuration struct {
	Auth cloudflare.Authentication `json:"auth"`
	Records []cloudflare.Record `json:"records"`
	Check struct {
		Rate int `json:"rate"`
		Targets []string `json:"targets"`
	} `json:"check"`
/*	FCGI struct {
		Protocol string `json:"protocol"`
		Listen string `json:"listen"`
	} `json:"fcgi"`
	Webserver struct {
		Protocol string `json:"protocol"`
		Listen string `json:"listen"`
		Certificate string `json:"certificate"`
		CertificateKey string `json:"certificate_key"`
	} `json:"webserver"`
	Plugin struct {
		Path string `json:"path"`
		Load []string `json:"load"`
	} `json:"plugin"`
}*/

var SharedInformation = sharedInformation{nil, "", nil}