package main

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/tidwall/gjson"
)

func main() {
	counter := 499
	max := 100000
	allValidCoords := map[string][]string{}
	validZip := true
	for ; counter <= max; counter++ {
		validZip = true
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
		data, err := ioutil.ReadFile("./data/" + zipCode + ".txt")
		if err != nil {
			fmt.Printf("Error reading file %v.txt: %v\n", zipCode, err.Error())
			continue
		}
		stringData := string(data)
		months := gjson.Get(stringData, "months").Array()
		for _, month := range months {
			min, err := strconv.Atoi(month.Get("min").String())
			max, err := strconv.Atoi(month.Get("max").String())
			if err != nil {
				fmt.Printf("Error converting int to string %v.txt: %v\n", zipCode, err.Error())
			}
			if month.Get("min").Type.String() == "Null" || month.Get("max").Type.String() == "Null" {
				fmt.Println("nulls\n")
				validZip = false
				break
			}
			//  CONFIUGRE THESE MIN AND MAX AVERAGE TEMPERATURE VALUES
			if min < 50 || max > 90 {
				validZip = false
				break
			}
		}
		if validZip == true {
			fmt.Printf("zip is valid: %v\n", zipCode)
			latLong := []string{gjson.Get(stringData, "lat").String(), gjson.Get(stringData, "long").String()}
			allValidCoords[zipCode] = latLong
		}
	}
	startString := `eqfeed_callback({"zips":[`
	endString := `]})`
	totalCoords := len(allValidCoords)
	fmt.Printf("total: %v", totalCoords)
	coordCounter := 0
	last := false
	for _, coords := range allValidCoords {
		if coordCounter+1 == totalCoords {
			last = true
		}
		startString += `{"lat":"` + coords[0] + `","long":"` + coords[1] + `"}`
		if !last {
			startString += ","
		}
		coordCounter++
	}
	startString += endString
	err := ioutil.WriteFile("./sample.js", []byte(startString), 0644)
	if err != nil {
		fmt.Printf("error writing sammple.js :  %v\n", err.Error())
		return
	}
}
