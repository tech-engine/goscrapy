package google_maps_scraper

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

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

func extractMapResults(data []byte) []Record {
	body := bytes.TrimPrefix(data, []byte(")]}'"))

	records := gjson.GetBytes(body, "0.1.#.14")
	query := gjson.GetBytes(body, "0.0")

	if !(records.Exists() && query.Exists()) {
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
	stateCode := strings.Split(res.Get("2.1").Str, " ")
	openingHours := parseOpeningHours(res.Get("203.0"))

	return &Record{
		Query:              query,
		Street:             res.Get("2.0").Str,
		City:               res.Get("183.1.3").Str,
		ZipCode:            res.Get("183.1.4").Str,
		State:              res.Get("183.1.5").Str,
		StateCode:          stateCode[0],
		Country:            res.Get("2.2").Str,
		CountryCode:        res.Get("183.1.6").Str,
		Review:             res.Get("4.7").Float(),
		ReviewsUrl:         unicodeDecode(res.Get("4.3.0").Str),
		ReviewsCount:       res.Get("4.8").Uint(),
		Website:            unicodeDecode(res.Get("7.1").Str),
		Title:              unicodeDecode(res.Get("11").Str),
		Categories:         parseStringList(res.Get("13")),
		TimeZone:           res.Get("30").Str,
		Description:        unicodeDecode(res.Get("32.1.1").Str),
		ShortDescription:   unicodeDecode(res.Get("32.0.1").Str),
		ReviewDistribution: parseReviewDistribution(res.Get("52.3")),
		WebResultsUrl:      unicodeDecode(res.Get("174.0").Str),
		Phone:              res.Get("178.0.3").Str,
		OpenHours:          openingHours,
		OpenHoursStr:       openingHours.String(),
		Status:             res.Get("203.1.8.0").Str,
		ReservationUrls:    parseStringList(res.Get("46.#.0")),
		OrderUrls:          parseStringList(res.Get("75.0.1.2.#.1.2.0")),
		OrderPlatforms:     parseStringList(res.Get("75.0.1.2.#.0.0")),
		GoogleReserveUrl:   unicodeDecode(res.Get("75.0.0.5.1.2.0").Str),
		Details:            parseDetails(res.Get("100.1")),
	}
}

func parseOpeningHours(raw gjson.Result) OpeningHours {
	oh := make(OpeningHours)
	for _, opening := range raw.Array() {
		oh[opening.Get("0").Str] = opening.Get("3.0.0").Str
	}
	return oh
}

func parseDetails(raw gjson.Result) Details {
	details := make(Details)
	for _, section := range raw.Array() {
		sectionName := section.Get("0").Str
		details[sectionName] = map[string]string{}
		for _, item := range section.Get("2").Array() {
			details[sectionName][item.Get("1").Str] = item.Get("2.2.3").Str
		}
	}
	return details
}

func parseReviewDistribution(raw gjson.Result) map[string]uint64 {
	dist := make(map[string]uint64, 5)
	for i, rating := range raw.Array() {
		dist[fmt.Sprintf("%d", i+1)] = rating.Uint()
	}
	return dist
}

func parseStringList(raw gjson.Result) []string {
	var list []string
	for _, item := range raw.Array() {
		list = append(list, unicodeDecode(item.Str))
	}
	return list
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

	latlng := gjson.GetBytes(body, "0.1.0.22.11")
	lat = latlng.Get("2").Float()
	lng = latlng.Get("3").Float()

	if lat == 0 && lng == 0 {
		latlng = gjson.GetBytes(body, "0.1.1.22.11")
		lat = latlng.Get("2").Float()
		lng = latlng.Get("3").Float()
	}

	qResult := gjson.GetBytes(body, "0.0")
	if !(latlng.Exists() && qResult.Exists()) {
		return 0, 0, "", false
	}

	return lat, lng, qResult.Str, true
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

