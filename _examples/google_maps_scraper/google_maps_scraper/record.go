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
	Query            string         `json:"query" csv:"query"`
	QueryLoc         location       `json:"-" csv:"-"`
	Status           string         `json:"status" csv:"status" gos:"203.1.8.0"`
	OpenHours        []DayOpening   `json:"opening_hours" csv:"-" gos:"203.0"`
	OpenHoursStr     string         `json:"opening_hours_str" csv:"opening_hours_str"`
	Phone            string         `json:"phone" csv:"phone" gos:"178.0.3"`
	WebResultsUrl    string         `json:"webresults_url" csv:"webresults_url" gos:"174.0"`
	ShortDescription string         `json:"short_description" csv:"short_description" gos:"32.0.1"`
	Description      string         `json:"description" csv:"description" gos:"32.1.1"`
	TimeZone         string         `json:"timezone" csv:"timezone" gos:"30"`
	Categories       []string       `json:"category" csv:"category" gos:"13"`
	Title            string         `json:"title" csv:"title" gos:"11"`
	Website          string         `json:"website" csv:"website" gos:"7.1"`
	Review             float64          `json:"review" csv:"review" gos:"4.7"`
	ReviewDistribution string2Uint64Map `json:"review_distribution" csv:"review_distribution"`
	ReviewsUrl         string           `json:"reviews_url" csv:"reviews_url" gos:"4.3.0"`
	ReviewsCount     uint64         `json:"reviews_count" csv:"reviews_count" gos:"4.8"`
	Street           string         `json:"street" csv:"street" gos:"2.0"`
	City             string         `json:"city" csv:"city" gos:"183.1.3"`
	ZipCode          string         `json:"zipcode" csv:"zipcode" gos:"183.1.4"`
	State            string         `json:"state" csv:"state" gos:"183.1.5"`
	StateCode        string         `json:"state_code" csv:"state_code"` // manual split
	Country          string         `json:"country" csv:"country" gos:"2.2"`
	CountryCode      string         `json:"country_code" csv:"country_code" gos:"183.1.6"`
	Details          []DetailSection `json:"details" csv:"-" gos:"100.1"`
	ReservationUrls  []string       `json:"reservation_urls" csv:"reservation_urls" gos:"46.#.0"`
	Orders           []OrderInfo    `json:"orders" csv:"-" gos:"75.0.1.2"`
	GoogleReserveUrl string         `json:"google_reserve_url" csv:"google_reserve_url" gos:"75.0.0.5.1.2.0"`

	// old fields kept for csv compatibility if needed, but mapped from sub-structs in post
	OrderUrls      []string `json:"order_urls" csv:"order_urls"`
	OrderPlatforms []string `json:"order_platforms" csv:"order_platforms"`
}

type OrderInfo struct {
	Platform string `gos:"0.0"`
	Url      string `gos:"1.2.0"`
}

type DayOpening struct {
	Day    string `gos:"0"`
	Timing string `gos:"3.0.0"`
}

type DetailItem struct {
	Label string `gos:"1"`
	Value string `gos:"2.2.3"`
}

type DetailSection struct {
	Name  string       `gos:"1"`
	Items []DetailItem `gos:"2"`
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
