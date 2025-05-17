package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type EnrichResult struct {
	Age         *int
	Gender      *string
	Nationality *string
}

func Enrich(name string) (*EnrichResult, error) {
	var (
		wg  sync.WaitGroup
		res EnrichResult
		mu  sync.Mutex
		err error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		var a struct {
			Age *int `json:"age"`
		}
		if e := callAPI(fmt.Sprintf("https://api.agify.io/?name=%s", name), &a); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			return
		}
		mu.Lock()
		res.Age = a.Age
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		var g struct {
			Gender *string `json:"gender"`
		}
		if e := callAPI(fmt.Sprintf("https://api.genderize.io/?name=%s", name), &g); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			return
		}
		mu.Lock()
		res.Gender = g.Gender
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		var n struct {
			Country []struct {
				CountryID   string  `json:"country_id"`
				Probability float64 `json:"probability"`
			} `json:"country"`
		}
		if e := callAPI(fmt.Sprintf("https://api.nationalize.io/?name=%s", name), &n); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			return
		}
		if len(n.Country) > 0 {
			mu.Lock()
			res.Nationality = &n.Country[0].CountryID
			mu.Unlock()
		}
	}()

	wg.Wait()
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func callAPI(url string, out interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
