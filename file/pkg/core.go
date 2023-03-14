package pkg

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func FileIsExist(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

/**
 * 下载大文件，支持断点续传
 * uri: 待下载地址
 * dir: 下载文件的存放目录，相对于项目根目录开始
 * fileName: 下载后的文件名称
 * timeOut: 超时时间
 */
func DownBigFile(uri, dir, fileName string, timeOut time.Duration) error {
	log.Printf("DownBigFile param uri:%v, dir:%v, fileName:%v, time:%v\n", uri, dir, fileName, timeOut)
	if !FileIsExist(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Println("DownBigFile os.MkdirAll err:", err.Error())
			return err
		}
	}
	dfn := dir + "/" + fileName
	var file *os.File
	var size int64
	if FileIsExist(dfn) {
		fi, err := os.OpenFile(dfn, os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Println("DownBigFile os.OpenFile err:", err)
			return err
		}
		stat, _ := fi.Stat()
		size = stat.Size()
		sk, err := fi.Seek(size, 0)
		if err != nil {
			log.Println("DownBigFile fi.Seek err:", err)
			_ = fi.Close()
			return err
		}
		if sk != size {
			log.Printf("DownBigFile seek length not equal file size, seek=%d,size=%d\n", sk, size)
			_ = fi.Close()
			return errors.New("seek length not equal file size")
		}
		file = fi
	} else {
		create, err := os.Create(dfn)
		if err != nil {
			log.Println("DownBigFile os.Create err:", err)
			return err
		}
		file = create
	}
	client := &http.Client{Timeout: timeOut}
	request := http.Request{Method: http.MethodGet}
	if size != 0 {
		header := http.Header{}
		header.Set("Range", "bytes="+strconv.FormatInt(size, 10)+"-")
		request.Header = header
	}
	parse, err := url.Parse(uri)
	if err != nil {
		log.Println("DownBigFile url.Parse err:", err)
		return err
	}
	request.URL = parse
	get, err := client.Do(&request)
	if err != nil {
		log.Println("DownBigFile client.Do err:", err)
		return err
	}
	defer func() {
		err := get.Body.Close()
		if err != nil {
			log.Println("DownBigFile get.Body.Close err:", err)
		}
		err = file.Close()
		if err != nil {
			log.Println("DownBigFile file.Close err:", err)
		}
	}()
	if get.ContentLength == 0 {
		log.Println("DownBigFile ", fileName, ", already downloaded")
		return nil
	}
	body := get.Body
	writer := bufio.NewWriter(file)
	bs := make([]byte, 10*1024*1024) //每次读取的最大字节数，不可为0
	for {
		var read int
		read, err = body.Read(bs)
		if err != nil {
			if err != io.EOF {
				log.Println("DownBigFile body.Read err:", err)
			} else {
				err = nil
			}
			break
		}
		_, err = writer.Write(bs[:read])
		if err != nil {
			log.Println("DownBigFile writer.Write err:", err)
			break
		}
	}
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		log.Println("DownBigFile writer.Flush err:", err)
		return err
	}
	log.Println("DownBigFile ", fileName, ", download success")
	return nil
}

/**
 * 下载文件
 * uri: 待下载地址
 * dir: 下载文件的存放目录，相对于项目根目录开始
 * fileName: 下载后的文件名称
 */
func DownFile(uri, dir, fileName string) error {
	log.Printf("DownFile param uri:%v, dir:%v, fileName:%v", uri, dir, fileName)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Println("DownFile os.MkdirAll err:", err.Error())
		return err
	}

	dfn := dir + "/" + fileName
	out, err := os.Create(dfn)
	if err != nil {
		log.Println("DownFile os.Create err:", err)
		return err
	}
	defer out.Close()

	resp, err := http.Get(uri)
	if err != nil {
		log.Println("DownFile http.Get err:", err)
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("DownFile io.Copy err:", err)
		return err
	}
	log.Println("DownFile ", fileName, ", download success")
	return nil
}
