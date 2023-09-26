// Date: 26/06/2019 // 2023/09/11
// Created By ybenel
package media

import (
	"errors"
	"os"

	// "io/ioutil"
	"path"
	"sort"
	"strings"
	"sync"

	mylog "github.com/M1ndo/TokYo/pkg/log"
)

// Library manages importing and retrieving video data.
type Library struct {
	mu     sync.RWMutex
	Paths  map[string]*Path
	Videos map[string]*Video
}

// NewLibrary returns new instance of Library.
func NewLibrary() *Library {
	lib := &Library{
		Paths:  make(map[string]*Path),
		Videos: make(map[string]*Video),
	}
	return lib
}

// AddPath adds a media path to the library.
func (lib *Library) AddPath(p *Path) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()
	// make sure new path doesn't collide with existing ones
	for _, p2 := range lib.Paths {
		if p.Path == p2.Path {
			return errors.New("media: duplicate library path")
		}
		if p.Prefix == p2.Prefix {
			return errors.New("media: duplicate library prefix")
		}
	}
	lib.Paths[p.Path] = p
	return nil
}

// Import adds all valid videos from a given path.
func (lib *Library) Import(logger *mylog.Logger, p *Path) error {
	files, err := os.ReadDir(p.Path)
	// files, err := ioutil.ReadDir(p.Path)
	if err != nil {
		return err
	}
	for _, info := range files {
		// log.Println(info.Name())
		err = lib.Add(logger, path.Join(p.Path, info.Name()))
		if err != nil {
			// Ignore files that can't be parsed
			continue
		}
	}
	return nil
}

// Add adds a single video from a given file path.
func (lib *Library) Add(logger *mylog.Logger, filepath string) error {
	// log.Println("Filepath: %s", filepath)
	lib.mu.Lock()
	defer lib.mu.Unlock()
	d := path.Dir(filepath)
	p, ok := lib.Paths[d]
	if !ok {
		logger.Log.Warn("media: path %s not found", d)
		return errors.New("media: path not found")
	}
	// log.Println("P: %s", p) // Prints All files paths.
	n := path.Base(filepath)
	// log.Printf(n) // Prints all files in directory /videos
	v, err := ParseVideo(p, n)
	if err != nil {
		logger.Log.Warn(err)
		return err
	}
	// log.Println(v) // Prints array from tags returns metadata
	lib.Videos[v.ID] = v
	logger.Log.Info("Added:", v.Path)
	return nil
}

// Remove removes a single video from a given file path.
func (lib *Library) Remove(logger *mylog.Logger, filepath string) {
	lib.mu.Lock()
	defer lib.mu.Unlock()
	d := path.Dir(filepath)
	p, ok := lib.Paths[d]
	if !ok {
		logger.Log.Warn("media: path %s not found", d)
		return
	}
	n := path.Base(filepath)
	// ID is name without extension
	idx := strings.LastIndex(n, ".")
	if idx == -1 {
		idx = len(n)
	}
	id := n[:idx]
	if len(p.Prefix) > 0 {
		id = path.Join(p.Prefix, id)
	}
	v, ok := lib.Videos[id]
	if ok {
		delete(lib.Videos, id)
		logger.Log.Info("Removed:", v.Path)
	}
}

// Playlist returns a sorted Playlist of all videos.
func (lib *Library) Playlist() Playlist {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	pl := make(Playlist, len(lib.Videos))
	i := 0
	for _, v := range lib.Videos {
		pl[i] = v
		i++
	}
	sort.Sort(pl)
	return pl
}

// Handle Content Type Media
func (lib *Library) GetContentType(ext string) string {
	contentTypes := map[string]string{
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".webm": "video/webm",
		".weba": "video/webm",
		".flac": "audio/flac",
		".ogg": "audio/ogg",
		".m4a": "audio/m4a",
		".m4r": "audio/m4a",
		".opus": "audio/opus",
		".wav": "audio/wav",
	}
	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}
	return "application/octet-stream"
}
