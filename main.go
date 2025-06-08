package main

import (
	"encoding/json"
	"flag"
	"fmt"
	tag "github.com/unitnotes/audiotag"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Music defines the structure of music file
type Music struct {
	MusicId          string `json:"musicId"`
	Name             string `json:"name"`
	Singer           string `json:"singer"`
	LyricId          string `json:"lyricId"`
	Type             int    `json:"type"`
	MusicUri         string `json:"musicUri"`
	MusicDownloadUri string `json:"musicDownloadUri"`
	LyricUri         string `json:"lyricUri"`
	CoverUri         string `json:"coverUri"`
	FilePath         string `json:"-"`
}

// Response defines API response structure
type Response struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data []Music `json:"data"`
}

var musicList []Music
var musicIdMap = make(map[string]Music)
var musicExtMap = map[string]bool{
	".flac": true,
	".mp3":  true,
	".wav":  true,
	".aac":  true,
}

func main() {
	var (
		portFlag string
		dirFlag  string
	)

	flag.StringVar(&portFlag, "p", "8080", "Specify the port for service monitoring")
	flag.StringVar(&dirFlag, "d", "", "Specify the directory where the music file resides")
	flag.Parse()

	if dirFlag == "" {
		log.Fatal("A directory must be specified through -d")
	}

	path := dirFlag

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If it is a file and the extension matches
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(filePath))
			if musicExtMap[ext] {
				processMusicFile(filePath)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("An error occurred while traversing the directory %s: %v\n", path, err)
	}

	// HTTP routing
	http.HandleFunc("/music", musicHandler)
	http.HandleFunc("/music/file/download/", musicFileDownloadHandler)
	http.HandleFunc("/music/file/", musicFileHandler)
	http.HandleFunc("/music/lyric/", lyricHandler)
	http.HandleFunc("/music/cover/", coverHandler)

	// Start HTTP service
	log.Printf("The service has been started and is listening :%s\n", portFlag)
	log.Fatal(http.ListenAndServe(":"+portFlag, nil))
}

func processMusicFile(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("Unable to open the file: %v\n", err)
		return
	}
	m, err := tag.ReadFrom(f)
	if err != nil {
		log.Printf("Failed to read the tag: %v\n", err)
		return
	}

	// Create a unique ID (simply use file path hash here)
	musicId := strconv.FormatInt(int64(filePath[len(filePath)-1]), 10)
	// Building music objects
	music := Music{
		MusicId:  musicId,
		Name:     m.Title(),
		Singer:   m.Artist(),
		LyricId:  musicId,
		Type:     determineMusicType(filePath),
		FilePath: filePath,
	}

	musicList = append(musicList, music)
	musicIdMap[musicId] = music
}

func determineMusicType(filename string) int {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".flac":
		return 1
	case ".mp3":
		return 2
	case ".wav":
		return 3
	case ".aac":
		return 4
	default:
		return 0
	}
}

func musicHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	for i := range musicList {
		music := &musicList[i]
		music.MusicUri = fmt.Sprintf("http://%s/music/file/%s", r.Host, music.MusicId)
		music.CoverUri = fmt.Sprintf("http://%s/music/cover/%s", r.Host, music.MusicId)
		music.LyricUri = fmt.Sprintf("http://%s/music/lyric/%s", r.Host, music.MusicId)
		music.MusicDownloadUri = fmt.Sprintf("http://%s/music/file/download/%s", r.Host, music.MusicId)
	}
	response := Response{
		Code: 200,
		Msg:  "查询成功",
		Data: musicList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getContentTypeByMusicType(musicType int) string {
	switch musicType {
	case 1: // FLAC
		return "audio/flac"
	case 2: // MP3
		return "audio/mpeg"
	case 3: // WAV
		return "audio/wav"
	case 4: // AAC
		return "audio/aac"
	default:
		return "application/octet-stream"
	}
}

func musicFileHandler(w http.ResponseWriter, r *http.Request) {
	musicId := strings.TrimPrefix(r.URL.Path, "/music/file/")
	if val, ok := musicIdMap[musicId]; ok {
		filePath := val.FilePath
		f, err := os.Open(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		var contentType = getContentTypeByMusicType(val.Type)
		w.Header().Set("Content-Type", contentType)

		filename := filepath.Base(filePath)
		encodedFilename := url.PathEscape(filename)
		headerValue := fmt.Sprintf("inline; filename=\"%s\"; filename*=UTF-8''%s", filename, encodedFilename)
		w.Header().Set("Content-Disposition", headerValue)

		modTime := time.Time{}
		http.ServeContent(w, r, filename, modTime, f)
	} else {
		http.NotFound(w, r)
	}
}

func musicFileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	musicId := strings.TrimPrefix(r.URL.Path, "/music/file/download/")
	music, ok := musicIdMap[musicId]
	if !ok {
		http.NotFound(w, r)
		return
	}

	filename := filepath.Base(music.FilePath)
	encodedFilename := url.PathEscape(filename)
	headerValue := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", filename, encodedFilename)
	w.Header().Set("Content-Disposition", headerValue)

	http.ServeFile(w, r, music.FilePath)
}

func lyricHandler(w http.ResponseWriter, r *http.Request) {
	musicId := strings.TrimPrefix(r.URL.Path, "/music/lyric/")
	if val, ok := musicIdMap[musicId]; ok {
		lyricPath := strings.TrimSuffix(val.FilePath, filepath.Ext(val.FilePath)) + ".lrc"

		if _, err := os.Stat(lyricPath); err == nil {
			f, err := os.Open(lyricPath)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer f.Close()

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")

			filename := filepath.Base(lyricPath)
			modTime := time.Time{}
			http.ServeContent(w, r, filename, modTime, f)
		} else {
			http.NotFound(w, r)
		}
	} else {
		http.NotFound(w, r)
	}
}

func extractCover(filePath string) ([]byte, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, "", err
	}

	pic := m.Picture()
	if pic == nil {
		return nil, "", fmt.Errorf("NO COVER")
	}

	return pic.Data, pic.MIMEType, nil
}

func coverHandler(w http.ResponseWriter, r *http.Request) {
	musicId := strings.TrimPrefix(r.URL.Path, "/music/cover/")
	if val, ok := musicIdMap[musicId]; ok {
		coverData, mimeType, err := extractCover(val.FilePath)
		if err != nil {
			http.Error(w, "Cover not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", mimeType)
		w.Write(coverData)
	} else {
		http.NotFound(w, r)
	}
}
