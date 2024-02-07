package scrapejsp

/*
   json and csv struct field tags are required, if you want the Record to be exported
   or processed by builtin pipelines
*/

type Record struct {
	UserId    int    `csv:"userId" json:"userId"`
	Id        int    `csv:"id" json:"id"`
	Title     string `csv:"title" json:"title"`
	Completed bool   `csv:"completed" json:"completed"`
}
