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
	"github.com/trinhdrew1418/gphotos-cli/utils/filetypes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	photoslib "google.golang.org/api/photoslibrary/v1"
)

const apiVer = "v1"
const basePath = "https://photoslibrary.googleapis.com/"

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
		client := config.Client(context.Background(), tok)
		gphotoServ, err := photoslib.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve google photos client: %v", err)
		}

		var filenames []string
		switch {

		//TODO: distinguish from folder and otherwise
		case len(args) > 1:
			for _, file := range args {
				if filetypes.IsImage(file) {
					filenames = append(filenames, file)
				}
			}
			fmt.Println("Uploading the following files: ")
			for _, filename := range filenames {
				fmt.Println(filename)
			}
		default:
			filenames = append(filenames, args[0])
			//TODO: verify the files + the proper extension
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
			Description:     tokens[i].filename,
			SimpleMediaItem: &photoslib.SimpleMediaItem{UploadToken: tokens[i].token},
		}
		mediaItems[i] = &newMediaItem
	}

	resp, err := srv.MediaItems.BatchCreate(&photoslib.BatchCreateMediaItemsRequest{
		NewMediaItems: mediaItems,
	}).Do()

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
	go processTokens(client, &filenames, &wg, tokenQueue)
	wg.Add(1)
	go collectTokens(&wg, tokenQueue, &tokens)
	wg.Wait()

	return tokens
}

func processTokens(client *http.Client, filenames *[]string, wg *sync.WaitGroup, tokenQueue chan UploadInfo) {
	defer wg.Done()
	defer close(tokenQueue)

	var w sync.WaitGroup
	for _, filename := range *filenames {
		w.Add(1)
		go channelTok(client, filename, &w, tokenQueue)
	}
	w.Wait()
}
func channelTok(client *http.Client, filename string, w *sync.WaitGroup, tokenQueue chan UploadInfo) {
	tok, err := getToken(client, filename)
	if err != nil {
		log.Fatalf("unable to make POST request")
	}
	w.Done()
	tokenQueue <- UploadInfo{tok, filename}
}
func collectTokens(wg *sync.WaitGroup, tokenQueue chan UploadInfo, tokens *[]UploadInfo) {
	defer wg.Done()
	for s := range tokenQueue {
		println(s.filename)
		*tokens = append(*tokens, s)
	}
}

// get upload token
func getToken(client *http.Client, filename string) (token string, err error) {
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

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("fucked up making http request %v", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("upload tok error %v", err)
	}

	uploadToken := string(b)

	return uploadToken, err
}

// obtains cached token
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

func loadToken() *oauth2.Token {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Fatalf("token not cached: %v", err)
	}

	return tok
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
