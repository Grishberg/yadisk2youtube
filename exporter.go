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
)

type Uploader interface {
	Upload(fileName string)
}

type YaDiskDownloader struct {
}

func main() {
	var accessToken string
	flag.StringVar(&accessToken, "token", "", "Access Token")

	if accessToken == "" {
		accessToken = os.Getenv("YADISK_ACCESS_TOKEN")
	}

	mediaTypeImage := src.MediaType{}
	v := mediaTypeImage.Video()
	mediaTypes := []src.MediaType{*v}

	flag.Parse()

	if accessToken == "" {
		fmt.Println("\nPlease provide an access_token, one can be found at https://oauth.yandex.ru")
		flag.PrintDefaults()
		os.Exit(1)
	}

	client := src.NewClient(accessToken)

	fmt.Printf("Fetching flat file list ...\n")
	var offset uint32 = 1
	for i := 0; i < 30; i++ {
		fmt.Println("read offset: ", offset)
		readed := getFlatFileListWithOffset(accessToken, client, mediaTypes, offset)
		if readed == 0 {
			break
		}
		offset += readed
	}
}

func getFlatFileListWithOffset(accessToken string,
	client *src.Client,
	mediaTypes []src.MediaType,
	offset uint32) uint32 {
	options := src.FlatFileListRequestOptions{Media_type: mediaTypes, Offset: &offset}
	info, err := client.NewFlatFileListRequest(options).Exec()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i, v := range info.Items {
		downloadItemIfNeeded(accessToken, client, v)
		fmt.Println(i, v.Name, v.Path)
	}
	if info.Limit != nil {
		fmt.Printf("\tLimit: %d\n", info.Limit)
	}
	if info.Offset != nil {
		fmt.Printf("\tOffset: %d\n", *info.Offset)
	}
	return uint32(len(info.Items))
}

func downloadItemIfNeeded(accessToken string,
	client *src.Client,
	item src.ResourceInfoResponse) {
	path := item.Path
	response, err := client.NewDownloadRequest(path).Exec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	DownloadFile(accessToken, response.Href, "tmp")
	fmt.Println(response.Href)
	os.Exit(0)

}
