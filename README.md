# GoScrapy: Web Scraping Framework in Go
<p align="center">
  <img width="200" src="./logo.png">
</p>

GoScrapy is a powerful open-source web scraping framework written in the Go programming language. It offers a user-friendly interface for extracting data from websites, making it an ideal tool for various data collection and analysis tasks.

## Getting Started
Follow these steps to start using GoScrapy:

### 1: Project Initialization
To initialize a project, open your terminal and run the following command:

```sh
go mod init <project_name>
```
Replace <project_name> with your desired project name. For example:

```sh
go mod init scrape_this_site
```

### 2. Installation
After initialization, install the GoScrapy CLI Tool using this command:

```sh
go install github.com/tech-engine/goscrapy@latest
```
Note: This command will save the GoScrapy CLI tool on your computer, eliminating the need to run it again for future project creations.

### 3. Verify Installation
To verify a successful installation, check the version of GoScrapy using the following command:

```sh
goscrapy --version
```
### 4. Creating a New Project
Create a new project using GoScrapy by following below steps:

```sh
goscrapy startproject <project_name>
```
Replace <project_name> with your chosen project name. For example:

```sh
goscrapy startproject scrape_this_site
```
This command will create a new project directory with the necessary files and structure to begin working with GoScrapy.

```sh
PS D:\My-Projects\go\go-test-scrapy> goscrapy startproject scrapethissite

🚀 GoScrapy generating project files. Please wait!

✔️  scrapethissite\constants.go
✔️  scrapethissite\core.go
✔️  scrapethissite\errors.go
✔️  scrapethissite\job.go
✔️  scrapethissite\output.go
✔️  scrapethissite\spider.go
✔️  scrapethissite\types.go

✨ Congrates. scrapethissite created successfully.
```

## Usage
### Defining a Scraping Task
GoScrapy operates around the below three concept.

- **[Job](#job):** Describes an input to the spider.
- **[Output](#output):** Represents an output produced by the spider.
- **[Spider](#spider):** Contains the main logic of the scraper.


### Job
The Job represents an input to the scraper. It defines the specific task for the scraping process. In the provided code __`job.go`__, a Job struct is defined by fields like id and query. The id field is compulsory, and you can add custom fields to the Job structure.

```go
// id field is compulsory in a Job defination. You can add your custom to Job
type Job struct {
	id string
	query string // custom field
}

// add your custom receiver functions below
func (j *Job) SetQuery(query string) {
	j.query = query
	return
}
```

### Output
The Output represents the output produced by the spider. It encapsulates the records obtained from scraping, any potential errors, and a reference to the associated Job. The Output struct, as defined in the __`output.go`__ code, contains methods to retrieve records, error information, and other details. The Records, Error, Job, and IsEmpty functions provide access to various aspects of the scraping output.

```go
// do not modify this file
type Output struct {
	records []Record
	err     error
	job     *Job
}
```

### Spider
The Spider encapsulate the main logic of the scraper from the initiation of requests, parsing of responses, and data extraction.
<!-- Here goes the spider.go -->

## Example
This example illustrates how to utilize the GoScrapy framework to scrape data from the website https://www.scrapethissite.com. The example covers the following files:

- **[spider.go](#spidergo---spider-creation)** - Spider Creation
- **[types.go](#typesgo---data-structure-definition)** - Data Structure Definition
- **[main.go](#maingo---spider-execution)** - Spider Execution

### spider.go - Spider Creation
Define the spider responsible for handling the scraping logic in your spider.go file. The following code snippet sets up the spider:

```go
package scrapethissite

import (
	"context"
	"errors"
	"net/url"

	"github.com/tech-engine/goscrapy/pkg/core"
)

func NewSpider() (*Spider, error) {
	return &Spider{}, nil
}

func (s *Spider) StartRequest(ctx context.Context, job *Job) {

	// for each request we must call NewRequest() and never reuse it
	req := s.NewRequest()

	var headers map[string]string

	// GET is the default request method
	req.SetUrl(s.baseUrl.String()).
		SetMetaData("JOB", job).
		SetHeaders(headers)

		/* POST
		    req.SetUrl(s.baseUrl.String()).
		        SetMethod("POST").
				SetMetaData("JOB", job).
				SetHeaders(headers).
		        SetBody(<BODY_HERE>)
		*/

		// call the next parse method
	s.Request(ctx, req, s.parse)
}

func (s *Spider) parse(ctx context.Context, response core.ResponseReader) {
	// response.Body()
	// response.StatusCode()
	// response.Headers()
	// check output.go for the fields
	// s.yield(output)
}

```
The NewSpider returns a spider instance.

### types.go - Data Structure Definition
In your types.go file, define the data structure that corresponds to the records you're scraping. Here's the structure for the Record type:

```go
/*
   json and csv struct field tags are required, if you want the Record to be exported
   or processed by builtin pipelines
*/

type Record struct {
	Title string `json:"title" csv:"title"`
}
```

The Record struct defines the fields that match the JSON data structure from the website.

### main.go - Spider Execution
In your __`main.go`__ file, set up and execute your spider using the GoScrapy framework by following these steps:

For an illustrative example, you can refer to the [sample code here](./_examples/scrapethissite/main.go).

## Pipelines - Managing and Transforming Scraped Data
In the GoScrapy framework, pipelines play a pivotal role in managing, transforming, and fine-tuning the scraped data to meet your project's specific needs. Pipelines provide a powerful mechanism for executing a sequence of actions that are executed on the scraped data.

The concept of pipelines in GoScrapy brings a versatile toolkit that enables you to process the scraped data efficiently and flexibly. These pipelines not only enhance the quality and structure of the collected data but also empower you to streamline data processing operations.

### Built-in Pipelines
GoScrapy offers a range of built-in pipelines, designed to facilitate different aspects of data manipulation and organization. Some of the noteworthy built-in pipelines include:

- Export2CSV
- Export2JSON
- Export2GSHEET

### Incorporating Pipelines into Your Scraping Workflow
To seamlessly integrate pipelines into your scraping workflow, you can utilize the Pipelines.Add method.

Here is an example on how you can add pipelines to your scraping process:

Export to JSON Pipeline:

```go
// goScrapy instance
goScrapy.Pipelines.Add(pipelines.Export2JSON[*customProject.Job, []customProject.Record]())
```