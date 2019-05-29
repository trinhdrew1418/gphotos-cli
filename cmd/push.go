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
	"encoding/json"
	"os"
	"log"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	photoslib "google.golang.org/api/photoslibrary/v1"
)

const apiVer = "v1"
const basePath = "https://photoslibrary.googleapis.com/"

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload files",
	Long: "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		config := getConfig()

		tokFile := "token.json"
		tok, err := tokenFromFile(tokFile)
		if err != nil {
			log.Fatalf("token not cached: %v", err)
		}

		client := config.Client(context.Background(), tok)
		gphotoServ, err := photoslib.New(client) 
		if err != nil {
			log.Fatalf("Unable to retrieve google photos client: %v", err)
		}

		push(gphotoServ, client)
	},
}

func push(srv *photoslib.Service, client *http.Client) {
	filename := "test.png"

	token, err := getUploadToken(client, filename)
	if err != nil {
		log.Fatalf("something is fucked %v", err)
	}

	resp, err := srv.MediaItems.BatchCreate(&photoslib.BatchCreateMediaItemsRequest{
		AlbumId: "",
		NewMediaItems: []*photoslib.NewMediaItem{
			&photoslib.NewMediaItem{
				Description: filename,
				SimpleMediaItem: &photoslib.SimpleMediaItem{UploadToken: token},
			},
		},
	}).Do()

	if err == nil {
		fmt.Println(resp.NewMediaItemResults[0].Status.Message)
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

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("fucked up making http request %v", err)
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	
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
