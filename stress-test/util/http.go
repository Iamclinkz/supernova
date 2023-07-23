package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"supernova/scheduler/model"
	"sync"
)

func RegisterTrigger(schedulerAddress string, trigger *model.Trigger) error {
	return SendRequest(schedulerAddress+"/trigger", "POST", trigger)
}

func RegisterTriggers(schedulerAddress string, triggers []*model.Trigger) {
	var wg sync.WaitGroup
	concurrencyLimit := make(chan struct{}, 500)

	for _, trigger := range triggers {
		concurrencyLimit <- struct{}{}

		wg.Add(1)
		go func(t *model.Trigger) {
			defer wg.Done()

			err := RegisterTrigger(schedulerAddress, t)
			if err != nil {
				panic(err)
			}

			// 将令牌放回channel
			<-concurrencyLimit
		}(trigger)
	}

	wg.Wait()
	close(concurrencyLimit)
}

func RegisterJob(schedulerAddress string, job *model.Job) error {
	return SendRequest(schedulerAddress+"/job", "POST", job)
}

func SendRequest(url, method string, data interface{}) error {
	client := &http.Client{}

	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(resp.Body)
	}
	log.Printf("Request to %s completed with status code: %d\n", url, resp.StatusCode)
	return nil
}
