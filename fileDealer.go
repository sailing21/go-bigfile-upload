package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"time"
)

const bufferSize = 2097152

type fileDealer struct {
	bufferSize int
}

/**
偏存储
*/
func pieceSave(file multipart.File, pieceMd5 string) int64 {
	//创建上传目录
	path := "./upload/pieces/"
	os.Mkdir(path, os.ModePerm)
	//创建上传文件
	filePath := fmt.Sprintf("%s%s", path, pieceMd5)
	f, e := os.Stat(filePath)
	if os.IsExist(e) {
		fmt.Println("片段存在")
		return f.Size()
	}

	cur, err := os.Create(filePath)
	defer cur.Close()
	if err != nil {
		log.Fatal(err)
	}
	//把上传文件数据拷贝到我们新建的文件
	r, _ := io.Copy(cur, file)
	return r // 大小
}

/**
文件合并
*/
func (fd *fileDealer) mergeFile(filename string, chunks []string, cnt int, lastModified int64) (string, string, string) {
	p := "./upload/" + filename
	os.Remove(p)
	fmt.Println(p)
	//创建文件
	f, _ := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	itemPath := ""
	for i := 0; i < cnt; i++ {
		itemPath = "./upload/pieces/" + chunks[i]
		contents, _ := ioutil.ReadFile(itemPath)
		f.Write(contents) //写入文件
	}
	f.Close()
	os.Chtimes(p, time.Unix(0, lastModified*int64(time.Millisecond)), time.Unix(0, lastModified*int64(time.Millisecond)))
	md5Res, _ := fd.MD5sum(p)
	return p, md5Res, p
}

// MD5sum returns MD5 checksum of filename
func (fd *fileDealer) MD5sum(filename string) (string, error) {
	fd.bufferSize = bufferSize
	info, err := os.Stat(filename)
	if err != nil {
		return "", err
	} else if info.IsDir() {
		return "", nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	for buf, reader := make([]byte, fd.bufferSize), bufio.NewReader(file); ; {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		hash.Write(buf[:n])
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum, nil
}
