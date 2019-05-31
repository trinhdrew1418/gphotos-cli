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
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	photoslib "google.golang.org/api/photoslibrary/v1"
	"io/ioutil"
	"log"
	"os"
)

const (
	credentialFile = "/src/github.com/trinhdrew1418/gphotos-cli/auth/client_id.json"
	tokenFile      = "/src/github.com/trinhdrew1418/gphotos-cli/auth/token.json"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "store token",
	Long:  "TODO",

	// this function obtains the token and saves it
	Run: func(cmd *cobra.Command, args []string) {
		// get the config
		config := getConfig()

		// use the config to the get the token
		tok := getTokenFromWeb(config)

		// save the token into a file
		saveToken(tok)
	},
}

func saveToken(token *oauth2.Token) {
	fmt.Println("Saving token file")
	f, err := os.OpenFile(os.Getenv("GOPATH")+tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// make config
func getConfig() *oauth2.Config {
	b, err := ioutil.ReadFile(os.Getenv("GOPATH") + credentialFile)
	if err != nil {
		log.Fatalf("unable to read client credentials %v", err)
	}

	config, err := google.ConfigFromJSON(b, photoslib.PhotoslibraryScope)
	if err != nil {
		log.Fatalf("Unable to produce config %v", err)
	}

	return config
}

// get the token from the web
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser and paste the authorization"+
		"code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)

	if err != nil {
		log.Fatalf("Unable to retrieve token from the web %v", err)
	}

	return tok
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
	tok, err := tokenFromFile(os.Getenv("GOPATH") + tokenFile)
	if err != nil {
		log.Fatalf("token not cached: %v", err)
	}

	return tok
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
