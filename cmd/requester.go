package cmd

import (
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/cheynewallace/tabby"
	"github.com/fatih/color"
)

type Result struct {
	line          string
	statusCode    int
	contentLength int
}

func printResponse(results []Result) {
	t := tabby.New()

	var code string
	for _, result := range results {
		switch result.statusCode {
		case 200, 201, 202, 203, 204, 205, 206:
			code = color.GreenString(strconv.Itoa(result.statusCode))
		case 300, 301, 302, 303, 304, 307, 308:
			code = color.YellowString(strconv.Itoa(result.statusCode))
		case 400, 401, 402, 403, 404, 405, 406, 407, 408, 413, 429:
			code = color.RedString(strconv.Itoa(result.statusCode))
		case 500, 501, 502, 503, 504, 505, 511:
			code = color.MagentaString(strconv.Itoa(result.statusCode))
		}
		t.AddLine(code, color.BlueString(strconv.Itoa(result.contentLength)+" bytes"), result.line)
	}
	t.Print()

}

func requestMethods(uri string, headers []header, proxy *url.URL) {
	color.Cyan("\n[####] HTTP METHODS [####]")

	var lines []string
	lines, err := parseFile("payloads/httpmethods")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(lines))

	results := []Result{}

	for _, line := range lines {
		go func(line string) {
			defer wg.Done()

			statusCode, response, err := request(line, uri, headers, proxy)
			if err != nil {
				log.Fatal(err)
			}

			results = append(results, Result{line, statusCode, len(response)})
		}(line)
	}
	wg.Wait()
	printResponse(results)
}

func requestHeaders(uri string, headers []header, proxy *url.URL) {
	color.Cyan("\n[####] VERB TAMPERING [####]")

	var lines []string
	lines, err := parseFile("payloads/headers")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(lines))

	results := []Result{}

	for _, line := range lines {
		go func(line string) {
			defer wg.Done()

			h := strings.Split(line, " ")
			headers := append(headers, header{h[0], h[1]})

			statusCode, response, err := request("GET", uri, headers, proxy)
			if err != nil {
				log.Fatal(err)
			}

			results = append(results, Result{h[0] + ": " + h[1], statusCode, len(response)})
		}(line)
	}
	wg.Wait()
	printResponse(results)
}

func requestEndPaths(uri string, headers []header, proxy *url.URL) {
	color.Cyan("\n[####] CUSTOM PATHS [####]")

	var lines []string
	lines, err := parseFile("payloads/endpaths")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(lines))

	results := []Result{}

	for _, line := range lines {
		go func(line string) {
			defer wg.Done()
			statusCode, response, err := request("GET", uri+line, headers, proxy)
			if err != nil {
				log.Fatal(err)
			}

			results = append(results, Result{uri + line, statusCode, len(response)})
		}(line)
	}
	wg.Wait()
	printResponse(results)
}

func requestMidPaths(uri string, headers []header, proxy *url.URL) {
	var lines []string
	lines, err := parseFile("payloads/midpaths")
	if err != nil {
		log.Fatal(err)
	}

	h := strings.Split(uri, "/")
	var uripath string

	if uri[len(uri)-1:] == "/" {
		uripath = h[len(h)-2]
	} else {
		uripath = h[len(h)-1]
	}

	baseuri := strings.ReplaceAll(uri, uripath, "")
	baseuri = baseuri[:len(baseuri)-1]

	var wg sync.WaitGroup
	wg.Add(len(lines))

	results := []Result{}

	for _, line := range lines {
		go func(line string) {
			defer wg.Done()

			var fullpath string
			if uri[len(uri)-1:] == "/" {
				fullpath = baseuri + line + uripath + "/"
			} else {
				fullpath = baseuri + "/" + line + uripath
			}

			statusCode, response, err := request("GET", fullpath, headers, proxy)
			if err != nil {
				log.Fatal(err)
			}

			results = append(results, Result{fullpath, statusCode, len(response)})
		}(line)
	}
	wg.Wait()
	printResponse(results)
}

func requestCapital(uri string, headers []header, proxy *url.URL) {
	color.Cyan("\n[####] CAPITALIZATION [####]")

	h := strings.Split(uri, "/")
	var uripath string

	if uri[len(uri)-1:] == "/" {
		uripath = h[len(h)-2]
	} else {
		uripath = h[len(h)-1]
	}
	baseuri := strings.ReplaceAll(uri, uripath, "")
	baseuri = baseuri[:len(baseuri)-1]

	var wg sync.WaitGroup
	wg.Add(len(uripath))

	results := []Result{}

	for _, z := range uripath {
		go func(z string) {
			defer wg.Done()

			newpath := strings.ReplaceAll(uripath, string(z), strings.ToUpper(string(z)))

			var fullpath string
			if uri[len(uri)-1:] == "/" {
				fullpath = baseuri + newpath + "/"
			} else {
				fullpath = baseuri + "/" + newpath
			}

			statusCode, response, err := request("GET", fullpath, headers, proxy)
			if err != nil {
				log.Fatal(err)
			}

			results = append(results, Result{fullpath, statusCode, len(response)})
		}(string(z))
	}
	wg.Wait()
	printResponse(results)
}

func requester(uri string, userAgent string, proxy string) {
	if len(proxy) != 0 {
		if !strings.Contains(proxy, "http") {
			proxy = "http://" + proxy
		}
		color.Magenta("\n[*] USING PROXY: %s\n", proxy)
	}
	userProxy, _ := url.Parse(proxy)
	h := strings.Split(uri, "/")
	if len(h) < 4 {
		uri += "/"
	}
	if len(userAgent) == 0 {
		userAgent = "dontgo403/0.2"
	}

	headers := []header{
		{"User-Agent", userAgent},
	}
	requestMethods(uri, headers, userProxy)
	requestHeaders(uri, headers, userProxy)
	requestEndPaths(uri, headers, userProxy)
	requestMidPaths(uri, headers, userProxy)
	requestCapital(uri, headers, userProxy)
}
