package service

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func fetch[T any](url string) (*T, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

type AgifyResp struct {
	Age *int `json:"age"`
}
type GenderizeResp struct {
	Gender *string `json:"gender"`
}
type NationalizeResp struct {
	Country []struct {
		CountryID   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}

func Enrich(name string) (age *int, gender *string, nationality *string, err error) {
	if a, e := fetch[AgifyResp](fmt.Sprintf("https://api.agify.io/?name=%s", name)); e == nil {
		age = a.Age
	}
	if g, e := fetch[GenderizeResp](fmt.Sprintf("https://api.genderize.io/?name=%s", name)); e == nil {
		gender = g.Gender
	}
	if n, e := fetch[NationalizeResp](fmt.Sprintf("https://api.nationalize.io/?name=%s", name)); e == nil && len(n.Country) > 0 {
		id := n.Country[0].CountryID
		nationality = &id
	}
	return
}
