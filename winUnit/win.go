package winUnit

import (
	"errors"
	"fmt"
	"github.com/lxn/win"
	"image"
	"regexp"
	"strconv"
	"syscall"
	"unsafe"
	"vedioCollector/util"
)

var (
	user32             = syscall.MustLoadDLL("user32.dll")
	procEnumWindows    = user32.MustFindProc("EnumWindows")
	procGetWindowTextW = user32.MustFindProc("GetWindowTextW")
	procGetWindowsDC   = user32.MustFindProc("GetWindowDC")
	procIsWindow       = user32.MustFindProc("IsWindow")
)

func EnumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.SyscallN(procEnumWindows.Addr(), uintptr(enumFunc), uintptr(lparam), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetWindowText(hwnd win.HWND, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.SyscallN(procGetWindowTextW.Addr(), uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func FindWindow(title string) (win.HWND, error) {
	var hwnd win.HWND
	var curWinName string
	cb := syscall.NewCallback(func(h win.HWND, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}

		curWinName = syscall.UTF16ToString(b)
		regx, _ := regexp.Compile(title)
		match := regx.MatchString(curWinName)
		if match {
			// note the window
			hwnd = h
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})
	err := EnumWindows(cb, 0)
	if err != nil {
		return 0, err
	}
	if hwnd == 0 {
		return 0, fmt.Errorf("No window with title '%s' found", title)
	}
	return hwnd, nil
}

func GetWindowsDC(hwdn win.HWND) win.HDC {
	ret, _, _ := syscall.SyscallN(procGetWindowsDC.Addr(), uintptr(hwdn), 0)
	return win.HDC(ret)
}

func IsWindow(hwdn win.HWND) bool {
	r1, _, _ := syscall.SyscallN(procIsWindow.Addr(), uintptr(hwdn), 0)
	res, _ := strconv.ParseBool(strconv.Itoa(int(r1)))
	return res
}

func ScreenshotWindow(hwdn win.HWND) (img *image.RGBA, err error) {
	hdc := GetWindowsDC(hwdn)
	if hdc == 0 {
		return nil, errors.New("GetDC failed")
	}
	defer win.ReleaseDC(hwdn, hdc)

	//拷贝创建一个新的视频缓存
	memoryDevice := win.CreateCompatibleDC(hdc)
	if memoryDevice == 0 {
		return nil, errors.New("CreateCompatibleDC failed")
	}
	defer win.DeleteDC(memoryDevice)

	// 获取原有窗口的位置
	var curHwdnRect win.RECT
	win.GetWindowRect(hwdn, &curHwdnRect)

	//计算截图的宽高
	height := float64(curHwdnRect.Bottom) - float64(curHwdnRect.Top)
	width := float64(curHwdnRect.Right) - float64(curHwdnRect.Left)

	rect := image.Rect(0, 0, int(width), int(height))

	//创建和源窗口一样宽高的一个RGB图像对象
	img, err = util.CreateImage(rect)
	if err != nil {
		return nil, err
	}

	bitmap := win.CreateCompatibleBitmap(hdc, int32(width), int32(height))
	if bitmap == 0 {
		return nil, errors.New("CreateCompatibleBitmap failed")
	}
	defer win.DeleteObject(win.HGDIOBJ(bitmap))

	var header win.BITMAPINFOHEADER
	header.BiSize = uint32(unsafe.Sizeof(header))
	header.BiPlanes = 1
	header.BiBitCount = 32
	header.BiWidth = int32(width)
	header.BiHeight = int32(-height)
	header.BiCompression = win.BI_RGB
	header.BiSizeImage = 0

	bitmapDataSize := uintptr(((int64(width)*int64(header.BiBitCount) + 31) / 32) * 4 * int64(height))
	hmem := win.GlobalAlloc(win.GMEM_MOVEABLE, bitmapDataSize)
	defer win.GlobalFree(hmem)
	memptr := win.GlobalLock(hmem)
	defer win.GlobalUnlock(hmem)

	old := win.SelectObject(memoryDevice, win.HGDIOBJ(bitmap))
	if old == 0 {
		return nil, errors.New("SelectObject failed")
	}
	defer win.SelectObject(memoryDevice, old)

	if !win.BitBlt(memoryDevice, 0, 0, int32(width), int32(height), hdc, int32(0), int32(0), win.SRCCOPY) {
		return nil, errors.New("BitBlt failed")
	}

	if win.GetDIBits(hdc, bitmap, 0, uint32(height), (*uint8)(memptr), (*win.BITMAPINFO)(unsafe.Pointer(&header)), win.DIB_RGB_COLORS) == 0 {
		return nil, errors.New("GetDIBits failed")
	}

	// 为每个像素点赋值
	i := 0
	src := uintptr(memptr)
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			v0 := *(*uint8)(unsafe.Pointer(src))
			v1 := *(*uint8)(unsafe.Pointer(src + 1))
			v2 := *(*uint8)(unsafe.Pointer(src + 2))

			// BGRA => RGBA, and set A to 255
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = v2, v1, v0, 255

			i += 4
			src += 4
		}
	}
	return img, err
}
