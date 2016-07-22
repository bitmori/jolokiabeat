package beater

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
)

// CollectData -> get data from jolokia
func (ego *Jolokiabeat) CollectData(j Jok, s Server) error {
	serverURL, err := url.Parse("http://" + s.Host + ":" + s.Port + j.Context)
	if err != nil {
		return fmt.Errorf("Error when parsing server url: %v", err)
	}

	fields := make(common.MapStr)
	for _, metric := range j.Metrics {
		measurement := metric.Name

		req, err := ego.prepareRequest(serverURL.String(), s, metric)
		if err != nil {
			return err
		}

		out, err := ego.doRequest(req)

		if err != nil {
			fmt.Printf("Error handling response: %s\n", err)
		} else {
			if values, ok := out["value"]; ok {
				switch t := values.(type) {
				case common.MapStr:
					for k, v := range t {
						fields[measurement+"_"+k] = v
					}
				case interface{}:
					fields[measurement] = t
				}
			} else {
				fmt.Println("Missing key 'value' in output response.")
			}
		}
	}

	typ := "jolokia"

	if j.Type != "" {
		typ = j.Type
	} else if s.Type != "" {
		typ = s.Type
	}

	event := common.MapStr{
		"@timestamp": common.Time(time.Now()),
		"type":       typ,
		"context":    j.Context,
		"host":       s.Host,
		"port":       s.Port,
		"server":     s.Name,
	}
	if ego.metricUnderRoot {
		logp.Debug(selector, "metricUnderRoot = %d", ego.metricUnderRoot)
		ego.publisher.PublishEvent(common.MapStrUnion(event, fields))
	} else {
		event[ego.metricFieldName] = fields
		ego.publisher.PublishEvent(event)
	}
	return nil
}

func (ego *Jolokiabeat) doRequest(req *http.Request) (common.MapStr, error) {
	resp, err := ego.requester.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending http request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			req.RequestURI,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK),
		)
		return nil, err
	}

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal json
	var jsonOut common.MapStr
	if err = json.Unmarshal([]byte(body), &jsonOut); err != nil {
		return nil, errors.New("Error decoding JSON response")
	}

	if status, ok := jsonOut["status"]; ok {
		if status != float64(200) {
			return nil, fmt.Errorf("Not expected status value in response body: %3.f", status)
		}
	} else {
		return nil, fmt.Errorf("Missing status in response body")
	}

	return jsonOut, nil
}

func (ego *Jolokiabeat) prepareRequest(jolokiaURL string, server Server, metric Metric) (*http.Request, error) {
	bodyContent := common.MapStr{
		"type":  "read",
		"mbean": metric.Mbean,
	}

	if metric.Attribute != "" {
		bodyContent["attribute"] = metric.Attribute
		if metric.Path != "" {
			bodyContent["path"] = metric.Path
		}
	}

	requestBody, err := json.Marshal(bodyContent)
	if err != nil {
		return nil, fmt.Errorf("Error parsing MapStr to JSON: %v", err)
	}

	req, err := http.NewRequest("POST", jolokiaURL, bytes.NewBuffer(requestBody))

	if err != nil {
		return nil, fmt.Errorf("Error creating http post request: %v", err)
	}

	req.Header.Add("Content-type", "application/json")

	return req, nil
}
