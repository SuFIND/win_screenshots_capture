# win_screenshots_capture
## Description 描述
custom capture windows program screenshots</br>
自定义捕获windows程序的截图小工具

## Quick Start 快速开始
1. Create your configuration file cfg.toml in the same directory as the program</br>
在程序所在的目录创建你的配置文件 cfg.toml
    ```toml
    WindowName="Apex" #The name of the window you want to capture 需要捕获的窗口的名称
    ScreenshotIntervalMilliSec=1500  #Screenshot capture interval 截图捕获间隔
    ImgQuality=50  #jpg image quality(0~100) jpg图片质量(0~100)
    ```
2. go run main.go or go build main.go and run exe program</br>
执行go run main.go运行或者使用go build 编译