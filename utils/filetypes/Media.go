package filetypes

import (
	"strings"
)

// TODO: more extensive file extension list probably need
var ImageTypes = map[string]bool{
	"jpg":  true,
	"webp": true,
	"png":  true,
	"gif":  true,
	"ico":  true,
	"bmp":  true,
	"dng":  true,
	"tiff": true,
	"tif":  true,
	"cr2":  true, // canon
	"crw":  true,
	"nrw":  true, // nikon
	"nef":  true,
	"arw":  true, // spny alpha
}

var VideoTypes = map[string]bool{
	"mp4":  true,
	"mpg":  true,
	"mmv":  true,
	"tod":  true,
	"wmv":  true,
	"asf":  true,
	"avi":  true,
	"divx": true,
	"mov":  true,
	"m4v":  true,
	"3gp":  true,
	"3g2":  true,
	"m2t":  true,
	"m2ts": true,
	"mts":  true,
	"mkv":  true,
}

func IsImage(filename string) bool {
	return isMediaType(filename, ImageTypes)
}

func IsVideo(filename string) bool {
	return isMediaType(filename, VideoTypes)
}

func IsMedia(filename string) bool {
	return IsVideo(filename) || IsImage(filename)
}

func isMediaType(filename string, mapChecker map[string]bool) bool {
	parts := strings.Split(filename, ".")
	return mapChecker[strings.ToLower(parts[len(parts)-1])]
}
