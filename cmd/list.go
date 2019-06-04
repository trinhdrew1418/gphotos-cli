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
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"google.golang.org/api/photoslibrary/v1"
	"strconv"
	"time"
)

var (
	pastNumDays string
	startDate   string
	endDate     string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, gphotoService := getClientService(photoslibrary.PhotoslibraryScope)

		dateOptions := map[string]int{
			"Today":                       0,
			"Past week":                   1,
			"Past month":                  2,
			"Specific number of days ago": 3,
			"Specific date range":         4,
		}

		println()
		pmpt := promptui.Select{
			Label: "Select a time frame: ",
			Items: []string{"Today", "Past week", "Past month", "Specific number of days ago", "Specific date range"},
		}

		_, resp, err := pmpt.Run()

		switch dateOptions[resp] {
		case 0:
			year, month, day := time.Now().Date()
		case 1:
			year, month, day := time.Now().AddDate(0, 0, -7).Date()
		case 2:
			year, month, day := time.Now().AddDate(0, -1, 0).Date()
		case 3:
			var numDays int
		case 4:
		}

		println("Select up to 10 categories from the following: ")
		println("ANIMALS LANDMARKS PETS UTILITY BIRTHDAYS LANDSCAPES RECEIPTS")
		println("WEDDINGS CITYSCAPES NIGHT SCREENSHOTS WHITEBOARDS DOCUMENTS")
		println("PEOPLE SELFIES FOOD PERFORMANCES SPORT")

		fmt.Scan()

	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
