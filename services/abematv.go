package services

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"atvdl/utils"

	"fyne.io/fyne/v2/widget"
)

var (
	decFolder    = ""
	decName      = ""
	UIProgress   *widget.ProgressBar
	progressStep float64 = 0.0
	s5Proxy              = ""
	JSConsole            = []string{
		"var n = t.data",
		"Array.from(t.data.iyt, function(byte){return ('0' + (byte & 0xFF).toString(16)).slice(-2);}).join('')",
	}
)

// AbemaTVBasic 基类
type AbemaTVBasic struct {
	PlaylistURL string
	Key         string
	Output      string
	iv          string
	m3u8Host    string
	videoHost   string
	videos      []interface{}
}

var AbemaTV *AbemaTVBasic

func init() {
	AbemaTV = new(AbemaTVBasic)
}

// IPCheck 检查IP
func (atv *AbemaTVBasic) URLCheck(url string) bool {
	urlRule := regexp.MustCompile(`https?://.*abematv.*`)
	result := urlRule.MatchString(url)
	return result
}

// IPCheck 检查IP
func (atv *AbemaTVBasic) IPCheck(url string) bool {
	defer func() {
		if e := recover(); e != nil {
			return
		}
	}()
	headers := utils.MiniHeaders{
		"User-Agent": utils.UserAgent,
	}
	if s5Proxy != "" {
		utils.Minireq.Proxy(s5Proxy)
	}
	res := utils.Minireq.Get(url, headers)
	return res.RawRes.StatusCode == 200
}

// fetchData 获取远程数据
func (atv *AbemaTVBasic) fetchData(url string) (data string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			return
		}
	}()
	headers := utils.MiniHeaders{
		"User-Agent": utils.UserAgent,
	}
	if s5Proxy != "" {
		utils.Minireq.Proxy(s5Proxy)
	}
	res := utils.Minireq.Get(url, headers)
	data = string(res.RawData())
	return
}

// setM3U8Host 设置 M3U8 的地址
func (atv *AbemaTVBasic) setM3U8Host() {
	atv.m3u8Host = strings.Split(atv.PlaylistURL, "playlist.m3u8")[0]
}

// setVideoHost 设置 Video 的地址
func (atv *AbemaTVBasic) setVideoHost() {
	atv.videoHost = strings.Split(atv.PlaylistURL, "program")[0]
}

