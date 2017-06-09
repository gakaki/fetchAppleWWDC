package fetchAppleWWDC2017

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
)

func batchFetchVideoDetails() []Video {
	//json 解析之后
	buf, err := ioutil.ReadFile("output.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "File Error: %s\n", err)
	}
	var videos []Video
	json.Unmarshal(buf, &videos)
	fmt.Println(videos, err)

	//go chan 并发
	maxWorkerCount := 20
	queue := make(chan Video, maxWorkerCount)
	runtime.GOMAXPROCS(runtime.NumCPU())

	wg := sync.WaitGroup{}

	var videosNew []Video
	for i := 0; i < maxWorkerCount; i++ {
		go func() {
			defer wg.Done()
			wg.Add(1)
			for v := range queue {
				v = fetchVideoDetail(v)
				fmt.Println(v.VideoSD)
				videosNew = append(videosNew, v)
			}
		}()
	}

	for _, v := range videos {
		queue <- v
	}
	close(queue)
	wg.Wait()

	//最后json 写入
	videosJson, _ := json.MarshalIndent(videosNew, "", " ")
	ioutil.WriteFile("output_detail.json", videosJson, 0644)
	return videosNew
}
func fetchVideoDetail(v Video) Video {

	url := v.DetailUrl
	doc, e := getContentFromUrl(url)

	if e != nil {
		fmt.Fprintf(os.Stderr, ">>>>>>>network Error: %s\n", e)
		return Video{}
	}

	v.Desc = doc.Find(".details p").Eq(0).Text()

	link_node := doc.Find(".links").Eq(0)

	var typeS = "link"

	doc.Find(".video a").Each(func(j int, node *goquery.Selection) {
		href := node.AttrOr("href", "")

		if strings.Contains(href, "_hd_") {
			v.VideoHD = href
		}
		if strings.Contains(href, "_sd_") {
			v.VideoSD = href
		}

	})

	link_node.Find("li.document,li.download").Each(func(j int, node *goquery.Selection) {
		documentA := node.Find("a").Eq(0)
		href := documentA.AttrOr("href", "")
		text := documentA.Text()

		if strings.Contains(href, "pdf") {
			typeS = "pdf"
		}
		if strings.Contains(href, "zip") {
			typeS = "code"
		}
		resource := Resource{}
		resource.Title = text
		resource.URL = href
		resource.Type = typeS

		v.Resources = append(v.Resources, resource)

	})

	return v
}
