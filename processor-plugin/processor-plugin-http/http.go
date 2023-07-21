package processor_plugin_http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"supernova/pkg/api"
	"time"

	"github.com/armon/circbuf"
)

//从dkron抄过来的。。改成适配supernova/executor/processor的版本了

const (
	// timeout seconds
	timeout = 30
	// maxBufSize limits how much data we collect from a handler.
	// This is to prevent Serf's memory from growing to an enormous
	// amount due to a faulty handler.
	maxBufSize = 256000
)

// HTTP process http request
type HTTP struct {
}

// Process Process method of the plugin
// "executor": "http",
//
//	"executor_config": {
//	    "method": "GET",             // Request method in uppercase
//	    "url": "http://example.com", // Request url
//	    "headers": "[]"              // Json string, such as "[\"Content-Type: application/json\"]"
//	    "body": "",                  // POST body
//	    "timeout": "30",             // Request timeout, unit seconds
//	    "expectCode": "200",         // Expect response code, such as 200,206
//	    "expectBody": "",            // Expect response body, support regexp, such as /success/
//	    "debug": "true"              // Debug option, will log everything when this option is not empty
//	}
func (s *HTTP) Process(job *api.Job) *api.JobResult {
	out, err := s.ExecuteImpl(job)
	result := &api.JobResult{
		Ok:     err == nil,
		Err:    err.Error(),
		Result: string(out),
	}

	return result
}

func (s *HTTP) GetGlueType() string {
	return "Http"
}

// ExecuteImpl do http request
func (s *HTTP) ExecuteImpl(job *api.Job) ([]byte, error) {
	output, _ := circbuf.NewBuffer(maxBufSize)
	var debug bool
	if job.Source["debug"] != "" {
		debug = true
		log.Printf("config  %#v\n\n", job.Source)
		output.Write([]byte(fmt.Sprintf("Config: %#v\n", job.Source)))
	}

	if job.Source["url"] == "" {
		return output.Bytes(), errors.New("url is empty")
	}

	if job.Source["method"] == "" {
		return output.Bytes(), errors.New("method is empty")
	}

	req, err := http.NewRequest(job.Source["method"], job.Source["url"], bytes.NewBuffer([]byte(job.Source["body"])))
	if err != nil {
		return output.Bytes(), err
	}
	req.Close = true

	var headers []string
	if job.Source["headers"] != "" {
		if err := json.Unmarshal([]byte(job.Source["headers"]), &headers); err != nil {
			output.Write([]byte("Error: parsing headers failed\n"))
		}
	}

	for _, h := range headers {
		if h != "" {
			kv := strings.Split(h, ":")
			req.Header.Set(kv[0], strings.TrimSpace(kv[1]))
		}
	}
	if debug {
		log.Printf("request  %#v\n\n", req)
	}

	client, warns := createClient(job.Source)
	for _, warn := range warns {
		output.Write([]byte(fmt.Sprintf("Warning: %s.\n", warn.Error())))
	}

	resp, err := client.Do(req)
	if err != nil {
		return output.Bytes(), err
	}

	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return output.Bytes(), err
	}

	if debug {
		log.Printf("response  %#v\n\n", resp)
		log.Printf("response body  %#v\n\n", string(out))
	}

	// write the response to output
	_, err = output.Write(out)
	if err != nil {
		return output.Bytes(), err
	}

	// match response code
	if job.Source["expectCode"] != "" && !strings.Contains(job.Source["expectCode"]+",", fmt.Sprintf("%d,", resp.StatusCode)) {
		return output.Bytes(), errors.New("received response code does not match the expected code")
	}

	// match response
	if job.Source["expectBody"] != "" {
		if m, _ := regexp.MatchString(job.Source["expectBody"], string(out)); !m {
			return output.Bytes(), errors.New("received response body did not match the expected body")
		}
	}

	// Warn if buffer is overritten
	if output.TotalWritten() > output.Size() {
		log.Printf("'%s %s': generated %d bytes of output, truncated to %d",
			job.Source["method"], job.Source["url"],
			output.TotalWritten(), output.Size())
	}

	return output.Bytes(), nil
}

// createClient always returns a new http client. Any errors returned are
// errors in the configuration.
func createClient(config map[string]string) (http.Client, []error) {
	var errs []error

	_timeout, err := atoiOrDefault(config["timeout"], timeout)
	if config["timeout"] != "" && err != nil {
		errs = append(errs, fmt.Errorf("invalid timeout value: %s", err.Error()))
	}

	tlsconf := &tls.Config{}
	tlsconf.InsecureSkipVerify, err = strconv.ParseBool(config["tlsNoVerifyPeer"])
	if config["tlsNoVerifyPeer"] != "" && err != nil {
		errs = append(errs, fmt.Errorf("not disabling certificate validation: %s", err.Error()))
	}

	if config["tlsCertificateFile"] != "" {
		cert, err := tls.LoadX509KeyPair(config["tlsCertificateFile"], config["tlsCertificateKeyFile"])
		if err == nil {
			tlsconf.Certificates = append(tlsconf.Certificates, cert)
		} else {
			errs = append(errs, fmt.Errorf("not using client certificate: %s", err.Error()))
		}
	}

	if config["tlsRootCAsFile"] != "" {
		tlsconf.RootCAs, err = loadCertPool(config["tlsRootCAsFile"])
		if err != nil {
			errs = append(errs, fmt.Errorf("using system root CAs instead of configured CAs: %s", err.Error()))
		}
	}

	return http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsconf},
		Timeout:   time.Duration(_timeout) * time.Second,
	}, errs
}

// loadCertPool creates a CertPool using the given file
func loadCertPool(filename string) (*x509.CertPool, error) {
	certsFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certsFile) {
		return nil, fmt.Errorf("no certificates in file")
	}

	return roots, nil
}

// atoiOrDefault returns the integer value of s, or a default value
// if s could not be converted, along with an error.
func atoiOrDefault(s string, _default int) (int, error) {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i, nil
	}
	return _default, fmt.Errorf("\"%s\" not understood (%s), using default value of %d", s, err, _default)
}
