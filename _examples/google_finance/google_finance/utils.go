package google_finance

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/tech-engine/goscrapy/pkg/builtin/gosm"
	"github.com/tidwall/gjson"
)

// buildBatchRequest constructs the hex-delimited JSON body for Google Finance RPC calls.
func buildBatchRequest(ticker string) (string, string) {
	isCrypto := strings.Contains(ticker, "-") && !strings.Contains(ticker, ":")

	// Helper to create the ticker tuple used in RPC requests.
	var t []any
	if isCrypto {
		parts := strings.Split(ticker, "-")
		t = []any{nil, nil, []string{parts[0], parts[1]}}
	} else {
		parts := strings.Split(ticker, ":")
		if len(parts) < 2 {
			t = []any{nil, []string{ticker, ""}}
		} else {
			t = []any{nil, []string{parts[0], parts[1]}}
		}
	}

	nBEQBcType := 5
	if isCrypto {
		nBEQBcType = 6
	}

	requests := []struct {
		id  string
		req []any
	}{
		{id: "xh8wxf", req: []any{[]any{t}, 1}},
		{id: "HqGpWd", req: []any{[]any{t}}},
		{id: "Pr8h2e", req: []any{[]any{t}}},
		{id: "AiCwsd", req: []any{[]any{t}, 3}},
		{id: "nBEQBc", req: []any{nBEQBcType, 3, []any{t}}},
	}

	rpcids := ""
	var batchData [][]any
	for i, r := range requests {
		if i > 0 {
			rpcids += ","
		}
		rpcids += r.id
		reqJSON, _ := json.Marshal(r.req)
		batchData = append(batchData, []any{r.id, string(reqJSON), nil, fmt.Sprintf("%d", i+1)})
	}

	outerJSON, _ := json.Marshal([]any{batchData})
	return fmt.Sprintf("f.req=%s", url.QueryEscape(string(outerJSON))), rpcids
}

// parseBatchResults extracts individual RPC results from Google's chunked response format.
func parseBatchResults(data []byte) map[string]gjson.Result {
	stripped := string(data)
	if idx := strings.Index(stripped, "\n"); idx != -1 && strings.HasPrefix(stripped, ")]}'") {
		stripped = stripped[idx+1:]
	}

	results := make(map[string]gjson.Result)
	lines := strings.Split(stripped, "\n")
	hexRegex := regexp.MustCompile(`^[0-9a-fA-F]+$`)

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" && hexRegex.MatchString(line) && i+1 < len(lines) {
			gjson.Parse(lines[i+1]).ForEach(func(_, entry gjson.Result) bool {
				if entry.Get("0").String() == "wrb.fr" {
					results[entry.Get("1").String()] = gjson.Parse(entry.Get("2").String())
				}
				return true
			})
			i++
		}
	}
	return results
}

// mapQuote extracts basic ticker information from the xh8wxf response.
func mapQuote(ticker string, raw gjson.Result, record *Record) {
	// Automatically map fields defined by gjson tags in record.go
	_ = gosm.Map(raw, record)

	// Special handling for crypto ticker/exchange locations
	if strings.Contains(ticker, "-") && !strings.Contains(ticker, ":") {
		record.Ticker = raw.Get("0.0.0.21").String()
		record.Exchange = ""
	}

	// Map numeric type to human-readable string
	typeMap := map[float64]string{0: "stock", 1: "index", 3: "crypto", 5: "etf"}
	if tStr, ok := typeMap[raw.Get("0.0.0.3").Float()]; ok {
		record.Type = tStr
	}
}

// mapCompany extracts detailed organizational info from the HqGpWd response.
func mapCompany(raw gjson.Result, record *Record) {
	// Automatically map all company fields using tags
	_ = gosm.Map(raw, record)
}

// mapChart extracts historical performance data from the AiCwsd response.
func mapChart(raw gjson.Result, record *Record) {
	chart := raw.Get("0.0")
	if !chart.Exists() {
		return
	}

	var points []ChartPoint
	chart.Get("3").ForEach(func(_, period gjson.Result) bool {
		period.Get("1").ForEach(func(_, pt gjson.Result) bool {
			dateArr := pt.Get("0").Array()
			priceArr := pt.Get("1").Array()
			if len(dateArr) >= 3 && len(priceArr) >= 1 {
				var ptRec ChartPoint
				_ = gosm.Map(pt, &ptRec)
				ptRec.Date = fmt.Sprintf("%04.0f-%02.0f-%02.0f", dateArr[0].Float(), dateArr[1].Float(), dateArr[2].Float())
				points = append(points, ptRec)
			}
			return true
		})
		return true
	})

	record.Chart = &ChartData{
		PreviousClose: chart.Get("6").Float(),
		Points:        points,
	}
}

// mapNews extracts recent related news articles from the nBEQBc response.
func mapNews(raw gjson.Result, record *Record) {
	news := raw.Get("0")
	if !news.Exists() || !news.IsArray() {
		return
	}

	news.ForEach(func(_, item gjson.Result) bool {
		if item.IsArray() && item.Get("1").Exists() {
			var ni NewsItem
			_ = gosm.Map(item, &ni)
			record.News = append(record.News, ni)
		}
		return true
	})
}
