// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/api/photoslibrary/v1"
	"log"
	"os"
	"strings"
)

var (
	startDate  Date
	endDate    Date
	categories []string
)

const (
	MediaItemsIDLoc = "/src/github.com/trinhdrew1418/gphotos-cli/cache/IDs.json"
)

type Date struct {
	day   int
	month int
	year  int
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Searches your google photos library for qualifying search entries",
	Long: `Find relevant google photos entries in your library. This command will prompt
you for photos qualifying photos throughout a specified time range, content type, file type, camera type.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, gphotoService := getClientService(photoslibrary.PhotoslibraryScope)
		var noDate bool
		var noCat bool
		var answer string

		startDate, endDate = GetDates(&noDate)
		println()
		categories = GetCategories(&noCat)
		println()

		filt := photoslibrary.Filters{}
		if !noCat {
			filt.ContentFilter = &photoslibrary.ContentFilter{IncludedContentCategories: categories}
		}

		if !noDate {
			filt.DateFilter = &photoslibrary.DateFilter{
				Ranges: []*photoslibrary.DateRange{
					{
						StartDate: &photoslibrary.Date{
							Day:   int64(startDate.day),
							Month: int64(startDate.month),
							Year:  int64(startDate.year),
						},
						EndDate: &photoslibrary.Date{
							Day:   int64(endDate.day),
							Month: int64(endDate.month),
							Year:  int64(endDate.year),
						},
					},
				},
			}
		}

		resp, err := gphotoService.MediaItems.Search(&photoslibrary.SearchMediaItemsRequest{Filters: &filt}).Do()
		if err != nil {
			log.Fatal("Failed media search")
		}

		for _, mItem := range resp.MediaItems {
			println(mItem.BaseUrl)
			println()
		}

		print("Do you want to cache your search request to download at a later time? ([y]/n): ")
		fmt.Scan(&answer)

		if strings.ToLower(answer) == "y" {
			f, err := os.OpenFile(os.Getenv("GOPATH")+MediaItemsIDLoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				log.Fatalf("Unable to cache oauth token: %v", err)
			}

			var IDs []string
			for _, mItem := range resp.MediaItems {
				IDs = append(IDs, mItem.Id)
			}

			defer f.Close()
			err = json.NewEncoder(f).Encode(IDs)
			if err != nil {
				log.Fatal(err)
			}
			println("Successfully cached search query")
		}

	},
}

func FilterTypes(items *[]*photoslibrary.MediaItem, kind string, sorter func(chan *photoslibrary.MediaItem, map[string][]*photoslibrary.MediaItem)) {
	var answer string

	println("Indexing files...")
	sortedFiles := *sortFiles(items, sortByFileType)

	println("Here are the available" + kind + "types: ")
	for key := range sortedFiles {
		println(" - ", key)
	}
	print("Input which types you'd like to include: ")
	_, err := fmt.Scan(&answer)
	if err != nil {
		log.Fatal("Unable to read response")
	}
	var merged []*photoslibrary.MediaItem
	keys := strings.Split(answer, " ")
	for _, key := range keys {
		merged = append(merged, sortedFiles[key]...)
	}

	items = &merged
}

func sortFiles(items *[]*photoslibrary.MediaItem,
	sorter func(chan *photoslibrary.MediaItem, map[string][]*photoslibrary.MediaItem)) *map[string][]*photoslibrary.MediaItem {

	var sorted map[string][]*photoslibrary.MediaItem
	feed := make(chan *photoslibrary.MediaItem)

	for i := 0; i < MAX_WORKERS; i++ {
		go sorter(feed, sorted)
	}

	for _, mItem := range *items {
		feed <- mItem
	}

	return &sorted
}

func sortByFileType(feed chan *photoslibrary.MediaItem, sorted map[string][]*photoslibrary.MediaItem) {
	defer close(feed)

	for mItem := range feed {
		val := sorted[strings.Split(mItem.MimeType, "/")[1]]
		val = append(val, mItem)
		sorted[strings.Split(mItem.MimeType, "/")[1]] = val
	}
}

func sortByCameraType(feed chan *photoslibrary.MediaItem, sorted map[string][]*photoslibrary.MediaItem) {
	defer close(feed)

	for mItem := range feed {
		val := sorted[mItem.MediaMetadata.Photo.CameraMake+" "+mItem.MediaMetadata.Photo.CameraModel]
		val = append(val, mItem)
		sorted[mItem.MediaMetadata.Photo.CameraMake+" "+mItem.MediaMetadata.Photo.CameraModel] = val
	}
}
