// Date: 26/06/2019
// Created By ybenel
package media

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

// Video represents metadata for a single video.
type Video struct {
	ID          string
	Title       string
	Album       string
	Description string
	Thumb       []byte
	ThumbType   string
	Modified    string
	Size        int64
	FileType    template.HTML
	Path        string
	FilePath    string
	Timestamp   time.Time
	Restricted  bool
}

// valid supported extensions
var extensions = []string{".webm", ".wav", ".mp4", ".mp3", ".opus", ".ogg", ".flac", ".m4a", ".m4r", ".acc", ".wav", ".weba"}

// ParseVideo parses a video file's metadata and returns a Video.
func ParseVideo(p *Path, name string) (*Video, error) {
	pth := path.Join(p.Path, name)
	f, err := os.Open(pth)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("Ignoring %s is a directory ", name)
	}
	size := info.Size()
	timestamp := info.ModTime()
	modified := timestamp.Format("2006-01-02 03:04 PM")
	// ID is name without extension
	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		idx = len(name)
	}
	id := name[:idx]
	if len(p.Prefix) > 0 {
		// if there's a prefix prepend it to the ID
		id = path.Join(p.Prefix, name[:idx])
	}
	v := &Video {
		ID:          id,
		Modified:    modified,
		Size:        size,
		Path:        pth,
		Timestamp:   timestamp,
		Restricted:  p.Private,
	}
	v.FilePath = "/v/" + p.Prefix + "/" + name //  strings.Replace(pth, "videos/", "", 1)
	fileExt := path.Ext(name)
	exists := false
	for _, ext := range extensions {
		if fileExt == ext {
			exists = true
			break
		}
	}
	if !exists {
		return nil, fmt.Errorf("Unsupported file format %s ", name)
	}
	name = strings.ReplaceAll(strings.Split(name, ".")[0], "_", " ")
	m, err := tag.ReadFrom(f)
	v.FileType = template.HTML(NewLibrary().GetContentType(fileExt))
	if err != nil {
				v.Title = name
				v.Album = "Unknown Album"
				v.Description = "NO Description"
				return v, nil
	}
	title := m.Title()
	// Default title is filename
	if title == "" {
		title = name
	}
	// print(title, "\n")
	v.Title = title
	v.Album = m.Album()
	v.Description = m.Comment()
	// Add thumbnail (if exists)
	pic := m.Picture()
	if pic != nil {
		v.Thumb = pic.Data
		v.ThumbType = pic.MIMEType
	}
	return v, nil
}
