package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	help         bool
	isoutput     bool
	outputsuffix string = "_urls.txt"
)

//Fetch get & save
func Fetch(uri string) {
	//"OR REPLACE/IGNORE for replace"
	u, err := url.Parse(uri)
	if err != nil {
		log.Errorf("parse :%s error, details: %s ~", uri, err)
		return
	}
	time.Sleep(time.Millisecond * 300)
	t := time.Now()
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(uri)
	if err != nil {
		log.Errorf("fetch :%s error, details: %s ~", uri, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Errorf("fetch %v error, status code: %d ~", uri, resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("read body error ~ %s", err)
		return
	}
	if len(body) == 0 {
		log.Warnf("nil tile %v ~", uri)
		return
	}
	host := u.Hostname()
	fileName := host + u.Path
	dir := filepath.Dir(fileName)
	os.MkdirAll(dir, os.ModePerm)
	err = ioutil.WriteFile(fileName, body, os.ModePerm)
	if err != nil {
		log.Warnf("write file %v ~", err)
	}

	secs := time.Since(t).Seconds()
	log.Infof("fetch %s , %.3fs, %.2f kb, %s ...", uri, secs, float32(len(body))/1024.0, uri)
}

func main() {
	flag.BoolVar(&help, "h", false, "Show this help")
	flag.BoolVar(&isoutput, "o", false, fmt.Sprintf("output urls file, default name is inputfilename%s", outputsuffix))
	flag.Usage = usage
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Please specify the input file ~\n")
		flag.Usage()
		return
	}
	inputfile := args[0]
	infile, err := os.Open(inputfile)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	var outw *bufio.Writer
	if isoutput {
		name := strings.TrimSuffix(inputfile, filepath.Ext(inputfile))

		outfile, err := os.Create(name + outputsuffix)
		if err != nil {
			log.Fatalf("create output file error: %v", err)
		}
		defer outfile.Close()
		outw = bufio.NewWriter(outfile)
	}

	prefix := `fetch("`
	var cnt int
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			t := strings.TrimPrefix(line, prefix)
			idx := strings.Index(t, `"`)
			url := t[:idx]
			Fetch(url)
			// log.Println(url)
			if isoutput {
				fmt.Fprintln(outw, url)
			}
			cnt++
		}
	}
	if isoutput {
		outw.Flush()
	}
	log.Printf("pick up %d urls", cnt)
}

func usage() {
	fmt.Fprintf(os.Stderr, `fetch urls from input file: fetcher/v0.0.1
Usage: fetcher [-h] [-o urls.txt] input.txt
       input file format is chrome copy all as fetch

Options:
`)
	flag.PrintDefaults()
}
