package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// NginxClient allows you to fetch NGINX metrics from the stub_status page.
type KongClient struct {
	apiEndpoint string
	httpClient  *http.Client
}

type StubConnections struct {
	Database struct {
		Reachable bool `json:"reachable"`
	} `json:"database"`
	Server struct {
		ConnectionsAccepted float64 `json:"connections_accepted"`
		ConnectionsActive   float64 `json:"connections_active"`
		ConnectionsHandled  float64 `json:"connections_handled"`
		ConnectionsReading  float64 `json:"connections_reading"`
		ConnectionsWaiting  float64 `json:"connections_waiting"`
		ConnectionsWriting  float64 `json:"connections_writing"`
		TotalRequests       float64 `json:"total_requests"`
	} `json:"server"`
}

func NewKongClient(httpClient *http.Client, apiEndpoint string) (*KongClient, error) {
	client := &KongClient{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
	}
	_, err := client.GetStubStats()
	return client, err
}

func (client *KongClient) GetStubStats() (*StubConnections, error) {
	resp, err := client.httpClient.Get(client.apiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get %v: %v", client.apiEndpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected %v response, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %v", err)
	}

	var stats StubConnections
	err = parseStubStats(body, &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body %q: %v", string(body), err)
	}

	return &stats, nil
}

func parseStubStats(data []byte, stats *StubConnections) error {
	err := json.Unmarshal(data, stats)
	if err != nil {
		return err
	}
	return nil
}
