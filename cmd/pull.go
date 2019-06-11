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
	"google.golang.org/api/photoslibrary/v1"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		exists := existsCache()
		var mediaItemIDs []string
		var answer string
		_, gphotoService := getClientService(photoslibrary.PhotoslibraryScope

		if exists {
			print("Detected cached search query, do you want to download files from this? ([y]/n)")
			_, err :=fmt.Scan(&answer)
			if err != nil {
				log.Fatal(err)
			}

			if strings.ToLower(answer) == "y" {
				f, err := os.Open(os.Getenv("GOPATH") + MediaItemsIDLoc)
				if err != nil {
					log.Fatal("Unable to open file for some reason")
				}

				defer f.Close()
				err = json.NewDecoder(f).Decode(mediaItemIDs)
				if err != nil {
					log.Fatal("Unable to decode json")
				}
			} else {
				// delete the file
			}
		}

		if len(mediaItemIDs) > 0 {
			for _, id := range mediaItemIDs {
				resp, err := gphotoService.MediaItems.Get(id).Do()
				if resp != nil {
					downloadLink := resp.BaseUrl + "-d"
				}
			}

			return
		}

		//go through prompts


	},
}

func existsCache() bool {
	// does
}

func init() {
	rootCmd.AddCommand(pullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
