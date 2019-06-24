package retrievers

import (
	photoslibrary "google.golang.org/api/photoslibrary/v1"
	"log"
)

func GetAlbumsToID(s *photoslibrary.Service, write bool) *map[string]string {
	albumMap := *makeAlbumMap(s)
	albumToID := make(map[string]string)
	for title := range albumMap {
		if write {
			if albumMap[title].IsWriteable {
				albumToID[title] = albumMap[title].Id
			}
		} else {
			albumToID[title] = albumMap[title].Id
		}
	}

	return &albumToID
}

func makeAlbumMap(s *photoslibrary.Service) *map[string]*photoslibrary.Album {
	var albums []*photoslibrary.Album
	albumsResp, err := s.Albums.List().Do()
	if err != nil {
		log.Fatalln(err)
	}
	albums = albumsResp.Albums
	for albumsResp.NextPageToken != "" {
		albumsResp, err = s.Albums.List().PageToken(albumsResp.NextPageToken).Do()
		if err != nil {
			log.Fatalln(err)
		}
		albums = append(albums, albumsResp.Albums...)
	}

	stringToAlbum := make(map[string]*photoslibrary.Album)
	for _, album := range albums {
		stringToAlbum[album.Title] = album
	}

	return &stringToAlbum
}

func GetAlbumID(albumName string, s *photoslibrary.Service) string {
	albumMap := *makeAlbumMap(s)

	if val, ok := albumMap[albumName]; ok {
		if !val.IsWriteable {
			log.Fatalln("Album exists but is not writable")
		} else {
			println("Album found!")
		}
		return val.Id
	} else {
		log.Fatalln("Album does not exist")
		return ""
	}
}
