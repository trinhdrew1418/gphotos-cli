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

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/trinhdrew1418/gphotos-cli/utils/expobackoff"
	"github.com/trinhdrew1418/gphotos-cli/utils/filetypes"
	"github.com/trinhdrew1418/gphotos-cli/utils/retrievers"
	"golang.org/x/net/context"
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
	MAX_WORKERS = 4
)

var (
	recursive      bool
	verbose        bool
	pbar           mpb.Bar
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

		config := getConfig(photoslib.PhotoslibraryScope)
		tok := loadToken()
		newTok, err := config.TokenSource(context.TODO(), tok).Token()

		if err != nil {
			log.Fatal(err)
		}

		if newTok.AccessToken != tok.AccessToken {
			fmt.Println("Token refreshed")
			saveToken(newTok)
			tok = newTok
		}

		client := config.Client(context.Background(), newTok)
		gphotoServ, err := photoslib.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve google photos service: %v", err)
		}

		var album string

		if selectAlbum && workingAlbum != "" {
			println("Can only either use the album select or declare an album destination, not both")
			return
		}

		if workingAlbum != "" {
			workingAlbumID = retrievers.GetAlbumID(workingAlbum, gphotoServ)
		}

		if selectAlbum {
			albumToID := *retrievers.GetAlbumsToID(gphotoServ)

			if len(albumToID) >= 1 {
				titles := make([]string, len(albumToID))
				i := 0
				for k := range albumToID {
					titles[i] = k
					i++
				}
				prompt := promptui.Select{
					Label: "Select album",
					Items: titles,
				}

				_, album, err = prompt.Run()

				if err != nil {
					log.Fatalln(err)
				}
				workingAlbumID = albumToID[album]
			} else {
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
				fmt.Println("Uploading the following files to " + album + ":")
			} else {
				fmt.Println("Uploading the following files:")
			}
			for _, filename := range filenames {
				fmt.Println("-", filename)
			}
		}

		if amtBytes < (1 << 10) {
			println(fmt.Sprintf("Uploading %v files, %.4g B", len(filenames), float64(amtBytes)))
		} else if amtBytes < (1 << 10 * 2) {
			println(fmt.Sprintf("Uploading %v files, %.4g KB", len(filenames), float64(amtBytes)/float64(1<<(10*1))))
		} else if amtBytes < (1 << 10 * 3) {
			println(fmt.Sprintf("Uploading %d files, %.4g MB", len(filenames), float64(amtBytes)/float64(1<<(10*2))))
		} else {
			println(fmt.Sprintf("Uploading %d files, %.4g GB", len(filenames), float64(amtBytes)/float64(1<<(10*3))))
		}

		print("Do you wish to proceed ([y]/n)?: ")
		_, err = fmt.Scan(&answer)
		if err != nil {
			log.Fatalf("Unable to read answer")
		}
		println()

		if strings.ToLower(answer) != "y" {
			return
		}

		pbar = *progressbar.Make(int64(len(filenames)))
		pushFiles(gphotoServ, client, filenames)

		if len(failedToUpload) > 0 {
			println("The following files failed to upload: ")
			for _, fname := range failedToUpload {
				println(" - ", fname)
			}
		}
	},
}

func pushFiles(srv *photoslib.Service, client *http.Client, filenames []string) {
	tokens := make([]UploadInfo, 0)
	tokenQueue := make(chan UploadInfo)

	var wg sync.WaitGroup

	wg.Add(1)
	go requestUploadTokens(client, &filenames, &wg, tokenQueue)
	wg.Add(1)
	go createMedia(srv, &wg, tokenQueue, &tokens)
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

func createMedia(srv *photoslib.Service, wg *sync.WaitGroup, tokenQueue chan UploadInfo, tokens *[]UploadInfo) {
	defer wg.Done()

	for s := range tokenQueue {
		newMediaItem := photoslib.NewMediaItem{SimpleMediaItem: &photoslib.SimpleMediaItem{UploadToken: s.token}}
		mediaItems := []*photoslib.NewMediaItem{&newMediaItem}

		makeItems := srv.MediaItems.BatchCreate(&photoslib.BatchCreateMediaItemsRequest{
			AlbumId:       workingAlbumID,
			NewMediaItems: mediaItems}).Do

		resp, err := expobackoff.DoUntilSuccess(makeItems)

		if err == nil {
			result := resp.NewMediaItemResults[0]
			pbar.IncrBy(1)
			if verbose {
				fmt.Println(result.Status.Message, "uploaded ", s.filename)
			}
		} else {
			log.Fatalf("Did not create", s.filename)
		}
	}
}

func uploader(uploadTasks chan string, tokenQueue chan UploadInfo, client *http.Client, w *sync.WaitGroup) {
	for filename := range uploadTasks {
		tok, err := getUploadToken(client, filename)
		if err != nil {
			log.Fatalf("unable to make POST request", err)
		}
		if tok != "" {
			tokenQueue <- UploadInfo{tok, filename}
		}
	}
	w.Done()
}

// get upload token
func getUploadToken(client *http.Client, filename string) (token string, err error) {
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

	if resp.StatusCode != 200 {
		failedToUpload = append(failedToUpload, filename)
		return "", nil
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
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
	pushCmd.PersistentFlags().BoolVarP(&recursive, "recursive", "r", false, "Recursively select files to be uploaded from directories")
	pushCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print out uploaded files")
	pushCmd.PersistentFlags().BoolVarP(&selectAlbum, "select", "s", false, "Select the album you want to do work in")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
