package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"supernova/scheduler/model"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func RegisterTrigger(schedulerAddress string, trigger *model.Trigger) error {
	return SendRequest(schedulerAddress+"/trigger", "POST", trigger)
}

func doRegisterTriggers(schedulerAddress string, triggers []*model.Trigger) error {
	return SendRequest(schedulerAddress+"/trigger/batch", "POST", triggers)
}

func RegisterTriggers(schedulerAddress string, triggers []*model.Trigger) {
	start := time.Now()
	//var wg sync.WaitGroup
	//保证数据库不会被击穿
	//concurrencyLimit := make(chan struct{}, 30)

	groupSize := 4000
	groupCount := (len(triggers) + groupSize - 1) / groupSize

	for i := 0; i < groupCount; i++ {
		//wg.Add(1)
		to := (i + 1) * groupSize
		if to > len(triggers) {
			to = len(triggers)
		}
		//go func(start, end int) {
		//defer wg.Done()

		//concurrencyLimit <- struct{}{}
		err := doRegisterTriggers(schedulerAddress, triggers[i*groupSize:to])
		if err != nil {
			panic(err)
		}
		//time.Sleep(1 * time.Millisecond)
		//<-concurrencyLimit
		//}(i*groupSize, to)
	}

	//wg.Wait()
	//close(concurrencyLimit)
	costMs := time.Since(start).Milliseconds()
	if costMs == 0 {
		costMs = 1
	}
	log.Printf("triggers inserted success, count: %v, cost time:%vms, speed: %v entry/ms",
		len(triggers), costMs, float64(len(triggers))/float64(costMs))
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
