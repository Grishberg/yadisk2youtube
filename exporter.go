// exporter.go tries to download videos from yandes.disk
//
// usage:
//   > go get github.com/Grishberg/yandex-disk-restapi-go
//   > go run exporter.go -token=access_token
//
//   You can find an access_token for your app at https://oauth.yandex.ru
package main

import (
	"flag"
	"fmt"
	"github.com/Grishberg/yandex-disk-restapi-go/src"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const DISK_PREFIX = "disk:"

type Uploader interface {
	Upload(fileName string)
}

type YaDiskDownloader struct {
	accessToken string
	pathMask    string
	client      *src.Client
	history     *SqlHistory
}

func main() {
	var accessToken string
	var pathMask string

	flag.StringVar(&accessToken, "token", "", "Access Token")
	flag.StringVar(&pathMask, "path", "", "Search path mask")

	if accessToken == "" {
		accessToken = os.Getenv("YADISK_ACCESS_TOKEN")
	}
	pathMask = "VIDEO"

	history := NewHistory()
	defer history.Close()

	mediaTypeImage := src.MediaType{}
	v := mediaTypeImage.Video()
	mediaTypes := []src.MediaType{*v}

	flag.Parse()

	if accessToken == "" {
		fmt.Println("\nPlease provide an access_token, one can be found at https://oauth.yandex.ru")
		flag.PrintDefaults()
		os.Exit(1)
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_ = os.Mkdir(path.Join(pwd, "tmp"), os.ModePerm)

	client := src.NewClient(accessToken)
	downloader := YaDiskDownloader{
		accessToken,
		pathMask,
		client,
		history}

	fmt.Printf("Fetching flat file list ...\n")
	var offset uint32 = 1
	for i := 0; i < 100; i++ {
		fmt.Println("read offset: ", offset)
		readed := downloader.getFlatFileListWithOffset(mediaTypes, offset)
		if readed == 0 {
			break
		}
		offset += readed
	}
}

func (yd YaDiskDownloader) getFlatFileListWithOffset(mediaTypes []src.MediaType,
	offset uint32) uint32 {
	options := src.FlatFileListRequestOptions{Media_type: mediaTypes, Offset: &offset}
	info, err := yd.client.NewFlatFileListRequest(options).Exec()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	downloadCount := 0
	for _, item := range info.Items {
		if yd.shouldDownloadFile(item) {
			yd.downloadItem(item)
			downloadCount += 1

			fmt.Println("downloaded, ", downloadCount)
			if downloadCount >= 5 {
				os.Exit(0)
			}
		}
	}
	return uint32(len(info.Items))
}

func (yd YaDiskDownloader) shouldDownloadFile(item src.ResourceInfoResponse) bool {
	if yd.history.IsExists(item.Path) {
		return false
	}
	if !strings.Contains(item.Path, yd.pathMask) {
		return false
	}
	var extension = strings.ToUpper(filepath.Ext(item.Name))
	if extension == ".MP4" {
		return true
	}

	return false
}

func (yd YaDiskDownloader) downloadItem(item src.ResourceInfoResponse) {
	path := item.Path
	response, err := yd.client.NewDownloadRequest(path).Exec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	downloader := DownloaderWithProgress{yd.accessToken, &ConsoleProgressOuptut{}}
	downloadedPath := downloader.DownloadFile(response.Href, item.Name, "tmp")
	fmt.Println(downloadedPath)
	Upload(downloadedPath, item.Name, "Uploaded with yadisk2youtube uploader", "rest")

	os.Remove(downloadedPath)
	yd.history.AddToHistory(item.Path)
}
