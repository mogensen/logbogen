package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Address contains address fields specific to OpenStreetMap
// much more fields could be here? https://nominatim.org/release-docs/latest/api/Output/
type Address struct {
	HouseNumber   string `json:"house_number"`
	Road          string `json:"road"`
	Pedestrian    string `json:"pedestrian"`
	Footway       string `json:"footway"`
	Cycleway      string `json:"cycleway"`
	Highway       string `json:"highway"`
	Path          string `json:"path"`
	Suburb        string `json:"suburb"`
	City          string `json:"city"`
	Town          string `json:"town"`
	Village       string `json:"village"`
	Hamlet        string `json:"hamlet"`
	Neighbourhood string `json:"neighbourhood"`
	Municipality  string `json:"municipality"`
	County        string `json:"county"`
	Country       string `json:"country"`
	CountryCode   string `json:"country_code"`
	State         string `json:"state"`
	StateDistrict string `json:"state_district"`
	Postcode      string `json:"postcode"`
}

// SimpleDisplayName compiles a small local name for the given osm response
func (g GeocodeResponse) SimpleDisplayName() string {
	locality := []string{}

	// First find if it has a type specific name
	if g.Name != "" {
		locality = append(locality, g.Name)
	}

	// Add the geographic name
	if g.Address.City != "" && !strings.Contains(strings.Join(locality, ""), g.Address.City) {
		locality = append(locality, g.Address.City)
	} else if g.Address.Town != "" && !strings.Contains(strings.Join(locality, ""), g.Address.Town) {
		locality = append(locality, g.Address.Town)
	} else if g.Address.Village != "" && !strings.Contains(strings.Join(locality, ""), g.Address.Village) {
		locality = append(locality, g.Address.Village)
	} else if g.Address.Hamlet != "" && !strings.Contains(strings.Join(locality, ""), g.Address.Hamlet) {
		locality = append(locality, g.Address.Hamlet)
	} else if g.Address.Neighbourhood != "" && !strings.Contains(strings.Join(locality, ""), g.Address.Neighbourhood) {
		locality = append(locality, g.Address.Neighbourhood)
	} else if g.Address.Municipality != "" && !strings.Contains(strings.Join(locality, ""), g.Address.Municipality) {
		locality = append(locality, g.Address.Municipality)
	}

	// Then the country
	if g.Address.Country != "" {
		locality = append(locality, g.Address.Country)
	}

	return strings.Join(locality, ", ")
}

type GeocodeResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Error       string
	Address     Address `json:"address"`
}

func ReverseGeocode(lat, lng float64) (*GeocodeResponse, error) {

	fmt.Printf("1. Performing Http Get... : %s\n", reverseGeocodeURL(lat, lng))
	resp, err := http.Get(reverseGeocodeURL(lat, lng))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	// Convert response body to Todo struct
	geocodeResponse := &GeocodeResponse{}
	json.Unmarshal(bodyBytes, geocodeResponse)
	fmt.Printf("API Response as struct %+v\n", geocodeResponse)
	return geocodeResponse, nil
}

func geocodeURL(address string) string {
	return "https://nominatim.openstreetmap.org/search?format=jsonv2&zoom=16&limit=1&q=" + address
}

func reverseGeocodeURL(lat, lng float64) string {
	return "https://nominatim.openstreetmap.org/reverse?" + fmt.Sprintf("format=jsonv2&zoom=16&lat=%f&lon=%f", lat, lng)
}
