package main

/**
 * @website http://albulescu.ro
 * @author Cosmin Albulescu <cosmin@albulescu.ro>
 */

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

/**
* Is called when progress changes
 */
type ProgressOutput interface {
	UpdateProgress(progress float64)
}

type DownloaderWithProgress struct {
	accessToken    string
	progressOutput ProgressOutput
}

type ConsoleProgressOuptut struct{}

func (o *ConsoleProgressOuptut) UpdateProgress(progress float64) {
	fmt.Printf("%.0f", progress)
	fmt.Println("%")
}

func (d *DownloaderWithProgress) printDownloadPercent(done chan int64,
	path string,
	total int64) {

	var stop bool = false

	for {
		select {
		case <-done:
			stop = true
		default:

			file, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}

			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()

			if size == 0 {
				size = 1
			}

			var percent float64 = float64(size) / float64(total) * 100
			d.progressOutput.UpdateProgress(percent)
		}

		if stop {
			break
		}

		time.Sleep(time.Second)
	}
}

func (d *DownloaderWithProgress) DownloadFile(
	url string, name string, dest string) string {

	file := path.Base(name)

	log.Printf("Downloading file %s n", file, url)

	var path bytes.Buffer
	path.WriteString(dest)
	path.WriteString("/")
	path.WriteString(file)

	start := time.Now()

	out, err := os.Create(path.String())

	if err != nil {
		fmt.Println(path.String())
		panic(err)
	}

	defer out.Close()

	client := &http.Client{CheckRedirect: d.redirectPolicyFunc}
	size := d.calculateFileSize(client, url)

	done := make(chan int64)

	go d.printDownloadPercent(done, path.String(), int64(size))

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "OAuth "+d.accessToken)
	resp, err := client.Do(req)

	// resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)

	if err != nil {
		panic(err)
	}

	done <- n

	elapsed := time.Since(start)
	log.Printf("Download completed in %s", elapsed)
	return path.String()
}

func (d *DownloaderWithProgress) calculateFileSize(client *http.Client, url string) int {
	headResp, err := http.Head(url)

	if err != nil {
		panic(err)
	}

	defer headResp.Body.Close()

	size, err := strconv.Atoi(headResp.Header.Get("Content-Length"))

	if err != nil {
		panic(err)
	}
	return size
}

func (d *DownloaderWithProgress) redirectPolicyFunc(r *http.Request,
	via []*http.Request) error {
	r.Header.Add("Authorization", "OAuth "+d.accessToken)
	return nil
}
