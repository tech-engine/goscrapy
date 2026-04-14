package google_maps_scraper

// id field is compulsory in a Job defination. You can add your custom to Job
type Job struct {
	id         string
	query      string
	cursor     int
	maxRecords uint32
	loc        *location
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
func (j *Job) SetCursor(cursor int) *Job {
	j.cursor = cursor
	return j
}

func (j *Job) SetQuery(query string) *Job {
	j.query = query
	return j
}

func (j *Job) setLocation(lat, lng float64) *Job {

	if j.loc == nil {
		j.loc = &location{}
	}
	j.loc.lat = lat
	j.loc.lng = lng
	return j
}

func (j *Job) SetMaxRecords(max uint32) *Job {
	j.maxRecords = max
	return j
}
