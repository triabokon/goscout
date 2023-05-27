# Goscout

## Overview

Goscout is a simple, concurrent web crawler written in Go.

Given a starting URL, it visits and collects each URL on the same domain, it doesn't follow external links. Upon completion, it generates a sitemap of collected URLs.

The project is structured into three main modules:

1. **Crawler**: concurrently visits web pages on the same domain with the provided site URL.
2. **Parser**: parses web pages and extracts URLs from their HTML.
3. **Sitemap**: generates a sitemap from the collected URLs and writes it to the file.

## Getting Started

The project uses Go Modules for dependency management.
All the dependencies for the application are specified in the `go.mod` file.
Dependencies that are needed for linting and testing are specified in the `tools/go.mod`.

### Prerequisites

Make sure you have installed Go (version 1.20) on your system. Refer to the [official Go installation guide](https://golang.org/doc/install) if needed.

### Installation

Clone the repository:

```bash
git clone https://github.com/triabokon/goscout.git
```

Then, navigate into the cloned directory:

```bash
cd goscout
```

## Building the Project

Goscout uses Makefile for simplifying build commands. You can build the project using the following steps:

1. Download dependencies:
```bash
make download-deps
```
*Note: this will download all needed dependencies: for the application, mocks and linting.*

2. Build the project:
```bash
make build
```
This will build the binary and store it in the `./bin` directory.

## Running goscout
After building the project, you can run the binary:

```bash
./bin/goscout --site_url https://monzo.com/
```

Goscout has a CLI and there is `help` command output.

```bash
./bin/goscout --help
```

```
GoScout is a simple web-crawler tool.

Usage:
  goscout [flags]

Aliases:
  goscout, gs

Flags:
      --check_interval duration    time interval to check if there are any pages left to crawl (default 1s)
      --crawler_depth int          maximum depth the crawler would go (default 100)
      --crawler_queue_size int     maximum number of tasks that queue can store (default 100)
      --crawler_worker_count int   number of workers for crawler (default 100)
      --file_name string           filename to write sitemap (default "sitemap.xml")
  -h, --help                       help for goscout
      --http_timeout duration      timeout for http requests (default 10s)
      --site_url string            url of the site to crawl (default "https://monzo.com/")
      --sitemap_indent int         xml sitemap indent (default 1)
      --sitemap_xml_ns string      xml sitemap namespace (default "https://www.sitemaps.org/schemas/sitemap/0.9/")
```

## Usage example

Launch goscout:
```bash
./bin/goscout --site_url https://www.sitemaps.org/
```

Goscout console output:
```
Start crawler with 100 workers, queue size 100 and crawling depth 100
Crawling website https://www.sitemaps.org/
..
Crawler visited 47 pages, collected 48 unique urls in 3.481728502s time
Generating sitemap ...
Writing sitemap to sitemap.xml ...
Sitemap successfully written!
```

Content of the generated `sitemap.xml` file:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="https://www.sitemaps.org/schemas/sitemap/0.9/">
    <url>
        <loc>https://www.sitemaps.org/</loc>
        <url>
            <loc>https://www.sitemaps.org/faq.php</loc>
            <url>
                <loc>https://www.sitemaps.org/index.php</loc>
                <url>
                    <loc>https://www.sitemaps.org/lang.js</loc>
                </url>
            </url>
            ...
        </url>
        ...
        <url>
            <loc>https://www.sitemaps.org/lang.js</loc>
        </url>
    </url>
</urlset>
```

## Testing and linting

This project uses `golangci-lint` for linting, it's configuration is specified in `.golangci.yml`.

To lint the code, you need to run command:
```bash
make lint
```

Following command runs all tests:
```bash
make test
```
Unit tests use mocks that are generated with `mockgen`.

*Note: when changing some interfaces, mocks should be updated. This could be done with `make generate` command.*

## Makefile

Makefile is used for simplifying some utils commands of the project.

Here is output of `make help` command:

```
Usage: make <TARGETS> ... <OPTIONS>

Available targets are:

    help               Show this help
    clean              Remove binaries
    download-deps      Download and install dependencies
    tidy               Perform go tidy steps
    generate           Perform go generate
    lint               Run all linters
    test               Run unit tests
    build              Compile packages and dependencies
```

## Possible improvements

Some other things could be done to improve goscout:

1. URL validation could be improved. Goscout has basic URL validation, 
so it could fail when finding malformed URLs or URLs with escaping symbols, for example.
2. A retries system could be introduced for the HTTP client when fetching web pages, 
so goscout could potentially collect more URLs and become more fault-tolerant.
3. Profiling and benchmarking could be used to test and potentially find some memory or concurrency-related issues.
4. Test coverage could be improved
```bash
github.com/triabokon/goscout/internal/crawler   coverage: 60.9% of statements
github.com/triabokon/goscout/internal/parser    coverage: 85.1% of statements
github.com/triabokon/goscout/internal/sitemap   coverage: 48.6% of statements
```