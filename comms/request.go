package comms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// For converting zip code to city/state:
// https://github.com/USPS/api-examples?tab=readme-ov-file#city-and-state
// curl	-X 'GET' 'https://api.usps.com/addresses/v1/city-state?ZIPCode=30022' \
// 	--header 'accept: application/json' \
// 	--header 'X-User-Id: XXXXXXXXXXX' \
// 	--header 'Authorization: Bearer $TOKEN' \

// or just use this api
// https://openweathermap.org/api/geocoding-api

const geoCodingHost = "https://api.openweathermap.org/geo/1.0/direct"

type geoCodeAPIResponseItem struct {
	// TODO can you unmarshal without struct tags as long as name matches json name? try it
	name    string `json:"name"`
	lat     string `json:"lat"`
	lon     string `json:"long"`
	country string `json:"country"`
	state   string `json:"state"`
}

type geoCodeAPIResponse []geoCodeAPIResponseItem

type Location struct {
	Lat  string `json:"lat"`
	Lon  string `json:"long"`
	Name string `json:"name"`
}

var client = &http.Client{
	Timeout: time.Second * 5,
}

func GetLocation(query string, apiKey string) (Location, error) {
	// handle spaces
	spaceToComma := strings.Join(strings.Split(query, " "), ",")
	// TODO possibly convert two-letter state code to three-letter, because API only handles the latter
	fmt.Println("spaceToComma", spaceToComma)
	var l Location
	req, err := http.NewRequest(http.MethodGet, geoCodingHost, nil)
	if err != nil {
		fmt.Println(err)
		return l, err
	}
	q := url.Values{}
	q.Add("appid", apiKey)
	q.Add("q", spaceToComma)
	req.URL.RawQuery = q.Encode()
	fmt.Println(req.URL)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return l, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return l, err
	}
	var apiResponse geoCodeAPIResponse
	fmt.Println(resp.Status)
	fmt.Println("body", string(body))
	if err = json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Println(err)
		return l, err
	}
	fmt.Printf("%v", apiResponse)
	return l, nil
}
