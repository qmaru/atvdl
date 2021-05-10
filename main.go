package main

import (
	"fmt"
	"strings"

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
	myWin.Resize(fyne.NewSize(400, 320))
	myWin.SetFixedSize(true)
	myWin.CenterOnScreen()

	uiPlaylist := widget.NewEntry()
	uiKey := widget.NewEntry()
	uiProxy := widget.NewEntry()
	uiHelp := widget.NewLabel("")
	services.UIProgress = widget.NewProgressBar()

	uiPlaylist.SetPlaceHolder("Playlist")
	uiKey.SetPlaceHolder("Key")
	uiProxy.SetPlaceHolder("Socks5 Proxy")

	uiGetKey := widget.NewButton("Get Key", func() {
		keyHelp(myWin)
	})

	var uiDownload *widget.Button
	uiDownload = widget.NewButton("Download", func() {
		url := uiPlaylist.Text
		key := uiKey.Text
		proxy := uiProxy.Text

		if url != "" && key != "" {
			uiDownload.DisableableWidget.Disable()
			services.UIProgress.SetValue(0)

			abema := services.AbemaTV

			norURL := strings.TrimSuffix(url, "\r")
			nonURL := strings.TrimSuffix(norURL, "\n")
			clearURL := strings.TrimSpace(nonURL)

			abema.SetProxy(proxy)
			if len(key) == 32 && abema.IPCheck(clearURL) {
				abema.PlaylistURL = strings.TrimSpace(clearURL)
				abema.Key = key

				uiHelp.SetText("[1] Get Best Playlist...")
				bestURL := abema.BestM3U8URL()
				if bestURL != "" {
					uiHelp.SetText("[2] Get Video List...")
					videos := abema.GetVideoInfo(bestURL)

					dlInfo := fmt.Sprintf("[3] Downloading...(%d)", len(videos))
					uiHelp.SetText(dlInfo)
					abema.DownloadCore(videos, 8)

					uiHelp.SetText("[4] Merging...")
					abema.Merge()
					services.UIProgress.SetValue(1)

					uiHelp.SetText(abema.Output)
					uiDownload.DisableableWidget.Enable()
				} else {
					uiDownload.DisableableWidget.Enable()
					uiHelp.SetText("url error")
				}
			} else {
				uiDownload.DisableableWidget.Enable()
				uiHelp.SetText("Please Set Proxy or key error")
			}
		} else {
			uiDownload.DisableableWidget.Enable()
			uiHelp.SetText("url or key is empty")
		}
		uiDownload.DisableableWidget.Enable()
	})

	content := container.NewVBox(
		uiPlaylist,
		uiGetKey,
		uiKey,
		uiProxy,
		services.UIProgress,
		uiDownload,
		uiHelp,
	)

	myWin.SetContent(content)
	myWin.ShowAndRun()
}
