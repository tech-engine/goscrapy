package scrapeThisSite

/*
   json and csv struct field tags are required, if you want the Record to be exported
   or processed by builtin pipelines
*/

type Record struct {
	Title       string `json:"title" csv:"title"`
	Year        int    `json:"year" csv:"year"`
	Awards      int    `json:"awards" csv:"awards"`
	Nominations int    `json:"nominations" csv:"nominations"`
	BestPicture bool   `json:"best_picture,omitempty" csv:"best_picture,omitempty"`
}
