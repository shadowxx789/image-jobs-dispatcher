package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Repeater struct {
	ClientTimeout  time.Duration
	Attempts	   time.Duration
	URI string
	Headers		   http.Header
	Body		   string
	Count          int
}

type RepeaterInterface interface {
	MakeRequest(httpMethod Method, data io.Reader) ([]byte, error)
}

type Method int

const (
	GET = iota
	POST
)
func (m Method) ToString() string {
	switch m {
	case GET:
		return "GET"
	case POST:
		return "POST"
	default:
		return fmt.Sprintf("%d", int(m))
	}
}

//MakeRequest make request with supporting retry pattern
func (r *Repeater) MakeRequest(httpMethod Method, data io.Reader) ([]byte, error) {
	var res []byte
	client := http.Client{
		Timeout: r.ClientTimeout * time.Second,
	}
	request, err := http.NewRequest(httpMethod.ToString(), r.URI, data)
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("[ERROR] cannot create %s request: %#v; URL: %s", httpMethod.ToString(), err, r.URI)
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		log.Printf("[ERROR] can not make %s request: %#v", httpMethod.ToString(), err)
		if errClose := response.Body.Close(); errClose != nil {
			log.Printf("[ERROR] can not close response body %#v", errClose)
			return nil, err
		}
		sumTimeout := r.Attempts * time.Second
		ticker := time.NewTicker(sumTimeout)
		cancel := make(chan struct{})
		go func() {
			defer func() {
				cancel <- struct{}{}
			}()
			time.Sleep(10 * sumTimeout)
		}()
		for {
			select {
			case <-ticker.C:
				switch httpMethod {
				case GET:
					response, err = client.Get(r.URI)
				case POST:
					response, err = client.Post(r.URI, "application/json", data)
				default:
					err = errors.New("can not detect http method")
				}
				if err != nil {
					log.Printf("[ERROR] can not make %s request: %#v ", httpMethod.ToString(), err)
					if errClose := response.Body.Close(); errClose != nil {
						log.Printf("[ERROR] can not close response body %#v", errClose)
					}
					continue
				}
				break
			case <-cancel:
				log.Printf("[WARN] completed repeater call. API is not reachible")
				break
			}
		}
		return nil, err
	}
	if response == nil && err != nil {
		return nil, err
	}

	res, err = ioutil.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusOK {
		log.Printf("[ERROR] can not read response body %#v", err)
		if errClose := response.Body.Close(); errClose != nil {
			log.Printf("[ERROR] can not close response body %#v", errClose)
		}
		return nil, err
	}

	if errClose := response.Body.Close(); errClose != nil {
		log.Printf("[ERROR] can not close response body %#v", errClose)
	}
	return res, nil
}
