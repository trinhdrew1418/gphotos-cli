package retrievers

import (
	photoslibrary "google.golang.org/api/photoslibrary/v1"
	"log"
)

func GetAlbumsMap(s *photoslibrary.Service) *map[string]string {
	albumsResp, err := s.Albums.List().Do()
	if err != nil {
		log.Fatalln(err)
	}

	stringToId := make(map[string]string)
	for _, album := range albumsResp.Albums {
		if album.IsWriteable {
			stringToId[album.Title] = album.Id
		}
	}

	return &stringToId
}
