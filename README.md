# GoScrapy: Web Scraping Framework in Go
 [![Go Reference](https://pkg.go.dev/badge/github.com/tech-engine/goscrapy.svg)](https://pkg.go.dev/github.com/tech-engine/goscrapy) [![Alt Text](https://goreportcard.com/badge/github.com/tech-engine/goscrapy)](https://github.com/tech-engine/goscrapy)
<p align="center">
  <img width="800" src="./logo.webp">
</p>

**GoScrapy** aims to be a powerful web scraping framework in Go, inspired by Python's Scrapy framework. It offers an easy-to-use Scrapy-like experience for extracting data from websites, making it an ideal tool for various data collection and analysis tasks, especially for those coming from Python and wanting to try scraping in Golang..

## Getting Started

Goscrapy is tested with **go v1.21**

### 1: Project Initialization

```sh
go mod init books_to_scrape
```

### 2. Install goscrapy cli

```sh
go install github.com/tech-engine/goscrapy@latest
```
**Note**: make sure to always keep your goscrapy cli updated.

### 3. Verify Installation

```sh
goscrapy -v
```
### 4. Create a New Project

```sh
goscrapy startproject books_to_scrape
```
This will create a new project directory with all the files necessary to begin working with **GoScrapy**.

```sh
\iyuioy\go\go-test-scrapy> goscrapy startproject books_to_scrape

🚀 GoScrapy generating project files. Please wait!

✔️  books_to_scrape\constants.go
✔️  books_to_scrape\errors.go
✔️  books_to_scrape\job.go
✔️  main.go
✔️  books_to_scrape\record.go
✔️  books_to_scrape\spider.go

✨ Congrates. books_to_scrape created successfully.
```

### main.go
In your __`main.go`__ file, set up and execute your spider.

For detailed code, please refer to the [sample code here](./_examples/books.toscrape.com/main.go).

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"books_to_scrape/books_to_scrape"
	"sync"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	spider, errCh := books_to_scrape.New(ctx)
	
	go func() {
		defer wg.Done()

		err := <-errCh

		if err != nil && errors.Is(err, context.Canceled) {
			return
		}

		fmt.Printf("failed: %q", err)
	}()

	// trigger the Start Request
	spider.StartRequest(ctx, nil)

	OnTerminate(func() {
		fmt.Println("exit signal received: shutting down gracefully")
		cancel()
		wg.Wait()
	})
}
```

## Wiki
Please follow the [wiki](https://github.com/tech-engine/goscrapy/wiki) docs for details.

### Note

**GoScrapy** is not stable, so its API may change drastically. Please exercise caution when using it in production.

## License

**GoScrapy** is available under BSL with additional usage grant which allows for free internal use. Please make sure that you agree with the license before contributing to **GoScrapy** because by contributing to goscrapy project you are agreeing on the license.

## Roadmap

- ~~Cookie management~~
- ~~Builtin & Custom Middlewares support~~
- ~~Css & Xpath Selectors~~
- Logging
- Triggers
- Tests(work in progress)

## Get in touch
[Discord](https://discord.gg/FPvxETjYPH)
