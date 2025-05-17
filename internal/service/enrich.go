package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

type EnrichResult struct {
	Age         *int
	Gender      *string
	Nationality *string
}

func Enrich(name string) (*EnrichResult, error) {
	log.Debugf("service.Enrich: starting enrichment for name=%s", name)
	var (
		wg  sync.WaitGroup
		res EnrichResult
		mu  sync.Mutex
		err error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.agify.io/?name=%s", name)
		log.Debugf("service.Enrich: calling Agify API: %s", url)
		var a struct {
			Age *int `json:"age"`
		}
		if e := callAPI(url, &a); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			log.WithError(e).Error("service.Enrich: Agify call failed")
			return
		}
		mu.Lock()
		res.Age = a.Age
		mu.Unlock()
		log.Debugf("service.Enrich: Agify result: %v", a.Age)
	}()

	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
		log.Debugf("service.Enrich: calling Genderize API: %s", url)
		var g struct {
			Gender *string `json:"gender"`
		}
		if e := callAPI(url, &g); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			log.WithError(e).Error("service.Enrich: Genderize call failed")
			return
		}
		mu.Lock()
		res.Gender = g.Gender
		mu.Unlock()
		log.Debugf("service.Enrich: Genderize result: %v", g.Gender)
	}()

	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)
		log.Debugf("service.Enrich: calling Nationalize API: %s", url)
		var n struct {
			Country []struct {
				CountryID   string  `json:"country_id"`
				Probability float64 `json:"probability"`
			} `json:"country"`
		}
		if e := callAPI(url, &n); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			log.WithError(e).Error("service.Enrich: Nationalize call failed")
			return
		}
		if len(n.Country) > 0 {
			mu.Lock()
			res.Nationality = &n.Country[0].CountryID
			mu.Unlock()
			log.Debugf("service.Enrich: Nationalize result: %v", n.Country[0].CountryID)
		}
	}()

	wg.Wait()
	if err != nil {
		log.WithError(err).Error("service.Enrich: enrichment failed")
		return nil, err
	}
	log.Debugf("service.Enrich: completed enrichment for name=%s: %+v", name, res)
	return &res, nil
}

func callAPI(url string, out interface{}) error {
	log.Debugf("service.callAPI: GET %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
