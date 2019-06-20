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

	"github.com/spf13/cobra"
)

// albumsCmd represents the albums command
var albumsCmd = &cobra.Command{
	Use:   "albums",
	Short: "Manage albums, create new ones or add existing photos to existing albums",
	Long: `Manage your albums. Create new ones or add photos to existing
		ones. Create new ones by calling 

			gphotos-cli albums create

			`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("albums called")
	},
}

func init() {
	rootCmd.AddCommand(albumsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// albumsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// albumsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
