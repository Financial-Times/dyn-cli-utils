package dyn

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

type gslbService struct {
	TTL int `json:"ttl"`
}

type gslbRegion struct {
	ServeCount   int    `json:"serve_count"`
	FailoverMode string `json:"failover_mode"`
	FailoverData string `json:"failover_data"`
	RegionCode   string `json:"region_code"`
	Pool         []struct {
		Address   string `json:"address"`
		Label     string `json:"label"`
		Weight    int    `json:"weight"`
		ServeMode string `json:"serve_mode"`
	} `json:"pool"`
}

type httpResponse struct {
	url      string
	response map[string]interface{}
	err      error
}

type GslbConfig struct {
	LabelPattern *regexp.Regexp
	ServeMode    string
	Fqdn         string
}

func (d *dynectService) doGslbUpdate(url string, payload gslbRegion, ch chan<- *httpResponse) {
	logger := log.WithFields(log.Fields{
		url: url,
	})
	logger.WithFields(log.Fields{
		"event": "UPDATING_GSLB_REGION",
	}).Info("Updating GSLB region")
	logger.WithFields(log.Fields{
		"payload": payload,
	}).Debug("GSLB PUT payload")

	response := struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}{}
	err := d.client.Do("PUT", url, payload, &response)
	if err == nil && response.Status != "success" {
		err = errors.New("Dynect API call status was no success")
	}
	if err != nil {
		logger.WithFields(log.Fields{
			"event":    "DYNECT_API_GSLB_REGION_PUT_FAILURE",
			"err":      err,
			"response": response,
		}).Error("Failed to put gslb region with disabled pools")
		ch <- &httpResponse{
			url: url,
			err: err,
		}
		return
	}
	logger.WithFields(log.Fields{
		"response": response.Data,
	}).Debug("GSLB PUT response")
	ch <- &httpResponse{
		url:      url,
		response: response.Data,
	}
}

func (d *dynectService) UpdateGslbRegion(config GslbConfig) error {
	allGslbUrl := fmt.Sprintf("GSLBRegion/%s/%s/?detail=Y", url.PathEscape("ft.com"), url.PathEscape(config.Fqdn))
	logger := log.WithFields(log.Fields{
		"url": allGslbUrl,
	})

	getResponse := struct {
		Status string       `json:"status"`
		Data   []gslbRegion `json:"data"`
	}{}
	err := d.client.Do("GET", allGslbUrl, nil, &getResponse)
	if err == nil && getResponse.Status != "success" {
		err = errors.New("Dynect API call status was no success")
	}
	if err != nil {
		logger.WithFields(log.Fields{
			"event":    "DYNECT_API_GSLB_REGION_GET_FAILURE",
			"err":      err,
			"response": getResponse,
		}).Error("Failed to get gslb region")
		return err
	}
	data := getResponse.Data
	logger.WithFields(log.Fields{
		"response": data,
	}).Debug("GSLB GET response")

	putRequests := make(map[string]gslbRegion)

	for _, gslb := range data {
		for index, pool := range gslb.Pool {
			if !config.LabelPattern.MatchString(pool.Label) {
				log.WithFields(log.Fields{
					"region":    gslb.RegionCode,
					"pool":      pool.Address,
					"serveMode": pool.ServeMode,
				}).Debug("Pool does not need update")
				continue
			}
			log.WithFields(log.Fields{
				"region":       gslb.RegionCode,
				"pool":         pool.Address,
				"oldServeMode": pool.ServeMode,
				"newServeMode": config.ServeMode,
			}).Info("Setting pool for update")

			pool.ServeMode = config.ServeMode
			gslb.Pool[index] = pool
			putRequests[gslb.RegionCode] = gslb
		}
	}

	// limit update concurrency to 1, as Dyn only allows one in progress job
	// see https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
	concurrencyLimit := 1
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	responsesCh := make(chan *httpResponse)
	for regionCode, payload := range putRequests {
		gslbUrl := fmt.Sprintf("GSLBRegion/%s/%s/%s/", url.PathEscape("ft.com"), url.PathEscape(config.Fqdn), url.PathEscape(regionCode))
		go func(payload gslbRegion) {
			semaphoreChan <- struct{}{}
			d.doGslbUpdate(gslbUrl, payload, responsesCh)
		}(payload)
	}

	defer func() {
		close(semaphoreChan)
		close(responsesCh)
	}()
	var responses []*httpResponse
readChannel:
	for {
		select {
		case response := <-responsesCh:
			<-semaphoreChan
			responses = append(responses, response)
			if len(responses) == len(putRequests) {
				hasErrors := false
				for _, response := range responses {
					if response.err != nil {
						log.WithFields(log.Fields{
							"event": "DYNECT_GSLB_PUT_FAILURE",
							"err":   response.err,
							"url":   response.url,
						}).Error("GSLB update failed")
						hasErrors = true
					} else {
						log.WithFields(log.Fields{
							"event": "DYNECT_GSLB_PUT_SUCCESS",
							"url":   response.url,
						}).Info("GSLB update succeeded")
					}
				}
				if hasErrors {
					return errors.New("One or more of the GSLB updates failed")
				}
				break readChannel
			}
		case <-time.After(20 * time.Second):
			return errors.New("Update requests did not return after 20 seconds")
		}
	}

	return nil
}

func (d *dynectService) GetGslbTTL(config GslbConfig) (time.Duration, error) {
	gslbServiceUrl := fmt.Sprintf("GSLB/%s/%s/", url.PathEscape("ft.com"), url.PathEscape(config.Fqdn))
	logger := log.WithFields(log.Fields{
		"url": gslbServiceUrl,
	})

	getResponse := struct {
		Status string      `json:"status"`
		Data   gslbService `json:"data"`
	}{}
	err := d.client.Do("GET", gslbServiceUrl, nil, &getResponse)
	if err == nil && getResponse.Status != "success" {
		err = errors.New("Dynect API call status was no success")
	}
	if err != nil {
		logger.WithFields(log.Fields{
			"event":    "DYNECT_API_GSLB_SERVICE_GET_FAILURE",
			"err":      err,
			"response": getResponse,
		}).Error("Failed to get gslb service")
		return 0, err
	}
	data := getResponse.Data
	logger.WithFields(log.Fields{
		"response": data,
	}).Debug("GSLB service GET response")

	return time.Duration(data.TTL) * time.Second, nil
}
