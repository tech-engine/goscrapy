package google_finance

import "github.com/tech-engine/goscrapy/pkg/core"

type Record struct {
	J              *Job           `json:"-"`
	Ticker         string       `json:"ticker"`
	Exchange       string       `json:"exchange"`
	Name           string       `json:"name"`
	Type           string       `json:"type"`
	Price          float64      `json:"price"`
	Change         float64      `json:"change"`
	ChangePercent  float64      `json:"change_percent"`
	PreviousClose  float64      `json:"previous_close"`
	Currency       string       `json:"currency"`
	Timezone       string       `json:"timezone"`
	Company        *CompanyInfo `json:"company"`
	Chart          *ChartData   `json:"chart"`
	News           []NewsItem   `json:"news"`
}

type CompanyInfo struct {
	Description      string  `json:"description"`
	CEO              string  `json:"ceo"`
	Employees        string  `json:"employees"`
	MarketCap        string  `json:"market_cap"`
	Open             float64 `json:"open"`
	High             float64 `json:"high"`
	Low              float64 `json:"low"`
	FiftyTwoWeekHigh float64 `json:"fifty_two_week_high"`
	FiftyTwoWeekLow  float64 `json:"fifty_two_week_low"`
	PERatio          float64 `json:"pe_ratio"`
	Volume           string  `json:"volume"`
	Sector           string  `json:"sector"`
}

type ChartData struct {
	PreviousClose float64      `json:"previous_close"`
	Points        []ChartPoint `json:"points"`
}

type ChartPoint struct {
	Date   string  `json:"date"`
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}

type NewsItem struct {
	Title     string `json:"title"`
	Source    string `json:"source"`
	URL       string `json:"url"`
	Timestamp int64  `json:"timestamp"`
}

func (r *Record) Record() *Record { return r }
func (r *Record) RecordKeys() []string   { return nil }
func (r *Record) RecordFlat() []any      { return nil }
func (r *Record) Job() core.IJob         { return r.J }

