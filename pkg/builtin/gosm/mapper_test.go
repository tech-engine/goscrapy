package gosm

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/worker"
	"github.com/tidwall/gjson"
)

type MockResponse struct {
	core.ISelector
	body []byte
}

func (m *MockResponse) Header() http.Header          { return nil }
func (m *MockResponse) Body() io.ReadCloser          { return nil }
func (m *MockResponse) Bytes() []byte                { return m.body }
func (m *MockResponse) StatusCode() int              { return 200 }
func (m *MockResponse) Cookies() []*http.Cookie      { return nil }
func (m *MockResponse) Request() *http.Request       { return nil }
func (m *MockResponse) Meta(string) (any, bool)      { return nil, false }
func (m *MockResponse) Detach() core.IResponseReader { return m }

func NewMockResponse(htmlContent string) *MockResponse {
	sel, _ := worker.NewSelector(strings.NewReader(htmlContent))
	return &MockResponse{
		ISelector: sel,
		body:      []byte(htmlContent),
	}
}

var testHTML = `
<html>
	<body>
		<h1>The Go Programming Language</h1>
		<span class="price">45.99</span>
		<div class="category">Education</div>
		<p class="author">Alan Donovan</p>
		<a class="buy-link" href="https://example.com/buy">Buy Now</a>
		<p class="star-rating Three"></p>
		<ul class="tags">
			<li class="tag">Programming</li>
			<li class="tag">Go</li>
			<li class="tag">Google</li>
		</ul>
	</body>
</html>
`

