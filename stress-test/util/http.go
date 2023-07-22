package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"supernova/scheduler/model"
)

func RegisterTrigger(schedulerAddress string, trigger *model.Trigger) error {
	return SendRequest(schedulerAddress+"/triggers", "POST", trigger)
}

func RegisterJob(schedulerAddress string, job *model.Job) error {
	return SendRequest(schedulerAddress+"/jobs", "POST", job)
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

	log.Printf("Request to %s completed with status code: %d\n", url, resp.StatusCode)
	return nil
}
