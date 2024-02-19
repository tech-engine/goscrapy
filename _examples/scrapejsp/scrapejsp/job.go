package scrapejsp

// id field is compulsory in a Job defination. You can add your custom to Job
type Job struct {
	id    string
	// query string
}

// do not delete/edit
func NewJob(id string) *Job {
	return &Job{
		id: id,
	}
}

// do not delete/edit
func (j *Job) Id() string {
	return j.id
}

// do not delete
func (j *Job) Reset() {
	j.id = ""
}

// add your custom receiver functions below
// func (j *Job) SetQuery(query string) {
// 	j.query = query
// 	return
// }
