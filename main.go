package main

import (
	"encoding/json"
	"fmt"
	tag "github.com/unitnotes/audiotag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Music 定义音乐文件结构
type Music struct {
	MusicId  string `json:"musicId"`
	Name     string `json:"name"`
	Singer   string `json:"singer"`
	LyricId  string `json:"lyricId"`
	Type     int    `json:"type"`
	MusicUri string `json:"musicUri"`
	LyricUri string `json:"lyricUri"`
	CoverUri string `json:"coverUri"`
	FilePath string `json:"-"`
}

// Response 定义API响应结构
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
	// 获取启动时指定的目录或文件参数（最多5个）
	paths := os.Args[1:]
	if len(paths) == 0 {
		log.Fatal("请在启动时指定至少一个目录或文件")
	}
	if len(paths) > 5 {
		log.Println("警告：最多只处理前5个目录或文件")
		paths = paths[:5]
	}

	// 遍历指定目录，查找音乐文件
	for _, path := range paths {
		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 如果是文件且扩展名匹配
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(filePath))
				if musicExtMap[ext] {
					processMusicFile(filePath)
				}
			}

			return nil
		})

		if err != nil {
			log.Printf("遍历路径 %s 时出错: %v\n", path, err)
		}
	}

	// HTTP路由
	http.HandleFunc("/music", musicHandler)
	http.HandleFunc("/music/file/", musicFileHandler)
	http.HandleFunc("/music/lyric/", lyricHandler)
	http.HandleFunc("/music/cover/", coverHandler)

	// 启动HTTP服务
	log.Println("服务已启动，正在监听 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func processMusicFile(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("无法打开文件: %v\n", err)
		return
	}
	m, err := tag.ReadFrom(f)
	if err != nil {
		log.Printf("读取标签失败: %v\n", err)
		return
	}

	// 创建唯一ID（这里简单使用文件路径哈希）
	musicId := strconv.FormatInt(int64(filePath[len(filePath)-1]), 10)

	// 构建音乐对象
	music := Music{
		MusicId:  musicId,
		Name:     m.Title(),
		Singer:   m.Artist(),
		LyricId:  musicId,
		Type:     determineMusicType(filePath),
		MusicUri: "/music/file/" + musicId,
		LyricUri: "/music/lyric/" + musicId,
		CoverUri: "/music/cover/" + musicId,
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
	// 设置CORS头部
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 设置JSON响应内容
	response := Response{
		Code: 200,
		Msg:  "查询成功",
		Data: musicList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func musicFileHandler(w http.ResponseWriter, r *http.Request) {
	musicId := strings.TrimPrefix(r.URL.Path, "/music/file/")
	if val, ok := musicIdMap[musicId]; ok {
		http.ServeFile(w, r, val.FilePath)
	} else {
		http.NotFound(w, r)
	}
}

func lyricHandler(w http.ResponseWriter, r *http.Request) {
	musicId := strings.TrimPrefix(r.URL.Path, "/music/lyric/")
	if val, ok := musicIdMap[musicId]; ok {
		lyricPath := strings.TrimSuffix(val.FilePath, filepath.Ext(val.FilePath)) + ".lrc"
		// 检查歌词文件是否存在
		if _, err := os.Stat(lyricPath); err == nil {
			http.ServeFile(w, r, lyricPath)
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
		return nil, "", fmt.Errorf("无封面")
	}

	return pic.Data, pic.MIMEType, nil
}

func coverHandler(w http.ResponseWriter, r *http.Request) {
	musicId := strings.TrimPrefix(r.URL.Path, "/music/cover/")
	if val, ok := musicIdMap[musicId]; ok {
		coverData, mimeType, err := extractCover(val.FilePath)
		if err != nil {
			http.Error(w, "封面未找到", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", mimeType)
		w.Write(coverData)
	} else {
		http.NotFound(w, r)
	}
}
