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
	"github.com/trinhdrew1418/gphotos-cli/utils"
	"github.com/trinhdrew1418/gphotos-cli/utils/progressbar"
	"github.com/vbauerster/mpb"
	"strings"

	"github.com/spf13/cobra"
	"github.com/trinhdrew1418/gphotos-cli/utils/expobackoff"
	"github.com/trinhdrew1418/gphotos-cli/utils/filetypes"
	"github.com/trinhdrew1418/gphotos-cli/utils/retrievers"
	photoslib "google.golang.org/api/photoslibrary/v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

const (
	apiVer      = "v1"
	basePath    = "https://photoslibrary.googleapis.com/"
	MAX_WORKERS = 5
)

var (
	recursive      bool
	verbose        bool
	pbar           *mpb.Bar
	p              *mpb.Progress
	failedToUpload []string
)

type UploadInfo struct {
	token    string
	filename string
}

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload files",
	Long: "Uploads files and folders to your google photos library and/or specified" +
		"google photos album",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please give an argument")
			return
		}

		client, gphotoServ := getClientService(photoslib.PhotoslibraryScope)

		if workingAlbum != "" && selectAlbum {
			log.Fatal("Only select an album or input an album name, not both")
		}

		if workingAlbum != "" {
			workingAlbumID = retrievers.GetAlbumID(workingAlbum, gphotoServ)
		}

		if selectAlbum {
			workingAlbumID = retrievers.GetAlbum(gphotoServ, true)
			if workingAlbumID == "" {
				println("No writable albums")
			}
		}

		var filenames []string
		var directories []string

		for _, filepath := range args {
			if utils.IsFile(filepath) && filetypes.IsMedia(filepath) {
				filenames = append(filenames, filepath)
			} else {
				directories = append(directories, filepath)
			}
		}

		for _, dirPath := range directories {
			filenames = append(filenames, utils.GetFilepaths(dirPath, recursive, filetypes.IsMedia)...)
		}

		var amtBytes int64
		for _, filepath := range filenames {
			file, _ := os.Stat(filepath)
			amtBytes += file.Size()
		}

		var answer string
		if verbose {
			if workingAlbum != "" {
				fmt.Println("Uploading the following files to " + workingAlbum + ":")
			} else {
				fmt.Println("Uploading the following files:")
			}
			for _, filename := range filenames {
				fmt.Println("-", filename)
			}
		}

		if amtBytes < (1 << 10) {
			println(fmt.Sprintf("Uploading %v files, %.4g B", len(filenames), float64(amtBytes)))
		} else if amtBytes < (1 << (10 * 2)) {
			println(fmt.Sprintf("Uploading %v files, %.4g KB", len(filenames), float64(amtBytes)/float64(1<<(10*1))))
		} else if amtBytes < (1 << (10 * 3)) {
			println(fmt.Sprintf("Uploading %d files, %.4g MB", len(filenames), float64(amtBytes)/float64(1<<(10*2))))
		} else {
			println(fmt.Sprintf("Uploading %d files, %.4g GB", len(filenames), float64(amtBytes)/float64(1<<(10*3))))
		}

		print("Do you wish to proceed ([y]/n)?: ")
		_, err := fmt.Scan(&answer)
		if err != nil {
			log.Fatalf("Unable to read answer")
		}
		println()

		if strings.ToLower(answer) != "y" {
			return
		}

		p, pbar = progressbar.Make(int64(len(filenames)), "Uploading Files: ")
		pushFiles(gphotoServ, client, filenames)
		p.Wait()
		println()

		if len(failedToUpload) > 0 {
			println("The following files failed to upload: ")
			for _, fname := range failedToUpload {
				println(" - ", fname)
			}
		} else {
			println("No files failed to upload")
		}
	},
}

func pushFiles(srv *photoslib.Service, client *http.Client, filenames []string) {
	tokenQueue := make(chan UploadInfo)

	var wg sync.WaitGroup

	wg.Add(1)
	go requestUploadTokens(client, &filenames, &wg, tokenQueue)
	wg.Add(1)
	go createMedia(srv, &wg, tokenQueue)
	wg.Wait()
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

func createMedia(srv *photoslib.Service, wg *sync.WaitGroup, tokenQueue chan UploadInfo) {
	defer wg.Done()

	for s := range tokenQueue {
		newMediaItem := photoslib.NewMediaItem{SimpleMediaItem: &photoslib.SimpleMediaItem{UploadToken: s.token}}
		mediaItems := []*photoslib.NewMediaItem{&newMediaItem}

		makeItems := srv.MediaItems.BatchCreate(&photoslib.BatchCreateMediaItemsRequest{
			AlbumId:       workingAlbumID,
			NewMediaItems: mediaItems}).Do

		_, err := expobackoff.DoUntilSuccess(makeItems)

		if err != nil {
			log.Fatalf("Did not create", s.filename)
		}
	}
}

func uploader(uploadTasks chan string, tokenQueue chan UploadInfo, client *http.Client, w *sync.WaitGroup) {
	for filename := range uploadTasks {
		tok := getUploadToken(client, filename)
		if tok != "" {
			tokenQueue <- UploadInfo{tok, filename}
		}

		pbar.IncrBy(1)
	}
	w.Done()
}

// get upload token
func getUploadToken(client *http.Client, filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Cannot open a file %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/uploads", basePath, apiVer), file)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Add("X-Goog-Upload-File-Name", filename)
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	resp, err := expobackoff.RequestUntilSuccess(client.Do, req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode/100 != 2 {
		failedToUpload = append(failedToUpload, filename)
		return ""
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	uploadToken := string(b)

	return uploadToken
}

func init() {
	rootCmd.AddCommand(pushCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushCmd.PersistentFlags().String("foo", "", "A help for foo")
	pushCmd.PersistentFlags().BoolVarP(&recursive, "recursive", "r", false, "Recursively select files to be uploaded from directories")
	pushCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print out uploaded files")
	pushCmd.PersistentFlags().BoolVarP(&selectAlbum, "select", "s", false, "Select the album you want to do work in")
	pushCmd.PersistentFlags().StringVarP(&workingAlbum, "album", "a", "", "Input the album name that you want to upload to")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
