package main

import (
	"log"
	"sync"
	"time"
	"utils/file/pkg"
)

func main() {
	uri := "https://dl.google.com/go/go1.15.2.src.tar.gz"
	dir := "download/go"
	var list []*FileInfo
	list = append(list, &FileInfo{DownUri: uri, FileName: "go1.15.2.src.tar.gz", FileDir: dir})
	list = append(list, &FileInfo{DownUri: uri, FileName: "go1.15.3.src.tar.gz", FileDir: dir})

	var wg sync.WaitGroup
	var errs []error
	for _, v := range list {
		wg.Add(1)
		go func(uri, dir, fileName string, timeOut time.Duration) {
			defer wg.Done()
			var er error
			isExists := pkg.FileIsExist(dir + "/" + fileName)
			if !isExists { //判断文件是否存在？不存在：io.Copy；存在：断点续传
				er = pkg.DownFile(uri, dir, fileName)
			}
			if isExists || er != nil {
				for i := 0; i < 3; i++ {
					er = pkg.DownBigFile(uri, dir, fileName, timeOut)
					if er == nil {
						break
					}
				}
			}
			if er != nil {
				errs = append(errs, er)
			}
		}(v.DownUri, v.FileDir, v.FileName, 15*time.Minute)
	}
	wg.Wait()

	if len(errs) > 1 {
		log.Println("down fail, ", errs[0])
	}
}

type FileInfo struct {
	DownUri  string
	FileName string
	FileDir  string
}
