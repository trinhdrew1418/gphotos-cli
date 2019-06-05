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
	"strings"
	"time"
)

var (
	pastNumDays string
	startDate   string
	endDate     string
)

type Date struct {
	day   int
	month int
	year  int
}

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

		println("Select up to 10 categories from the following: ")
		println("ANIMALS LANDMARKS PETS UTILITY BIRTHDAYS LANDSCAPES RECEIPTS")
		println("WEDDINGS CITYSCAPES NIGHT SCREENSHOTS WHITEBOARDS DOCUMENTS")
		println("PEOPLE SELFIES FOOD PERFORMANCES SPORT")

		fmt.Scan()

	},
}

func GetDates() (Date, Date) {
	dateOptions := map[string]int{
		"Today":                       0,
		"Past week":                   1,
		"Past month":                  2,
		"Specific number of days ago": 3,
		"Specific date range":         4,
	}

	var startDate Date
	var endDate Date

	println()
	pmpt := promptui.Select{
		Label: "Select a time frame: ",
		Items: []string{"Today", "Past week", "Past month", "Specific number of days ago", "Specific date range"},
	}

	_, resp, err := pmpt.Run()

	switch dateOptions[resp] {
	case 0:
		startDate = getSomeDaysAgo(0)
		endDate = startDate
	case 1:
		endDate = getSomeDaysAgo(0)
		startDate = getSomeDaysAgo(7)
	case 2:
		endDate = getSomeDaysAgo(0)
		startDate = getSomeDaysAgo(30)
	case 3:
		var date string
		fmt.Scan(&date)
		strings.Split(date, "-")
		fmt.Scan(&date)
		strings.Split(date, "-")
		var numDays int
	case 4:
	}

	return startDate, endDate
}

func getSomeDaysAgo(num int) Date {
	Months := map[string]int{
		"January":   1,
		"February":  2,
		"March":     3,
		"April":     4,
		"May":       5,
		"June":      6,
		"July":      7,
		"August":    8,
		"September": 9,
		"October":   10,
		"November":  11,
		"December":  12,
	}

	var retDate Date
	year, month, day := time.Now().AddDate(0, 0, -num).Date()
	retDate.year = year
	retDate.month = Months[month.String()]
	retDate.month = Months[month.String()]
	retDate.day = day

	return retDate
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
