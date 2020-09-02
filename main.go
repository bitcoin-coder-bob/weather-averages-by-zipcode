package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func main() {
	// https://www.timeanddate.com/weather/@z-us-08008/climate
	url := "https://www.timeanddate.com/weather/@z-us-"
	client := &http.Client{}
	counter := 499
	max := 100000

	for ; counter <= max; counter++ {
		time.Sleep(200 * time.Millisecond)
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

		req, err := http.NewRequest("GET", url+zipCode+"/climate", nil)
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
			fmt.Printf("zip: %v   failed stauts: %v\n", zipCode, response.StatusCode)
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
		climateData := re.FindString(strContent)

		urlZipCode := "https://api.promaptools.com/service/us/zip-lat-lng/get/?zip="
		key := "&key=17o8dysaCDrgv1c"

		req, err = http.NewRequest("GET", urlZipCode+zipCode+key, nil)
		if err != nil {
			fmt.Printf("error requesting: %v   %v\n", zipCode, err.Error())
			continue
		}
		response, err = client.Do(req)
		if err != nil {
			fmt.Printf("error getting response from client when streaming from location:\n%v\n%s", response, err.Error())
			fmt.Println(zipCode)
			continue
		}
		if response.StatusCode != 200 {
			fmt.Printf("zip: %v   status: %v\n", zipCode, response.StatusCode)
			continue
		}

		content, err = ioutil.ReadAll(response.Body)
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
		if len(string(climateData)) > 0 {
			fmt.Printf("zip:%v\n", zipCode)
			value, err := sjson.Set(string(climateData)[9:], "lat", lat)
			if err != nil {
				fmt.Printf("error on zip:%v\n", zipCode)
			}
			value, _ = sjson.Set(value, "long", long)
			err = ioutil.WriteFile("./data/"+zipCode+".txt", []byte(value), 0644)
			if err != nil {
				fmt.Printf("error writing %v\n", zipCode)
				continue
			}
		}
	}
}
