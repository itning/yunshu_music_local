<h3 align="center">云舒音乐本地服务端</h3>

<div align="center">

[![GitHub stars](https://img.shields.io/github/stars/itning/yunshu_music_local.svg?style=social&label=Stars)](https://github.com/itning/yunshu_music_local/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/itning/yunshu_music_local.svg?style=social&label=Fork)](https://github.com/itning/yunshu_music_local/network/members)
[![GitHub watchers](https://img.shields.io/github/watchers/itning/yunshu_music_local.svg?style=social&label=Watch)](https://github.com/itning/yunshu_music_local/watchers)
[![GitHub followers](https://img.shields.io/github/followers/itning.svg?style=social&label=Follow)](https://github.com/itning?tab=followers)

</div>

<div align="center">

[![GitHub issues](https://img.shields.io/github/issues/itning/yunshu_music_local.svg)](https://github.com/itning/yunshu_music_local/issues)
[![GitHub license](https://img.shields.io/github/license/itning/yunshu_music_local.svg)](https://github.com/itning/yunshu_music_local/blob/master/LICENSE)
[![GitHub last commit](https://img.shields.io/github/last-commit/itning/yunshu_music_local.svg)](https://github.com/itning/yunshu_music_local/commits)
[![GitHub repo size in bytes](https://img.shields.io/github/repo-size/itning/yunshu_music_local.svg)](https://github.com/itning/yunshu_music_local)
[![Hits](https://hitcount.itning.com?u=itning&r=yunshu_music_local)](https://github.com/itning/hit-count)
[![language](https://img.shields.io/badge/language-Dart-green.svg)](https://github.com/itning/yunshu_music_local)

</div>

## 功能概述

- 自动扫描指定目录下的音频文件（支持 `.flac`, `.mp3`, `.wav`, `.aac`）
- 提取音频元数据（标题、艺术家等）
- 支持通过 HTTP 查询音乐列表
- 支持在线播放、下载音频文件
- 支持查找对应的 `.lrc` 歌词文件
- 支持提取音频封面图（如 ID3 中包含）

---

## 安装与运行

### 启动服务

```bash
./yunshu_music_local -d /path/to/music/dir -p 8080
```

- `-d` 指定音频文件所在目录（必须）
- `-p` 指定监听端口号，默认 `8080`

---

## API 接口说明

### 获取所有音乐信息

**GET** `/music`

#### 响应示例：

```json
{
  "code": 200,
  "msg": "查询成功",
  "data": [
    {
      "musicId": "1",
      "name": "歌曲名称",
      "singer": "歌手名字",
      "lyricId": "1",
      "type": 2,
      "musicUri": "http://localhost:8080/music/file/1",
      "musicDownloadUri": "http://localhost:8080/music/file/download/1",
      "lyricUri": "http://localhost:8080/music/lyric/1",
      "coverUri": "http://localhost:8080/music/cover/1"
    }
  ]
}
```

---

### 获取音频文件（在线播放）

**GET** `/music/file/{musicId}`

返回音频文件流，可直接用于 HTML5 `<audio>` 标签播放。

---

### 下载音频文件

**GET** `/music/file/download/{musicId}`

触发浏览器下载该音频文件。

---

### 获取歌词文件

**GET** `/music/lyric/{musicId}`

如果存在同名 `.lrc` 文件，则返回歌词内容。

---

### 获取音频封面图

**GET** `/music/cover/{musicId}`

从音频文件中提取的封面图（仅限嵌入在 ID3 等标签中的图像）。
