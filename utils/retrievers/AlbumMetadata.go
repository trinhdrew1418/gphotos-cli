package retrievers

import (
	photoslibrary "google.golang.org/api/photoslibrary/v1"
	"log"
)

func GetAlbumsToID(s *photoslibrary.Service) *map[string]string {
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

func makeAlbumMap(s *photoslibrary.Service) *map[string] *map[string]*photoslibrary.Album {
	albumsResp, err := s.Albums.List().Do()
	if err != nil {
		log.Fatalln(err)
	}

	stringToAlbum := make(map[string]string)
	for _, album := range albumsResp.Albums {
			stringToAlbum[album.Title] = &album
	}

	return &stringToAlbum
}

}

func GetAlbumID(albumName string, s *photoslibrary.Service) string {
	stringToId := GetAlbumsMap(s)

}
