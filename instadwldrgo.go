package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	var path string
	var url string

	//Teste if parameters are correctly informed
	if len(os.Args) < 2 {
		fmt.Println("Argument missing, intagram picture path is mandatory!")
		os.Exit(1)
	}
	url = os.Args[1]
	if len(os.Args) == 3 {
		path = os.Args[2]
		//TODO: allert if pathe is not in the correct format .\aaaaa\
	}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}
	var isVideo bool
	var mediaPath string
	isVideo = false
	//1 - check if
	//video, <meta name="medium" content="video" /><meta property="og:type" content="video" />
	//picture <meta name="medium" content="image" /><meta property="og:type" content="instapp:photo" />
	//if video use <meta property="og:video" content="https://scontent-gig2-1.cdninstagram.com/vp/2368a99a95fefe4bedec4163eb1d90ad/5CA83591/t50.2886-16/56210329_604357073401281_8217347639360880640_n.mp4?_nc_ht=scontent-gig2-1.cdninstagram.com" />
	//if picture display_url

	//First see if it is a video
	doc.Find("meta").
		Each(func(i int, item *goquery.Selection) {
			if item.AttrOr("property", "") == "og:video" {
				isVideo = true
				mediaPath = item.AttrOr("content", "")
				//fmt.Printf("%s\n", mediaPath)
				getPicture(mediaPath, path)
			}
		})
	//Second if not video is a picture and recursively see if it is multipicture
	if isVideo == false {
		doc.Find("script").
			Each(func(i int, item *goquery.Selection) {
				re := regexp.MustCompile(`"display_url":".*?"`)
				linkMatches := re.FindAllStringSubmatch(item.Text(), -1)
				var mediaPathArray []string
				if linkMatches != nil {
					//mediaPath = strings.TrimRight(strings.TrimLeft(linkMatches[0][0], "\"display_url\":\""), "\"")
					for i := range linkMatches {
						//fmt.Println("---- linkMatches -----")
						//fmt.Println(linkMatches[i][0])
						mediaPath = strings.TrimRight(strings.TrimLeft(linkMatches[i][0], "\"display_url\":\""), "\"")
						if len(mediaPathArray) == 0 {
							mediaPathArray = append(mediaPathArray, mediaPath)
						} else {
							var isPathPresent = false
							for j := range mediaPathArray {
								if mediaPath == mediaPathArray[j] {
									isPathPresent = true
								}
							}
							if isPathPresent == false {
								mediaPathArray = append(mediaPathArray, mediaPath)
							}
							isPathPresent = false
						}
					}
					// Now download each picture
					for k := range mediaPathArray {
						mediaPath = mediaPathArray[k]
						getPicture(mediaPath, path)
					}
				}

			})
	}
	fmt.Println("0")
}

func getPicture(url string, path string) {
	// Just a simple GET request to the image URL
	// We get back a *Response, and an error
	res, err := http.Get(url)

	if err != nil {
		log.Fatalf("http.Get -> %v", err)
	}

	// We read all the bytes of the image
	// Types: data []byte
	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatalf("ioutil.ReadAll -> %v", err)
	}

	// You have to manually close the body, check docs
	// This is required if you want to use things like
	// Keep-Alive and other HTTP sorcery.
	res.Body.Close()

	//   \.+(jpg|gif|png|avi|mp4)
	// You can now save it to disk or whatever...
	r, _ := regexp.Compile(`[\w-]+\.+(jpg|gif|png|avi|mp4)`)
	fileName := path + r.FindString(url)

	ioutil.WriteFile(fileName, data, 0666)

	log.Println(fileName)
}
