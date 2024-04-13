package main

import (
    "embed"
    "bufio"
    "crypto/tls"
    "encoding/csv"
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "regexp"
    "strings"
    "time"
)

var patternJSON embed.FS

type Web struct {
    ID            string   `json:"id"`
    Patterns      []string `json:"patterns"`
    RegexPatterns []string `json:"regexPatterns"`
    Name          string   `json:"name"`
    Type          string   `json:"type"`
}

type DetectionResult struct {
    URL  string `json:"url"`
    Web  string `json:"web"`
    Type string `json:"type"`
}

func printBanner() {
    fmt.Println(`
██╗    ██╗███████╗██████╗ ██████╗ ███████╗████████╗███████╗ ██████╗████████╗ ██████╗ ██████╗ 
██║    ██║██╔════╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗
██║ █╗ ██║█████╗  ██████╔╝██║  ██║█████╗     ██║   █████╗  ██║        ██║   ██║   ██║██████╔╝
██║███╗██║██╔══╝  ██╔══██╗██║  ██║██╔══╝     ██║   ██╔══╝  ██║        ██║   ██║   ██║██╔══██╗
╚███╔███╔╝███████╗██████╔╝██████╔╝███████╗   ██║   ███████╗╚██████╗   ██║   ╚██████╔╝██║  ██║
 ╚══╝╚══╝ ╚══════╝╚═════╝ ╚═════╝ ╚══════╝   ╚═╝   ╚══════╝ ╚═════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝                                                                                           
`)
}

func loadPatterns(filename string) ([]Web, error) {
    var webList []Web
    bytes, err := patternJSON.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(bytes, &webList)
    if err != nil {
        return nil, err
    }

    return webList, nil
}

func fetchPageSource(url string, timeout time.Duration, strict bool, followRedirect bool) (string, error) {
    // Configure HTTP client
    client := &http.Client{
        Timeout: timeout,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if !followRedirect {
                return http.ErrUseLastResponse
            }
            return nil
        },
    }

    // Strict certificate verification
    if strict {
        client.Transport = &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
        }
    } else {
        client.Transport = &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
    }

    resp, err := client.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    bytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(bytes), nil
}

func detectWeb(webList []Web, pageSource string) Web {
    for _, web := range webList {
        for _, pattern := range web.Patterns {
            if strings.Contains(pageSource, pattern) {
                return web
            }
        }

        for _, regexPattern := range web.RegexPatterns {
            matched, _ := regexp.MatchString(regexPattern, pageSource)
            if matched {
                return web
            }
        }
    }
    return Web{}
}

func processURL(url string, webList []Web, timeout time.Duration, strict bool, followRedirect bool, logEnabled bool) DetectionResult {
    pageSource, err := fetchPageSource(url, timeout, strict, followRedirect)
    if err != nil {
        if logEnabled {
            // Print error immediately
            fmt.Fprintf(os.Stderr, "Error fetching page source for %s: %s\n", url, err)
        }
        return DetectionResult{URL: url, Web: fmt.Sprintf("error: %s", err)}
    }

    web := detectWeb(webList, pageSource)
    if web.Name != "" {
        return DetectionResult{URL: url, Web: web.Name, Type: web.Type}
    }
    return DetectionResult{URL: url, Web: "unknown", Type: "web"}
}

func main() {
    // Print banner
    printBanner()

    // Define flags
    list := flag.String("l", "", "Specify a file containing a list of domains")
    url := flag.String("u", "", "Specify a single URL to check")
    outputFormat := flag.String("of", "text", "Output format (text, json, csv)")
    outputFile := flag.String("o", "", "Output file name")
    timeoutSec := flag.Int("to", 10, "Timeout duration for HTTP requests in seconds")
    strict := flag.Bool("s", false, "Strict certificate verification")
    followRedirect := flag.Bool("fd", false, "Follow redirects if the domain is redirecting")
    logEnabled := flag.Bool("log", false, "Specify -log if you want error logging")
    flag.Parse()

    timeout := time.Duration(*timeoutSec) * time.Second

    // Check for required flags
    if flag.NFlag() == 0 || (*list == "" && *url == "") {
        flag.Usage()
        return
    }

    var urls []string
    if *list != "" {
        // Read URLs from the file
        file, err := os.Open(*list)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to open file: %s\n", err)
            return
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            urls = append(urls, scanner.Text())
        }
    } else if *url != "" {
        urls = append(urls, *url)
    } else {
        flag.Usage()
        return
    }

    webList, err := loadPatterns("pattern.json")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading patterns: %s\n", err)
        return
    }

    var results []DetectionResult
    for _, u := range urls {
        result := processURL(u, webList, timeout, *strict, *followRedirect, *logEnabled)
        results = append(results, result)
        // Print result immediately with formatting
        if result.Web == "unknown" || strings.HasPrefix(result.Web, "error") {
            fmt.Printf("%s \n", result.URL)
            continue
        }
        fmt.Printf("%s [%s:%s]\n", result.URL, result.Type, result.Web)
    }

    switch *outputFormat {
    case "json":
        jsonResults, _ := json.Marshal(results)
        if *outputFile != "" {
            ioutil.WriteFile(*outputFile, jsonResults, 0644)
        } else {
            fmt.Println(string(jsonResults))
        }
    case "csv":
        if *outputFile != "" {
            file, err := os.Create(*outputFile)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Failed to create output file: %s\n", err)
                return
            }
            defer file.Close()

            writer := csv.NewWriter(file)
            for _, result := range results {
                writer.Write([]string{result.URL, result.Web, result.Type})
            }
            writer.Flush()
        } else {
            writer := csv.NewWriter(os.Stdout)
            for _, result := range results {
                writer.Write([]string{result.URL, result.Web, result.Type})
            }
            writer.Flush()
        }
    default:
        // Default to text output format
        if *outputFile != "" {
            //fmt.Println("Writing text results to file:", *outputFile) // Add this line for debugging
            textResults := ""
            for _, result := range results {
                textResults += fmt.Sprintf("%s [%s:%s]\n", result.URL, result.Type, result.Web)
            }
            ioutil.WriteFile(*outputFile, []byte(textResults), 0644)
        }
    }
}
