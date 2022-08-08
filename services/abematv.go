package services

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"atvdl/utils"

	"fyne.io/fyne/v2/widget"
)

var (
	UIProgress *widget.ProgressBar
	JSConsole  []string = []string{
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
	progress    float64
}

// LocalPath 自动判断根目录路径
func LocalPath(subPath string) (string, error) {
	// go build 可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	runPath, err := filepath.EvalSymlinks(filepath.Dir(exePath))
	if err != nil {
		return "", err
	}
	// go run 调试路径
	buildPath, err := filepath.EvalSymlinks(os.Getenv("GOTMPDIR"))
	if err != nil {
		return "", err
	}

	if strings.Contains(runPath, buildPath) {
		var absPath string
		// 获取当前文件 config.go 的路径
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			// 获取上上级目录 即根目录
			absPath = filepath.Dir(filepath.Dir(filename))
		}
		absPath = filepath.Join(absPath, subPath)
		return absPath, nil
	}
	runPath = filepath.Join(runPath, subPath)
	return runPath, nil
}

// AbemaURLCheck 检查URL
func AbemaURLCheck(url string) bool {
	urlRule := regexp.MustCompile(`https?://.*abematv.*`)
	result := urlRule.MatchString(url)
	return result
}

// AbemaIPCheck 检查IP
func AbemaIPCheck(url string) (bool, error) {
	headers := utils.MiniHeaders{"User-Agent": utils.UserAgent}
	res, err := utils.Minireq.Get(url, headers)
	if err != nil {
		return false, err
	}
	return res.Response.StatusCode == 200, nil
}

// fetchData 获取远程数据
func (atv *AbemaTVBasic) fetchData(url string) ([]byte, error) {
	headers := utils.MiniHeaders{"User-Agent": utils.UserAgent}
	res, err := utils.Minireq.Get(url, headers)
	if err != nil {
		return nil, err
	}
	rawData, err := res.RawData()
	if err != nil {
		return nil, err
	}
	return rawData, nil
}

// setM3U8Host 设置 M3U8 的地址
func (atv *AbemaTVBasic) setM3U8Host() string {
	m3u8Host := strings.Split(atv.PlaylistURL, "playlist.m3u8")
	if len(m3u8Host) != 0 {
		return m3u8Host[0]
	}
	return ""
}

// setVideoHost 设置 Video 的地址
func (atv *AbemaTVBasic) setVideoHost() string {
	videoHost := strings.Split(atv.PlaylistURL, "program")
	if len(videoHost) != 0 {
		return videoHost[0]
	}
	return ""
}

// hexStr 转换 key 和 iv 为16进制
func (atv *AbemaTVBasic) hexStr(s string) ([]byte, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetProxy 设置 Socks5
func (atv *AbemaTVBasic) SetProxy(proxy string) {
	utils.Minireq.Socks5Address = proxy
}

// rexFilterData 使用正则过滤数据
//
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
func (atv *AbemaTVBasic) dlcore(u interface{}) (interface{}, error) {
	url := u.(string)
	urlInfo := strings.Split(url, "#")
	urlNew := urlInfo[0]
	urlCode := urlInfo[1]
	savepath := filepath.Join(atv.Output, urlCode+"_dec.ts")

	headers := utils.MiniHeaders{"User-Agent": utils.UserAgent}
	res, err := utils.Minireq.Get(urlNew, headers)
	if err != nil {
		return nil, err
	}

	rawData, err := res.RawData()
	if err != nil {
		return nil, err
	}
	hexKey, err := atv.hexStr(atv.Key)
	if err != nil {
		return nil, err
	}
	hexIV, err := atv.hexStr(atv.iv)
	if err != nil {
		return nil, err
	}

	decData, err := utils.AESSuite.Decrypt(rawData, hexKey, hexIV)
	if err != nil {
		return nil, err
	}

	dst, err := os.Create(savepath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, bytes.NewReader(decData))
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(atv.Output)
	if err != nil {
		return nil, err
	}
	UIProgress.SetValue(atv.progress * float64(len(files)))

	time.Sleep(time.Second * 3)
	return "", nil
}

// BestM3U8URL 获取最佳分辨率的 URL
func (atv *AbemaTVBasic) BestM3U8URL() (string, error) {
	// data, err := utils.FileSuite.Read("demo/playlist.txt")
	// if err != nil {
	// 	return "", err
	// }
	data, err := atv.fetchData(atv.PlaylistURL)
	if err != nil {
		return "", err
	}

	videoDataList := atv.rexFilterData(`(?m)^#EXT-X-STREAM-INF.*`, string(data), 1)
	bestmatch := 0
	bandwidthMax := 0
	for bindex, videoData := range videoDataList {
		bandwidthDataSplit := strings.Split(videoData, ",")
		bandwidthData := bandwidthDataSplit[1]
		bandwidthKVSplit := strings.Split(bandwidthData, "=")
		bandwidthStr := bandwidthKVSplit[1]
		bandwidth, err := strconv.Atoi(bandwidthStr)
		if err != nil {
			return "", err
		}
		if bandwidth > bandwidthMax {
			bestmatch = bindex
		}
	}

	m3u8List := atv.rexFilterData(`(?m)^[\d]+.*`, string(data), 1)
	m3u8URL := m3u8List[bestmatch]
	return m3u8URL, nil
}

// GetVideoInfo 获取视频信息
func (atv *AbemaTVBasic) GetVideoInfo(m3u8URL string) ([]string, error) {
	m3u8Host := atv.setM3U8Host()
	m3u8URL = m3u8Host + m3u8URL

	videoHost := atv.setVideoHost()
	videoHost = videoHost[0 : len(videoHost)-1]

	// data, err := utils.FileSuite.Read("demo/m3u8.txt")
	// if err != nil {
	// 	return nil, err
	// }
	data, err := atv.fetchData(m3u8URL)
	if err != nil {
		return nil, err
	}

	ivData := atv.rexFilterData(`(?m)IV=0x([\w]+)`, string(data), 2)
	atv.iv = ivData[1]

	var videos []string
	vData := atv.rexFilterData(`(?m)^[^#].*`, string(data), 1)
	for i, v := range vData {
		video := fmt.Sprintf("%s%s#%04d", videoHost, v, i)
		videos = append(videos, video)
	}
	return videos, nil
}

// Merge 合并视频
func (atv *AbemaTVBasic) Merge() error {
	output := atv.Output
	mergeName := fmt.Sprintf("%s_all.ts", filepath.Base(output))
	mergePath := filepath.Join(filepath.Dir(output), mergeName)

	decVideoList, err := os.ReadDir(output)
	if err != nil {
		return err
	}
	mergeData, err := os.OpenFile(mergePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer mergeData.Close()

	if err != nil {
		return err
	}
	for _, video := range decVideoList {
		videoPath := filepath.Join(output, video.Name())
		videoFile, err := os.Open(videoPath)
		if err != nil {
			return err
		}
		videoBytes, err := io.ReadAll(videoFile)
		if err != nil {
			return err
		}
		_, err = mergeData.Write(videoBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

// DownloadCore 下载视频
func (atv *AbemaTVBasic) DownloadCore(videos []string, thread int) error {
	atv.progress = 0.8 / float64(len(videos))
	_, err := utils.TaskBoard(atv.dlcore, videos, thread)
	return err
}
