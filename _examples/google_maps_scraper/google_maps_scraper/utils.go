package google_maps_scraper

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/builtin/gosm"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tidwall/gjson"
)

func genGeocodingQuery(search string) (string, error) {
	query := url.Values{
		"tbm":      {"map"},
		"suggest":  {"p"},
		"gs_ri":    {"maps"},
		"hl":       {"en"},
		"authuser": {"0"},
		"q":        {search},
		"pb":       {PB_GEOCODING},
	}
	return query.Encode(), nil
}

func genSearchQuery(job Job) (string, error) {
	var pb string

	if job.cursor != 0 {
		pb = PB_SEARCH_PAGINATION_PREFIX + strconv.FormatFloat(job.loc.lng, 'f', 6, 32) + PB_SEARCH_PAGINATION_MID + strconv.FormatFloat(job.loc.lat, 'f', 6, 32) + PB_SEARCH_PAGINATION_SUFFIX + strconv.Itoa(job.cursor) + PB_SEARCH_PAGINATION_FINAL
	} else {
		pb = PB_SEARCH_BASE_PREFIX + strconv.FormatFloat(job.loc.lng, 'f', 6, 32) + PB_SEARCH_BASE_MID + strconv.FormatFloat(job.loc.lat, 'f', 6, 32) + PB_SEARCH_BASE_SUFFIX
	}

	query := url.Values{
		"tbm":      {"map"},
		"authuser": {"0"},
		"hl":       {"en"},
		"pb":       {pb},
		"q":        {job.query},
	}

	return query.Encode(), nil
}

func unicodeDecode(unicodeStr string) string {
	decodedStr, err := url.QueryUnescape(unicodeStr)

	if err != nil {
		return ""
	}
	return decodedStr
}

func ExtractMapResults(data []byte) []Record {
	body := bytes.TrimPrefix(data, []byte(")]}'"))

	// try both paths for records
	records := gjson.GetBytes(body, "64.#.1")
	if !records.Exists() || len(records.Array()) == 0 {
		records = gjson.GetBytes(body, "0.1.#.14")
	}

	query := gjson.GetBytes(body, "0.0")

	if !records.Exists() {
		return nil
	}

	_records := records.Array()
	finalRecords := make([]Record, 0, len(_records))

	var (
		wg sync.WaitGroup
		ch = make(chan *Record, len(_records))
	)

	wg.Add(len(_records))
	for _, _record := range _records {
		go func(res gjson.Result) {
			defer wg.Done()
			ch <- parseRecord(res, query.Str)
		}(_record)
	}

	wg.Wait()
	close(ch)

	for record := range ch {
		if record != nil {
			finalRecords = append(finalRecords, *record)
		}
	}

	return finalRecords
}

func parseRecord(res gjson.Result, query string) *Record {
	record := &Record{Query: query}

	// use gosm to map
	_ = gosm.Map(res, record)


	// post-process: unicode decoding
	record.Title = unicodeDecode(record.Title)
	record.Description = unicodeDecode(record.Description)
	record.ShortDescription = unicodeDecode(record.ShortDescription)
	record.Website = unicodeDecode(record.Website)
	record.ReviewsUrl = unicodeDecode(record.ReviewsUrl)
	record.WebResultsUrl = unicodeDecode(record.WebResultsUrl)
	record.GoogleReserveUrl = unicodeDecode(record.GoogleReserveUrl)

	// decode slices
	for i, v := range record.Categories {
		record.Categories[i] = unicodeDecode(v)
	}
	for i, v := range record.ReservationUrls {
		record.ReservationUrls[i] = unicodeDecode(v)
	}

	// decode nested fields
	for i := range record.Orders {
		record.Orders[i].Url = unicodeDecode(record.Orders[i].Url)
		record.Orders[i].Platform = unicodeDecode(record.Orders[i].Platform)
		record.OrderUrls = append(record.OrderUrls, record.Orders[i].Url)
		record.OrderPlatforms = append(record.OrderPlatforms, record.Orders[i].Platform)
	}

	for i := range record.Details {
		for j := range record.Details[i].Items {
			record.Details[i].Items[j].Value = unicodeDecode(record.Details[i].Items[j].Value)
		}
	}

	// formatting
	var ohLines []string
	for _, oh := range record.OpenHours {
		ohLines = append(ohLines, fmt.Sprintf("%s:%s", oh.Day, oh.Timing))
	}
	record.OpenHoursStr = strings.Join(ohLines, " | ")

	record.ReviewDistribution = parseReviewDistribution(res.Get("52.3"))

	// special split
	if parts := strings.Split(res.Get("2.1").Str, " "); len(parts) > 0 {
		record.StateCode = parts[0]
	}

	return record
}

func parseReviewDistribution(raw gjson.Result) map[string]uint64 {
	dist := make(map[string]uint64, 5)
	for i, rating := range raw.Array() {
		dist[fmt.Sprintf("%d", i+1)] = rating.Uint()
	}
	return dist
}



// below are utilities
func generateSearchUrl(baseUrl string, job Job) string {
	query, _ := genSearchQuery(job)
	return baseUrl + "/search?" + query
}

func generateGeocodingUrl(baseUrl string, job Job) string {
	query, _ := genGeocodingQuery(job.query)
	return baseUrl + "/s?" + query
}

func extractGeocoding(data []byte) (lat, lng float64, query string, ok bool) {
	body := bytes.TrimPrefix(data, []byte(")]}'"))

	type geoData struct {
		Lat float64 `gos:"2"`
		Lng float64 `gos:"3"`
	}

	type geoResponse struct {
		Query string   `gos:"0.0"`
		Geo0  *geoData `gos:"0.1.0.22.11"`
		Geo1  *geoData `gos:"0.1.1.22.11"`
	}

	var res geoResponse
	_ = gosm.Map(body, &res)

	if res.Geo0 != nil && (res.Geo0.Lat != 0 || res.Geo0.Lng != 0) {
		return res.Geo0.Lat, res.Geo0.Lng, res.Query, true
	}

	if res.Geo1 != nil && (res.Geo1.Lat != 0 || res.Geo1.Lng != 0) {
		return res.Geo1.Lat, res.Geo1.Lng, res.Query, true
	}

	return 0, 0, "", false
}

func prepareRequest(req *core.Request, url string, job Job) *core.Request {
	return req.Url(url).
		AddHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36").
		Meta("JOB", job)
}

func getJob(resp core.IResponseReader) (Job, bool) {
	data, ok := resp.Meta("JOB")
	if !ok {
		return Job{}, false
	}

	job, ok := data.(Job)
	if !ok {
		return Job{}, false
	}

	return job, true
}

