package google_maps_scraper

import (
	"reflect"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/tech-engine/goscrapy/pkg/core"
)

/*
   json and csv struct field tags are required, if you want the Record to be exported
   or processed by builtin pipelines
*/

type Record struct {
	J *Job `json:"-" csv:"-"` // JobId is required
	// add you custom fields here
	Query              string           `json:"query" csv:"query"`
	QueryLoc           location         `json:"-" csv:"-"`
	Status             string           `json:"status" csv:"status"`
	OpenHours          OpeningHours     `json:"opening_hours" csv:"-"`
	OpenHoursStr       string           `json:"opening_hours_str" csv:"opening_hours_str"`
	Phone              string           `json:"phone" csv:"phone"`
	WebResultsUrl      string           `json:"webresults_url" csv:"webresults_url"`
	ShortDescription   string           `json:"short_description" csv:"short_description"`
	Description        string           `json:"description" csv:"description"`
	TimeZone           string           `json:"timezone" csv:"timezone"`
	Categories         []string         `json:"category" csv:"category"`
	Title              string           `json:"title" csv:"title"`
	Website            string           `json:"website" csv:"website"`
	Review             float64          `json:"review" csv:"review"`
	ReviewDistribution string2Uint64Map `json:"review_distribution" csv:"review_distribution"`
	ReviewsUrl         string           `json:"reviews_url" csv:"reviews_url"`
	ReviewsCount       uint64           `json:"reviews_count" csv:"reviews_count"`
	Street             string           `json:"street" csv:"street"`
	City               string           `json:"city" csv:"city"`
	ZipCode            string           `json:"zipcode" csv:"zipcode"`
	State              string           `json:"state" csv:"state"`
	StateCode          string           `json:"state_code" csv:"state_code"`
	Country            string           `json:"country" csv:"country"`
	CountryCode        string           `json:"country_code" csv:"country_code"`
	Details            Details          `json:"details" csv:"details"`
	ReservationUrls    []string         `json:"reservation_urls" csv:"reservation_urls"`
	OrderUrls          []string         `json:"order_urls" csv:"order_urls"`
	OrderPlatforms     []string         `json:"order_platforms" csv:"order_platforms"`
	GoogleReserveUrl   string           `json:"google_reserve_url" csv:"google_reserve_url"`
}

// modify below code only if you know what you are doing
func (r *Record) Record() *Record {
	return r
}

func (r *Record) RecordKeys() []string {
	dataType := reflect.TypeOf(*r)
	if dataType.Kind() != reflect.Struct {
		panic("Record is not a struct")
	}

	numFields := dataType.NumField()
	keys := make([]string, numFields)

	for i := 0; i < numFields; i++ {
		field := dataType.Field(i)
		csvTag := field.Tag.Get("csv")
		keys[i] = csvTag
	}

	return keys
}

func (r *Record) RecordFlat() []any {

	inputType := reflect.TypeOf(*r)

	if inputType.Kind() != reflect.Struct {
		panic("Record is not a struct")
	}

	inputValue := reflect.ValueOf(*r)

	slice := make([]any, inputType.NumField())

	for i := 0; i < inputType.NumField(); i++ {
		slice[i] = inputValue.Field(i).Interface()
	}
	return slice
}

func (r *Record) Job() core.IJob {
	return r.J
}

func (d Details) MarshalCSV() (string, error) {
	return convertor.ToString(d), nil
}

func (d string2Uint64Map) MarshalCSV() (string, error) {
	return convertor.ToString(d), nil
}

type Details map[string]map[string]string
type string2Uint64Map map[string]uint64

type OpeningHours map[string]string

type location struct {
	lat float64
	lng float64
}

// opening hours receiver function
func (o OpeningHours) String() string {
	tmp := make([]string, 0)
	for day, timing := range o {
		tmp = append(tmp, day+":"+timing)
	}

	return strings.Join(tmp, " | ")
}