func TestMap_CSS_SingleValue(t *testing.T) {
	type Record struct {
		Title    string `gos_css:"h1"`
		Category string `gos_css:".category"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, "The Go Programming Language", r.Title)
	assert.Equal(t, "Education", r.Category)
}

func TestMap_CSS_Slice(t *testing.T) {
	type Record struct {
		Tags []string `gos_css:".tag"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Programming", "Go", "Google"}, r.Tags)
}

func TestMap_XPath_SingleValue(t *testing.T) {
	type Record struct {
		Price  float64 `gos_xpath:"//span[@class='price']"`
		Author string  `gos_xpath:"//p[@class='author']"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, 45.99, r.Price)
	assert.Equal(t, "Alan Donovan", r.Author)
}

func TestMap_XPath_Slice(t *testing.T) {
	type Record struct {
		Tags []string `gos_xpath:"//li[@class='tag']"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Programming", "Go", "Google"}, r.Tags)
}

func TestMap_CSS_XPath_Mixed(t *testing.T) {
	type Record struct {
		Title  string  `gos_css:"h1"`
		Price  float64 `gos_xpath:"//span[@class='price']"`
		Author string  `gos_xpath:"//p[@class='author']"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, "The Go Programming Language", r.Title)
	assert.Equal(t, 45.99, r.Price)
	assert.Equal(t, "Alan Donovan", r.Author)
}

func TestMap_CSS_Attribute(t *testing.T) {
	type Record struct {
		Link   string `gos_css:".buy-link@href"`
		Rating string `gos_css:".star-rating@class"`
		Title  string `gos_css:"h1"` // no @attr -> Text()
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/buy", r.Link)
	assert.Equal(t, "star-rating Three", r.Rating)
	assert.Equal(t, "The Go Programming Language", r.Title)
}

func TestMap_XPath_Attribute(t *testing.T) {
	type Record struct {
		Link string `gos_xpath:"//a[@class='buy-link']@href"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/buy", r.Link)
}

func TestMap_Attribute_Slices(t *testing.T) {
	type Record struct {
		AllTags []string `gos_css:".tag@class"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Equal(t, []string{"tag", "tag", "tag"}, r.AllTags)
}

func TestMap_Attribute_Missing(t *testing.T) {
	type Record struct {
		MissingAttr string `gos_css:"h1@nonexistent"`
		NoSelector  string `gos_css:"@href"`
	}

	var r Record
	err := Map(NewMockResponse(testHTML), &r)
	assert.NoError(t, err)
	assert.Empty(t, r.MissingAttr)
	assert.Empty(t, r.NoSelector)
}

func TestMap_JSON_ScalarTypes(t *testing.T) {
	src := gjson.Parse(`{"name":"Alice","age":30,"score":98.5,"active":true}`)

	type Record struct {
		Name   string  `gos:"name"`
		Age    int     `gos:"age"`
		Score  float64 `gos:"score"`
		Active bool    `gos:"active"`
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", r.Name)
	assert.Equal(t, 30, r.Age)
	assert.Equal(t, 98.5, r.Score)
	assert.True(t, r.Active)
}

func TestMap_JSON_MissingKey(t *testing.T) {
	src := gjson.Parse(`{"name":"Alice"}`)

	type Record struct {
		Name    string `gos:"name"`
		Missing string `gos:"nonexistent"`
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", r.Name)
	assert.Empty(t, r.Missing)
}

func TestMap_JSON_NestedStruct(t *testing.T) {
	src := gjson.Parse(`{"top":"hello","nested":{"a":"world","b":42}}`)

	type Inner struct {
		A string `gos:"a"`
		B int    `gos:"b"`
	}
	type Record struct {
		Top   string `gos:"top"`
		Inner Inner  `gos:"nested"`
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.Equal(t, "hello", r.Top)
	assert.Equal(t, "world", r.Inner.A)
	assert.Equal(t, 42, r.Inner.B)
}

func TestMap_JSON_NestedPtrStruct(t *testing.T) {
	src := gjson.Parse(`{"data":{"val":"ok"}}`)

	type Inner struct {
		Val string `gos:"val"`
	}
	type Record struct {
		Inner *Inner `gos:"data"`
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.NotNil(t, r.Inner)
	assert.Equal(t, "ok", r.Inner.Val)
}

func TestMap_JSON_ArraySlice(t *testing.T) {
	src := gjson.Parse(`{"items":[{"name":"A","price":1.5},{"name":"B","price":2.5},{"name":"C","price":3.5}]}`)

	type Record struct {
		Prices []float64 `gos:"items.#.price"`
		Names  []string  `gos:"items.#.name"`
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.Equal(t, []float64{1.5, 2.5, 3.5}, r.Prices)
	assert.Equal(t, []string{"A", "B", "C"}, r.Names)
}

func TestMap_UnexportedFieldsSkipped(t *testing.T) {
	src := gjson.Parse(`{"name":"visible","secret":"hidden"}`)

	type Record struct {
		Public  string `gos:"name"`
		private string `gos:"secret"` //nolint:unused
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.Equal(t, "visible", r.Public)
	assert.Empty(t, r.private)
}

func TestMap_RawBytesSource(t *testing.T) {
	type Record struct {
		X int `gos:"x"`
	}

	var r Record
	err := Map([]byte(`{"x":99}`), &r)
	assert.NoError(t, err)
	assert.Equal(t, 99, r.X)
}

func TestMap_RawStringSource(t *testing.T) {
	type Record struct {
		X int `gos:"x"`
	}

	var r Record
	err := Map(`{"x":77}`, &r)
	assert.NoError(t, err)
	assert.Equal(t, 77, r.X)
}

func TestMap_JSON_SliceOfStructs(t *testing.T) {
	src := gjson.Parse(`{"items":[{"name":"A","val":1},{"name":"B","val":2}]}`)

	type Item struct {
		Name string `gos:"name"`
		Val  int    `gos:"val"`
	}
	type Record struct {
		Items []Item `gos:"items"`
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.Len(t, r.Items, 2)
	assert.Equal(t, "A", r.Items[0].Name)
	assert.Equal(t, 1, r.Items[0].Val)
	assert.Equal(t, "B", r.Items[1].Name)
	assert.Equal(t, 2, r.Items[1].Val)
}

func TestMap_JSON_RecursiveArrayMapping(t *testing.T) {
	src := gjson.Parse(`{"203": [["Monday", 1], ["Tuesday", 2]]}`)

	type DayHour struct {
		Day string `gos:"0"`
		Val int    `gos:"1"`
	}
	type Record struct {
		Hours []DayHour `gos:"203"`
	}

	var r Record
	err := Map(src, &r)
	assert.NoError(t, err)
	assert.Len(t, r.Hours, 2)
	assert.Equal(t, "Monday", r.Hours[0].Day)
	assert.Equal(t, 1, r.Hours[0].Val)
	assert.Equal(t, "Tuesday", r.Hours[1].Day)
	assert.Equal(t, 2, r.Hours[1].Val)
}

func TestMap_NonPointerTarget_ReturnsError(t *testing.T) {
	type Record struct {
		X int `gos:"x"`
	}

	var r Record
	err := Map(`{"x":1}`, r)
	assert.Error(t, err)
}
