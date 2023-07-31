package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"supernova/scheduler/model"
	"sync"

	"github.com/cloudwego/kitex/pkg/klog"
)

func RegisterTrigger(schedulerAddress string, trigger *model.Trigger) error {
	return SendRequest(schedulerAddress+"/trigger", "POST", trigger)
}

func RegisterTriggers(schedulerAddress string, triggers []*model.Trigger) {
	var wg sync.WaitGroup
	//保证数据库不会被击穿
	concurrencyLimit := make(chan struct{}, 500)

	groupSize := 500
	groupCount := (len(triggers) + groupSize - 1) / groupSize

	for i := 0; i < groupCount; i++ {
		wg.Add(1)
		to := (i + 1) * groupSize
		if to > len(triggers) {
			to = len(triggers)
		}
		go func(start, end int) {
			defer wg.Done()

			for _, trigger := range triggers[start:end] {
				concurrencyLimit <- struct{}{}
				err := RegisterTrigger(schedulerAddress, trigger)
				if err != nil {
					panic(err)
				}
				<-concurrencyLimit
			}

		}(i*groupSize, to)
	}

	wg.Wait()
	close(concurrencyLimit)
	klog.Infof("triggers inserted success, count:%v", len(triggers))
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
	klog.Tracef("Request to %s completed with status code: %d\n", url, resp.StatusCode)
	return nil
}
