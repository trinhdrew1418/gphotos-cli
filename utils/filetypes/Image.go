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
	"cr2":  true, // canon
	"crw":  true,
	"nrw":  true, // nikon
	"nef":  true,
	"arw":  true, // spny alpha
}

func IsImage(filename string) bool {
	parts := strings.Split(filename, ".")
	return ImageTypes[parts[len(parts)-1]]
}
