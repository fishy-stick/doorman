package ddns

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	dnspodAPIBase = "https://dnsapi.cn"
	dnspodUA      = "Doorman/1.0 (doorman@github.com)"
)

type DNSPodConfig struct {
	Domain string `json:"domain"`
	Record string `json:"record"`
	ID     string `json:"id"`
	Token  string `json:"token"`
}

type DNSPodProvider struct{}

type dnspodResponse struct {
	Status struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
	Records []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"records"`
}

func (p *DNSPodProvider) Update(configStr string, ip string) (*UpdateResult, error) {
	var config DNSPodConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, fmt.Errorf("parse dnspod config: %w", err)
	}

	if config.Domain == "" || config.Record == "" || config.ID == "" || config.Token == "" {
		return nil, fmt.Errorf("dnspod config missing required fields")
	}

	currentIP, recordID, err := p.getRecordValue(config)
	if err != nil {
		return nil, fmt.Errorf("get current record: %w", err)
	}

	if currentIP == ip {
		return &UpdateResult{
			Updated:    false,
			OldIP:      currentIP,
			NewIP:      ip,
			Skipped:    true,
			SkipReason: "ip unchanged",
		}, nil
	}

	if err := p.updateRecord(config, recordID, ip); err != nil {
		return nil, fmt.Errorf("update record: %w", err)
	}

	return &UpdateResult{
		Updated: true,
		OldIP:   currentIP,
		NewIP:   ip,
	}, nil
}

func (p *DNSPodProvider) getRecordValue(config DNSPodConfig) (string, string, error) {
	data := url.Values{
		"login_token": {fmt.Sprintf("%s,%s", config.ID, config.Token)},
		"format":      {"json"},
		"domain":      {config.Domain},
		"sub_domain":  {config.Record},
		"record_type": {"A"},
	}

	resp, err := p.doRequest("POST", "/Record.List", data)
	if err != nil {
		return "", "", err
	}

	if resp.Status.Code != "1" {
		return "", "", fmt.Errorf("dnspod error: %s", resp.Status.Message)
	}

	for _, r := range resp.Records {
		if r.Name == config.Record && r.Type == "A" {
			return r.Value, r.ID, nil
		}
	}

	return "", "", nil
}

func (p *DNSPodProvider) updateRecord(config DNSPodConfig, recordID string, ip string) error {
	if recordID == "" {
		return p.createRecord(config, ip)
	}

	data := url.Values{
		"login_token": {fmt.Sprintf("%s,%s", config.ID, config.Token)},
		"format":      {"json"},
		"domain":      {config.Domain},
		"record_id":   {recordID},
		"sub_domain":  {config.Record},
		"record_type": {"A"},
		"record_line": {"默认"},
		"value":       {ip},
	}

	resp, err := p.doRequest("POST", "/Record.Ddns", data)
	if err != nil {
		return err
	}

	if resp.Status.Code != "1" {
		return fmt.Errorf("dnspod error: %s", resp.Status.Message)
	}

	return nil
}

func (p *DNSPodProvider) createRecord(config DNSPodConfig, ip string) error {
	data := url.Values{
		"login_token": {fmt.Sprintf("%s,%s", config.ID, config.Token)},
		"format":      {"json"},
		"domain":      {config.Domain},
		"sub_domain":  {config.Record},
		"record_type": {"A"},
		"record_line": {"默认"},
		"value":       {ip},
	}

	resp, err := p.doRequest("POST", "/Record.Create", data)
	if err != nil {
		return err
	}

	if resp.Status.Code != "1" {
		return fmt.Errorf("dnspod error: %s", resp.Status.Message)
	}

	return nil
}

func (p *DNSPodProvider) doRequest(method, path string, data url.Values) (*dnspodResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(method, dnspodAPIBase+path, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", dnspodUA)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result dnspodResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
