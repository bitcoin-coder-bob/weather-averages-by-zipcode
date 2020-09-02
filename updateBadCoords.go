// This file need not be used. Files where lat or long is 0.000000 means that was just coord returned by the endpoint, nothing
// I can do about that. Files without lat or long means that the endpoint didnt recognize tthe zipcode. Could be a military base.
// I found 718 zips that met this criteria which seems like a high amount.

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func main() {
	counter := 499
	max := 100000
	totalUpdated := 0
	client := &http.Client{}
	for ; counter <= max; counter++ {
		zipCode := ""
		if counter < 10 {
			zipCode = "0000" + strconv.Itoa(counter)
		} else if counter < 100 {
			zipCode = "000" + strconv.Itoa(counter)
		} else if counter < 1000 {
			zipCode = "00" + strconv.Itoa(counter)
		} else if counter < 10000 {
			zipCode = "0" + strconv.Itoa(counter)
		} else if counter < 100000 {
			zipCode = strconv.Itoa(counter)
		}
		// fmt.Printf("zipCode: %v\n", zipCode)
		data, err := ioutil.ReadFile("./data/" + zipCode + ".txt")
		if err != nil {
			// fmt.Printf("Error reading file %v.txt: %v\n", zipCode, err.Error())
			continue
		}
		stringData := string(data)

		lat := gjson.Get(stringData, "lat").String()
		long := gjson.Get(stringData, "long").String()
		// if lat == "" || long == "" {

		if lat == "" || long == "" || lat == "0.000000" || long == "0.000000" {
			fmt.Printf("here zipcode: %v\n", zipCode)
			reGetWeatherData := false
			if len(stringData) < 1 {
				reGetWeatherData = true
			}
			climateData := ""
			if reGetWeatherData {
				urlWeather := "https://www.timeanddate.com/weather/@z-us-"
				req, err := http.NewRequest("GET", urlWeather+zipCode+"/climate", nil)
				if err != nil {
					fmt.Println(zipCode)
					continue
				}
				response, err := client.Do(req)
				if err != nil {
					fmt.Printf("error getting response from client when streaming from location:\n%v\n%s", response, err.Error())
					fmt.Println(zipCode)
					continue
				}
				if response.StatusCode != 200 {
					// fmt.Println("failed response (stream from external source): status code %d", response.StatusCode)
					fmt.Printf("zip: %v   stauts: %v\n", zipCode, response.StatusCode)
					continue
				}

				content, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("error reading data from response body:\n%s", err.Error())
					fmt.Println(zipCode)
					continue
				}
				strContent := string(content)
				re := regexp.MustCompile(`var data.*`)
				climateData = re.FindString(strContent)
				fmt.Println("clmate data: %v\n", climateData)
			}

			urlZipCode := "https://api.promaptools.com/service/us/zip-lat-lng/get/?zip="
			key := "&key=17o8dysaCDrgv1c"
			fmt.Println(urlZipCode + zipCode + key)
			req, err := http.NewRequest("GET", urlZipCode+zipCode+key, nil)
			if err != nil {
				fmt.Printf("error requesting: %v   %v\n", zipCode, err.Error())
				continue
			}
			response, err := client.Do(req)
			if err != nil {
				fmt.Printf("error getting response from client when streaming from location:\n%v\n%s", response, err.Error())
				fmt.Println(zipCode)
				continue
			}
			if response.StatusCode != 200 {
				fmt.Printf("zip: %v   status: %v\n", zipCode, response.StatusCode)
				continue
			}

			content, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Printf("error reading data from response body:\n%s", err.Error())
				fmt.Println(zipCode)
				continue
			}
			status := gjson.Get(string(content), "status").String()
			if status != "1" {
				continue
			}
			coords := gjson.Get(string(content), "output").Array()[0]
			lat := coords.Get("latitude").String()
			long := coords.Get("longitude").String()
			fmt.Printf("coord data:%v %v\n", lat, long)
			jsonToEdit := climateData
			if reGetWeatherData {
				jsonToEdit = climateData
			} else {
				jsonToEdit = stringData
			}
			value, err := sjson.Set(string(jsonToEdit)[9:], "lat", lat)
			if err != nil {
				fmt.Printf("error on zip:%v\n", zipCode)
				continue
			}
			value, _ = sjson.Set(value, "long", long)
			// fmt.Printf("new data: %v\n", value)
			// err = ioutil.WriteFile("./data/"+zipCode+".txt", []byte(value), 0644)
			// if err != nil {
			// 	fmt.Printf("error writing %v\n", zipCode)
			// 	continue
			// }
			totalUpdated++
			// fmt.Printf("New content %v %v:  %v\n", zipCode, totalUpdated, value)
		}
	}
	fmt.Printf("total updpated: %v\n", totalUpdated)
}
