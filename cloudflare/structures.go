package cloudflare

type Authentication struct{
	Email string `json:"email"`
	GlobalApiKey string `json:"global_api_key"`
	Token string `json:"token"`
}

type Record struct {
	ZoneIdentifier string `json:"zone_identifier"`
	Identifier string `json:"identifier"`
	Name string `json:"name"`
}

/*type RecordStack struct {
	ZoneIdentifier []struct {
		Identifier []struct {
			Name string `json:"name"`
		} `json:"identifier"`
	} `json:"zone_identifier"`
}*/

type DNSRecordDetails struct {
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

type ListDNSRecords struct {
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

type UpdateDNSRecord struct {
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

type ListZones struct {
	Success bool `json:"success"`
	Errors []struct {
		Code int `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Messages []string `json:"messages"`
	Result []struct {
		Name string `json:"name"`
		Id string `json:"id"`
	} `json:"result"`
	ResultInfo struct {
		Page int `json:"page"`
		PerPage int `json:"per_page"`
		Count int `json:"count"`
		TotalCount int `json:"Total_count"`
	} `json:"result_info`
}