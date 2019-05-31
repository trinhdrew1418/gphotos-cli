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
	"fmt"
	"github.com/trinhdrew1418/gphotos-cli/utils/expobackoff"
	"github.com/trinhdrew1418/gphotos-cli/utils/filetypes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	photoslib "google.golang.org/api/photoslibrary/v1"
)

const apiVer = "v1"
const basePath = "https://photoslibrary.googleapis.com/"
const MAX_WORKERS = 5

type UploadInfo struct {
	token    string
	filename string
}

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload files",
	Long:  "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO display some help shit (same message as Long preferably)
		if len(args) < 1 {
			fmt.Println("Please give an argument")
			return
		}

		config := getConfig()
		tok := loadToken()
		newTok, err := config.TokenSource(context.TODO(), tok).Token()
		if err != nil {
			log.Fatalln(err)
		}

		if newTok.AccessToken != tok.AccessToken {
			fmt.Println("Token refreshed")
			saveToken(newTok)
			tok = newTok
		}

		client := config.Client(context.Background(), newTok)
		gphotoServ, err := photoslib.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve google photos client: %v", err)
		}

		var filenames []string

		for _, file := range args {
			if filetypes.IsImage(file) {
				filenames = append(filenames, file)
			}
		}

		fmt.Println("Uploading the following files: ")
		for _, filename := range filenames {
			fmt.Println(filename)
		}

		pushFiles(gphotoServ, client, filenames)
	},
}

func pushFiles(srv *photoslib.Service, client *http.Client, filenames []string, albumID ...string) {
	tokens := getUploadTokens(client, filenames)
	mediaItems := make([]*photoslib.NewMediaItem, len(filenames))

	for i := 0; i < len(filenames); i++ {
		println(tokens[i].filename)
		newMediaItem := photoslib.NewMediaItem{
			SimpleMediaItem: &photoslib.SimpleMediaItem{UploadToken: tokens[i].token},
		}
		mediaItems[i] = &newMediaItem
	}

	resp, err := srv.MediaItems.BatchCreate(&photoslib.BatchCreateMediaItemsRequest{
		NewMediaItems: mediaItems,
	}).Do()

	// TODO: check for other possible HTTP response errors
	if resp.HTTPStatusCode == 429 {
		durations := expobackoff.Calculate(expobackoff.NUM_RETRIES)
		for _, sleepDur := range durations {
			time.Sleep(sleepDur)
			resp, err = srv.MediaItems.BatchCreate(&photoslib.BatchCreateMediaItemsRequest{
				NewMediaItems: mediaItems}).Do()

			if resp.HTTPStatusCode == 200 {
				break
			}
		}
	}

	if err == nil {
		for _, result := range resp.NewMediaItemResults {
			fmt.Println(result.Status.Message)
		}
	} else {
		log.Fatalf("Did not create %v", err)
	}
}

func getUploadTokens(client *http.Client, filenames []string) []UploadInfo {
	tokens := make([]UploadInfo, 0)
	tokenQueue := make(chan UploadInfo)

	var wg sync.WaitGroup

	wg.Add(1)
	go requestUploadTokens(client, &filenames, &wg, tokenQueue)
	wg.Add(1)
	go collectTokens(&wg, tokenQueue, &tokens)
	wg.Wait()

	return tokens
}

func uploader(uploadTasks chan string, tokenQueue chan UploadInfo, client *http.Client, w *sync.WaitGroup) {
	for filename := range uploadTasks {
		tok, err := getUploadToken(client, filename)
		if err != nil {
			log.Fatalf("unable to make POST request")
		}
		if tok != "" {
			tokenQueue <- UploadInfo{tok, filename}
		}
	}
	w.Done()
}

func requestUploadTokens(client *http.Client, filenames *[]string, wg *sync.WaitGroup, tokenQueue chan UploadInfo) {
	defer wg.Done()
	defer close(tokenQueue)

	var w sync.WaitGroup
	uploadTasks := make(chan string)

	for i := 0; i < MAX_WORKERS; i++ {
		w.Add(1)
		go uploader(uploadTasks, tokenQueue, client, &w)
	}

	for _, filename := range *filenames {
		uploadTasks <- filename
	}
	close(uploadTasks)

	w.Wait()
}

func collectTokens(wg *sync.WaitGroup, tokenQueue chan UploadInfo, tokens *[]UploadInfo) {
	defer wg.Done()
	for s := range tokenQueue {
		println(s.filename)
		*tokens = append(*tokens, s)
	}
}

// get upload token
func getUploadToken(client *http.Client, filename string) (token string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Cannot open a file %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/uploads", basePath, apiVer), file)
	if err != nil {
		log.Fatalf("fucked up making the http format %v", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Add("X-Goog-Upload-File-Name", filename)
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	resp, err := expobackoff.RequestUntilSuccess(client.Do, req)

	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != 200 {
		println(filename + " failed to upload")
		return "", nil
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	uploadToken := string(b)

	return uploadToken, err
}

func init() {
	rootCmd.AddCommand(pushCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
