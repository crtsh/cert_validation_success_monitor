package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type VersionDate struct {
	Version		string
	Date		string
}

type CVSBCDatum struct {
	Date		string
	Label		string
	Histogram	[]int64
	Count		int64
	Sum			int64
}

type CVSBCData struct {
	Data		[]CVSBCDatum
	Buckets		[]int
}

func main() {
	const CHANNEL = "beta"
	resp, err := http.Get(fmt.Sprintf("https://aggregates.telemetry.mozilla.org/aggregates_by/submission_date/channels/%s/dates/", CHANNEL))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var versions_and_dates []VersionDate
	err = json.Unmarshal(body, &versions_and_dates)
	if err != nil {
		log.Fatalln(err)
	}

	min_version := math.MaxInt32
	max_version := 0
	for _, vd := range versions_and_dates {
		version, err := strconv.Atoi(vd.Version)
		if err != nil {
			log.Fatalln(err)
		}
		if version < min_version {
			min_version = version
		}
		if version > max_version {
			max_version = version
		}
	}

	counter := 0
	for v := min_version; v <= max_version; v++ {
		var dates []string
		for _, vd := range versions_and_dates {
			version, err := strconv.Atoi(vd.Version)
			if err != nil {
				log.Fatalln(err)
			}

			if v == version {
				dates = append(dates, vd.Date)
			}
		}

		const MAX_DATES_PER_CALL = 256
		for i := 0; i < len(dates); i += MAX_DATES_PER_CALL {
			j := i + MAX_DATES_PER_CALL;
			if j > len(dates) {
				j = len(dates)
			}

			counter += (j - i);
			fmt.Fprintf(os.Stderr, "Processing %d of %d\n", counter, len(versions_and_dates))

			resp, err = http.Get(fmt.Sprintf("https://aggregates.telemetry.mozilla.org/aggregates_by/submission_date/channels/beta/?version=%d&dates=%s&metric=CERT_VALIDATION_SUCCESS_BY_CA", v, strings.Join(dates[i:j], ",")))
			if err != nil {
				log.Fatalln(err)
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}
			defer resp.Body.Close()

			var cvsbc_data CVSBCData
			err = json.Unmarshal(body, &cvsbc_data)
			if err != nil {
				log.Fatalln(err)
			}

			for _, d := range cvsbc_data.Data {
				for n, h := range d.Histogram {
					fmt.Printf("%s\t%s\t%d\t%d\t%d\n", d.Date, CHANNEL, v, cvsbc_data.Buckets[n], h)
				}
			}
		}
	}
}
