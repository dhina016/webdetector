# WebDetector

WebDetector is a Go-based tool designed to identify web technologies used by websites. It analyzes web pages to detect frameworks and other web technologies based on predefined patterns and regular expressions.

## Features

- Single URL and bulk URL processing from a text file
- Supports various output formats including plain text, JSON, and CSV
- Configurable HTTP request options (timeout, follow redirects, SSL verification)
- Embedded pattern JSON for easy pattern updates

## Prerequisites

Before you begin, ensure you have installed Go (version 1.14 or later recommended) on your machine. You can download and install Go from [Go's official site](https://golang.org/dl/).

## Installation

Clone the repository to your local machine:

```bash
git clone https://github.com/yourusername/webdetector.git
cd webdetector
go build -o webdetector
```
or 
```bash
go install github.com/dhina016/webdetector@latest
```


## Usage

### Command-Line Options

- `-u`: Specify a single URL to check
- `-l`: Specify a file containing a list of domains
- `-of`: Output format (options: text, json, csv)
- `-o`: Output file name
- `-to`: Timeout duration for HTTP requests in seconds
- `-s`: Enable strict certificate verification
- `-fd`: Follow redirects if the domain is redirecting
- `-log`: Enable error logging

### Examples

**Note: Always use -fd flag for accurate result**

Check a single URL and print the result in JSON format:

```bash
./webdetector -fd -u http://example.com -of json
```

Process a list of URLs from a file, follow redirects, and write output in CSV format to a file:

```bash
./webdetector -fd -l urls.txt -fd -of csv -o results.csv
```

## Running with Docker

If you prefer to use Docker, you can build a Docker image using the provided Dockerfile:

```bash
docker build -t webdetector .
```

Run WebDetector inside a Docker container:

```bash
docker run --rm -v $(pwd):/data webdetector -fd -l /data/urls.txt -of json
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request.
