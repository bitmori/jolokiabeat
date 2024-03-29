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

	if s.Username != "" || s.Password != "" {
		serverURL.User = url.UserPassword(s.Username, s.Password)
	}

	proxyURL, err := url.Parse("http://" + j.Proxy.Host + ":" + j.Proxy.Port + j.Context)
	if err != nil {
		return fmt.Errorf("Error when parsing proxy server url: %v", err)
	}

	if j.Proxy.Username != "" || j.Proxy.Password != "" {
		proxyURL.User = url.UserPassword(j.Proxy.Username, j.Proxy.Password)
	}

	fields := make(common.MapStr)
	for _, metric := range j.Metrics {
		measurement := metric.Name

		req, err := ego.prepareRequest(serverURL.String(), s, metric, j.Mode == "proxy", proxyURL.String())
		if err != nil {
			return err
		}

		out, err := ego.doRequest(req)

		if err != nil {
			fmt.Printf("Error handling response: %s\n", err)
		} else {
			if values, ok := out["value"]; ok {
				switch t := values.(type) {
				case map[string]interface{}:
					for k, v := range t {
						switch t2 := v.(type) {
						case map[string]interface{}:
							for k2, v2 := range t2 {
								fields[measurement+"_"+k+"_"+k2] = v2
							}
						case interface{}:
							fields[measurement+"_"+k] = t2
						}
					}
				case interface{}:
					fields[measurement] = t
				}
			} else {
				fmt.Printf("Missing key 'value' in output response\n")
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

func (ego *Jolokiabeat) prepareRequest(serverURL string, server Server, metric Metric, enableProxy bool, proxyURL string) (*http.Request, error) {
	jolokiaURL := serverURL

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

	if enableProxy {
		serviceURL := fmt.Sprintf("service:jmx:rmi:///jndi/rmi://%s:%s/jmxrmi", server.Host, server.Port)

		target := common.MapStr{
			"url": serviceURL,
		}

		if server.Username != "" {
			target["user"] = server.Username
		}

		if server.Password != "" {
			target["password"] = server.Password
		}

		bodyContent["target"] = target

		jolokiaURL = proxyURL
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