// hexStr 转换 key 和 iv 为16进制
func (atv *AbemaTVBasic) hexStr(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// SetProxy 使用代理下载
func (atv *AbemaTVBasic) SetProxy(proxy string) {
	s5Proxy = proxy
}

// rexFilterData 使用正则过滤数据
//	mode:
//	1: FindAllString 全匹配
//	2: FindAllStringSubmatch 精确匹配
func (atv *AbemaTVBasic) rexFilterData(rule string, data string, mode int) []string {
	reg := regexp.MustCompile(rule)
	var results []string
	switch mode {
	case 1:
		results = reg.FindAllString(data, -1)
	case 2:
		resultsSub := reg.FindAllStringSubmatch(data, -1)
		if len(resultsSub) != 0 {
			results = resultsSub[0]
		} else {
			return []string{}
		}
	}

	if len(results) > 0 {
		return results
	}
	return []string{}
}

// dlcore 下载函数
func (atv *AbemaTVBasic) dlcore(u interface{}) interface{} {
	url := u.(string)
	urlInfo := strings.Split(url, "#")
	urlNew := urlInfo[0]
	urlCode := urlInfo[1]
	savepath := filepath.Join(decFolder, urlCode+"_dec.ts")

	fmt.Println("Download: ", urlNew)
	request := utils.NewHTTP(s5Proxy)
	res := request.Get(urlNew)

	rawData := res.RawData()
	hexKey := atv.hexStr(atv.Key)
	kexIV := atv.hexStr(atv.iv)

	decData := utils.AESSuite.Decrypt(rawData, hexKey, kexIV)

	dst, err := os.Create(savepath)
	if err != nil {
		fmt.Println(err)
	}

	io.Copy(dst, bytes.NewReader(decData))
	defer dst.Close()

	files, _ := ioutil.ReadDir(decFolder)
	UIProgress.SetValue(progressStep * float64(len(files)))

	time.Sleep(time.Second * 3)
	return nil
}

// BestM3U8URL 获取最佳分辨率的 URL
func (atv *AbemaTVBasic) BestM3U8URL() (m3u8URL string) {
	// data := string(utils.FileSuite.Read("demo/playlist.txt"))
	data := atv.fetchData(atv.PlaylistURL)

	if data != "" {
		m3u8List := atv.rexFilterData(`(?m)^[\d]+.*`, data, 1)
		videoDataList := atv.rexFilterData(`(?m)^#EXT-X-STREAM-INF.*`, data, 1)

		bestmatch := 0
		bandwidthMax := 0
		for bindex, videoData := range videoDataList {
			bandwidthDataSplit := strings.Split(videoData, ",")
			bandwidthData := bandwidthDataSplit[1]
			bandwidthKVSplit := strings.Split(bandwidthData, "=")
			bandwidthStr := bandwidthKVSplit[1]
			bandwidth, _ := strconv.Atoi(bandwidthStr)
			if bandwidth > bandwidthMax {
				bestmatch = bindex
			}
		}
		m3u8URL = m3u8List[bestmatch]
	}
	return
}

// GetVideoInfo 获取视频信息
func (atv *AbemaTVBasic) GetVideoInfo(m3u8URL string) []interface{} {
	atv.videos = []interface{}{}
	// 视频列表的
	atv.setM3U8Host()
	m3u8URL = atv.m3u8Host + m3u8URL

	// 视频下载地址的 HOST
	atv.setVideoHost()
	videoHost := atv.videoHost
	// 移除末尾的"/"
	videoHost = videoHost[0 : len(videoHost)-1]

	// data := string(utils.FileSuite.Read("demo/m3u8.txt"))
	data := atv.fetchData(m3u8URL)

	ivData := atv.rexFilterData(`(?m)IV=0x([\w]+)`, data, 2)
	atv.iv = ivData[1]

	vData := atv.rexFilterData(`(?m)^[^#].*`, data, 1)
	for i, v := range vData {
		video := fmt.Sprintf("%s%s#%04d", videoHost, v, i)
		atv.videos = append(atv.videos, video)
	}
	return atv.videos
}

// Merge 合并视频
func (atv *AbemaTVBasic) Merge() {
	outputVideo := decName + "_all.ts"
	decVideoList, _ := ioutil.ReadDir(decFolder)
	fileAll, _ := os.OpenFile(outputVideo, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	for _, video := range decVideoList {
		videoPath := filepath.Join(decFolder, video.Name())
		videoFile, _ := os.Open(videoPath)
		videoBytes, _ := ioutil.ReadAll(videoFile)
		fileAll.Write(videoBytes)
	}
	defer fileAll.Close()
	atv.Output = outputVideo
}

// DownloadCore 下载视频
func (atv *AbemaTVBasic) DownloadCore(videos []interface{}, thread int) {
	// 视频保存目录
	now := time.Now().Format("20060102150405")
	decName = "decrypt_" + now
	decFolder = filepath.Join(utils.FileSuite.LocalPath(true), decName)
	utils.FileSuite.Create(decFolder)
	// 下载视频
	progressStep = 0.8 / float64(len(videos))
	utils.TaskBoard(atv.dlcore, videos, thread)
}

// AtvDL 调用主函数
func AtvDL(playlistURL, key, proxy string) {
	if AbemaTV.IPCheck(playlistURL) {
		AbemaTV.PlaylistURL = playlistURL
		AbemaTV.Key = key
		AbemaTV.SetProxy(proxy)

		fmt.Println("[1] Get Best Playlist...")
		bestURL := AbemaTV.BestM3U8URL()
		fmt.Printf("  - [URL] %s\n", bestURL)

		fmt.Println("[2] Get Video List...")
		videos := AbemaTV.GetVideoInfo(bestURL)
		fmt.Printf("  [Video] %d\n", len(videos))

		fmt.Println("[3] Downloading...")
		AbemaTV.DownloadCore(videos, 8)

		fmt.Println("[4] Merging...")
		AbemaTV.Merge()
	} else {
		fmt.Println("Please Set Proxy")
	}
}
