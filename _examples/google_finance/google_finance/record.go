package google_finance

import "github.com/tech-engine/goscrapy/pkg/core"

type Record struct {
	J              *Job           `json:"-"`
	Ticker         string       `json:"ticker" gos:"0.0.0.1.0"` // Note: this needs conditional logic for crypto, but let's see
	Exchange       string       `json:"exchange" gos:"0.0.0.1.1"`
	Name           string       `json:"name" gos:"0.0.0.2"`
	Type           string       `json:"type" gos:"0.0.0.3"`
	Price          float64      `json:"price" gos:"0.0.0.5.0"`
	Change         float64      `json:"change" gos:"0.0.0.5.1"`
	ChangePercent  float64      `json:"change_percent" gos:"0.0.0.5.2"`
	PreviousClose  float64      `json:"previous_close" gos:"0.0.0.7"`
	Currency       string       `json:"currency" gos:"0.0.0.4"`
	Timezone       string       `json:"timezone" gos:"0.0.0.12"`
	Company        *CompanyInfo `json:"company" gos:"0.0"`
	Chart          *ChartData   `json:"chart"`
	News           []NewsItem   `json:"news"`
}

type CompanyInfo struct {
	Description      string  `json:"description" gos:"2"`
	CEO              string  `json:"ceo" gos:"5"`
	Employees        string  `json:"employees" gos:"6"`
	MarketCap        string  `json:"market_cap" gos:"7"`
	Open             float64 `json:"open" gos:"9"`
	High             float64 `json:"high" gos:"10"`
	Low              float64 `json:"low" gos:"11"`
	FiftyTwoWeekHigh float64 `json:"fifty_two_week_high" gos:"12"`
	FiftyTwoWeekLow  float64 `json:"fifty_two_week_low" gos:"13"`
	PERatio          float64 `json:"pe_ratio" gos:"16"`
	Volume           string  `json:"volume" gos:"18"`
	Sector           string  `json:"sector" gos:"71"`
}

type ChartData struct {
	PreviousClose float64      `json:"previous_close"`
	Points        []ChartPoint `json:"points"`
}

type ChartPoint struct {
	Date   string  `json:"date"`
	Price  float64 `json:"price" gos:"1.0"`
	Volume float64 `json:"volume" gos:"2"`
}

type NewsItem struct {
	Title     string `json:"title" gos:"1"`
	Source    string `json:"source" gos:"2"`
	URL       string `json:"url" gos:"0"`
	Timestamp int64  `json:"timestamp" gos:"4"`
}

func (r *Record) Record() *Record { return r }
func (r *Record) RecordKeys() []string   { return nil }
func (r *Record) RecordFlat() []any      { return nil }
func (r *Record) Job() core.IJob         { return r.J }
