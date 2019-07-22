package cloudflare

import(
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"bytes"
)

var Auth Authentication
var Records []Record

func request(requestType string, requestUrl string, requestData []byte, structure interface{}) error {
	client := &http.Client{}

	request, err := http.NewRequest(requestType, requestUrl, bytes.NewBuffer(requestData))

	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	if Auth.Key != "" {
		request.Header.Add("X-Auth-Email", Auth.Email)
		request.Header.Add("X-Auth-Key", Auth.Key)
	} else if Auth.Token != "" {
		request.Header.Add("Authorization", "Bearer " + Auth.Token)
	} else {
		return fmt.Errorf("No tokens/api keys supplied in config.")
	}

	resp, err := client.Do(request)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &structure)

	if err != nil {
		return err
	}

	return nil
}

func GetCFListZones(auth Authentication) (error, ListZones) {
	var listZones ListZones
	url := "https://api.cloudflare.com/client/v4/zones?per_page=50"
	err := request("GET", url, nil, &listZones)

	return err, listZones
}

func GetCFListDNSRecords(zoneIdentifier string) (error, ListDNSRecords) {
	var listDNSRecords ListDNSRecords
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=100&type=A", zoneIdentifier)
	err := request("GET", url, nil, &listDNSRecords)

	return err, listDNSRecords
}

func GetCFDNSRecordDetails(zoneIdentifier string, identifier string) (error, DNSRecordDetails) {
	var dNSRecordDetails DNSRecordDetails
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneIdentifier, identifier)
	err := request("GET", url, nil, &dNSRecordDetails)

	return err, dNSRecordDetails
}

func SetCFDNSRecord(recordId int, ip string) (error, UpdateDNSRecord) {
	var updateDNSRecord UpdateDNSRecord
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", Records[recordId].ZoneIdentifier, Records[recordId].Identifier)

	cFUpdateDNSRecordRequest := map[string]string{"type": "A", "name": Records[recordId].Name, "content": ip}

	jsonData, err := json.Marshal(cFUpdateDNSRecordRequest)

	if err != nil {
		return err, updateDNSRecord
	}

	err = request("PUT", url, jsonData, &updateDNSRecord)

	return err, updateDNSRecord
}
