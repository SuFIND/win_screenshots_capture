package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/lxn/win"
	"log"
	"os"
	"syscall"
	"time"
	"vedioCollector/util"
	"vedioCollector/winUnit"
)

type MyConfig struct {
	WindowName                 string
	ScreenshotIntervalMilliSec int
	ImgQuality                 int
}

func worker(hwdn win.HWND, imgQuality int) {
	img, err1 := winUnit.ScreenshotWindow(hwdn)
	if err1 != nil {
		return
	}

	filename := time.Now().UnixNano()
	subDirName := time.Now().Format("20060102")
	subDirPath := fmt.Sprintf("img/%v", subDirName)
	_, errS := os.Stat(subDirPath)
	if errS != nil {
		err2 := os.MkdirAll(subDirPath, os.FileMode(777))
		if err2 != nil {
			log.Fatalf("Create img dir %v failure!", subDirPath)
			return
		}
	}

	imgPath := fmt.Sprintf("%v/%v.jpg", subDirPath, filename)
	util.SaveJPEG(img, imgPath, imgQuality)
}

func main() {
	var config MyConfig
	_, err1 := toml.DecodeFile("cfg.toml", &config)
	if err1 != nil {
		log.Fatalln("Can't find config file cfg.toml in cur dir!")
		return
	}

	for {
		hwdn, err3 := winUnit.FindWindow(config.WindowName)
		if err3 != nil {
			// 如果，没有检测到Apex窗口程序，则休息一分钟在检测
			log.Println("failure to search program: " + config.WindowName + ". Will retry after 5 sec.")
			time.Sleep(5 * time.Second)
			continue
		}

		temp := make([]uint16, 200)
		_, _ = winUnit.GetWindowText(hwdn, &temp[0], int32(len(temp)))
		hwdnName := syscall.UTF16ToString(temp)
		log.Println("Success to listen program: " + hwdnName)

		for {
			isHwdnKeepLive := winUnit.IsWindow(hwdn)
			if isHwdnKeepLive != true {
				log.Println("Listen over!" + hwdnName)
				break
			}
			go worker(hwdn, config.ImgQuality)
			time.Sleep(time.Duration(config.ScreenshotIntervalMilliSec) * time.Millisecond)
		}

	}
}
