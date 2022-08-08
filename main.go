package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"atvdl/services"
	"atvdl/theme"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// keyHelp Get AES KEY
func keyHelp(myWin fyne.Window) {
	step1 := "1.chrome F12"
	step2 := "2.source"
	step3 := "3.theoplayer.d.js"
	step4 := "4.break: " + services.JSConsole[0]
	step5 := "5.copy code to console"
	step6 := "6.copy to key"
	uiKeyTip := fmt.Sprintf("%-28s\n%-33s\n%-28s\n%-25s\n%-6s\n%-30s", step1, step2, step3, step4, step5, step6)
	uiKeyDialog := dialog.NewConfirm("Get Key", uiKeyTip, func(res bool) {
		if res {
			myWin.Clipboard().SetContent(services.JSConsole[1])
		}
	}, myWin)

	uiKeyDialog.SetDismissText("Cancel")
	uiKeyDialog.SetConfirmText("Copy")
	uiKeyDialog.Show()
}

func main() {
	myApp := app.New()
	myApp.SetIcon(theme.MyLogo())
	myWin := myApp.NewWindow("AbemaTV")
	myWin.Resize(fyne.NewSize(800, 600))
	myWin.SetFixedSize(true)
	myWin.CenterOnScreen()

	uiPlaylist := widget.NewEntry()
	uiKey := widget.NewEntry()
	uiProxy := widget.NewEntry()
	uiOutput := widget.NewEntry()
	uiHelp := widget.NewLabel("")
	uiHelp.Wrapping = fyne.TextWrapBreak
	services.UIProgress = widget.NewProgressBar()

	uiPlaylist.SetPlaceHolder("Playlist")
	uiKey.SetPlaceHolder("Key")
	uiProxy.SetPlaceHolder("Socks5 Proxy")
	uiOutput.SetPlaceHolder("Output (Default: ./)")

	uiGetKey := widget.NewButton("Get Key", func() {
		keyHelp(myWin)
	})

	var uiDownload *widget.Button
	uiDownload = widget.NewButton("Download", func() {
		url := uiPlaylist.Text
		key := uiKey.Text
		proxy := uiProxy.Text
		output := uiOutput.Text

		if url != "" && key != "" {
			uiDownload.DisableableWidget.Disable()
			services.UIProgress.SetValue(0)

			abema := new(services.AbemaTVBasic)

			norURL := strings.TrimSuffix(url, "\r")
			nonURL := strings.TrimSuffix(norURL, "\n")
			playlistURL := strings.TrimSpace(nonURL)

			if services.AbemaURLCheck(playlistURL) {
				abema.SetProxy(proxy)
				ipValid, err := services.AbemaIPCheck(playlistURL)
				if err != nil {
					uiHelp.SetText(err.Error())
				} else {
					if len(key) == 32 && ipValid {
						abema.PlaylistURL = playlistURL
						abema.Key = key

						uiHelp.SetText("[1] Get Best Playlist...")
						bestURL, err := abema.BestM3U8URL()
						if err != nil {
							uiHelp.SetText(err.Error())
						} else {
							uiHelp.SetText("[2] Get Video List...")
							videos, err := abema.GetVideoInfo(bestURL)
							if err != nil {
								uiHelp.SetText(err.Error())
							} else {
								dlInfo := fmt.Sprintf("[3] Downloading...(%d)", len(videos))
								uiHelp.SetText(dlInfo)

								var output_root string
								if output == "" {
									localPath, _ := services.LocalPath("")
									output_root = localPath
								} else {
									output_root = output
								}

								output_folder := fmt.Sprintf("decrypt_%s", time.Now().Format("2006-01-02-15-04-05"))
								output_dir := filepath.Join(output_root, output_folder)
								os.Mkdir(output_dir, 0644)

								abema.Output = output_dir
								err = abema.DownloadCore(videos, 4)
								if err != nil {
									uiHelp.SetText(err.Error())
								} else {
									uiHelp.SetText("[4] Merging...")
									err = abema.Merge()
									if err != nil {
										uiHelp.SetText(err.Error())
									} else {
										services.UIProgress.SetValue(1)
										uiHelp.SetText(output_dir)
										uiDownload.DisableableWidget.Enable()
									}
								}
							}
						}
					} else {
						uiDownload.DisableableWidget.Enable()
						if len(key) != 32 {
							uiHelp.SetText("Key length error")
						} else if !ipValid {
							uiHelp.SetText("Please set socks5 proxy")
						}
					}
				}
			} else {
				uiDownload.DisableableWidget.Enable()
				uiHelp.SetText("Playlist Error")
			}
		} else {
			uiDownload.DisableableWidget.Enable()
			uiHelp.SetText("Playlist or Key is empty")
		}
		uiDownload.DisableableWidget.Enable()
	})

	content := container.NewVBox(
		uiPlaylist,
		uiGetKey,
		uiKey,
		uiOutput,
		uiProxy,
		services.UIProgress,
		uiDownload,
		uiHelp,
	)

	myWin.SetContent(content)
	myWin.ShowAndRun()
}
