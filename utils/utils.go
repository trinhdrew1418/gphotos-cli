package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func IsFile(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		log.Fatalln(path, "is not valid file or directory")
	}

	return file.Mode().IsRegular()
}

func IsDir(path string) bool {
	return !IsFile(path)
}

func GetFilepaths(directoryPath string, recurse bool, filter func(string) bool) []string {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Fatalln(err)
	}
	var filepaths []string
	var directorypaths []string

	for _, file := range files {
		if file.Mode().IsRegular() && filter(file.Name()) {
			filepaths = append(filepaths, filepath.Join(directoryPath, file.Name()))
		} else if file.Mode().IsDir() {
			directorypaths = append(directorypaths, filepath.Join(directoryPath, file.Name()))
		}
	}

	if recurse {
		for _, dirPath := range directorypaths {
			filepaths = append(filepaths, GetFilepaths(dirPath, recurse, filter)...)
		}
	}

	return filepaths
}
