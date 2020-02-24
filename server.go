package main

/**
文件拆分上传
todo 拆分成三个接口
改为存储分片，下载时合并
1.检查线上是否有(极速上传) 将所有修改时间改为0 才能保证md5一致 修改时间等信息存储到数据库中
2.分片上传
3.文件合并通知
*/
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var fd fileDealer
var rop redisOp

type JsonRes struct {
	Code    int
	Success int
	Msg     string
	Data    interface{}
}

func errorHanler(e error) { // 全局
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")
	//判断请求方式
	if r.Method == "POST" {
		//设置内存大小
		r.ParseMultipartForm(32 << 20) // 双目运算 32*2^20
		//获取上传的第一个文件
		file, _, err := r.FormFile("file")
		fileMd5 := r.Form.Get("fileMd5")
		pieceMd5 := r.Form.Get("pieceMd5") // 分片md5
		idx, _ := strconv.Atoi(r.Form.Get("index"))
		lastModified, _ := strconv.ParseInt(r.Form.Get("lastModified"), 10, 64)
		filename := r.Form.Get("filename")
		filetype := r.Form.Get("type")
		chunks, _ := strconv.Atoi(r.Form.Get("chunks"))
		defer file.Close()
		if err != nil {
			log.Fatal(err)
		}
		// 文件如果存在分片也是没有必要的
		info := rop.getFileinfo(fileMd5)
		if len(info) > 0 {
			js, _ := json.Marshal(JsonRes{0, 1, "文件存在", info})
			w.Write(js)
			return
		}

		pieceSave(file, pieceMd5) // 保存分片
		rop.chunkAdd(idx, fileMd5, pieceMd5)
		if chunks == rop.chunkIsFull(fileMd5) && rop.isMerging(fileMd5) == 0 { // 上传完并且没有正在合并的则可以执行合并
			rop.merging(fileMd5) // 合并先加锁
			all := rop.getMem(fileMd5)
			_, md5Name, filepath := fd.mergeFile(filename, all, chunks, lastModified)
			rop.fileInfo(fileMd5, filename, filepath, filetype) // 保存文件信息
			rop.delMerging(fileMd5)                             // 文件信息存储到map中后就可以解除锁了
			fmt.Println(md5Name)
			if md5Name != fileMd5 { // 核对前后端文件
				js, _ := json.Marshal(JsonRes{400, 0, "文件不一致", nil})
				w.Write(js)
				return
			}
			rop.clearSet(fileMd5) // 满了则清除集合 合并文件
			js, _ := json.Marshal(JsonRes{0, 1, "文件合并完成", map[string]string{"filepath": filepath}})
			w.Write(js)
			return
		}
		js, _ := json.Marshal(JsonRes{0, 1, "分片上传完成", idx})
		w.Write(js)

	} else {
		n, _ := fmt.Fprintf(w, "错误")
		if n > 0 {
			fmt.Println("错误")
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadFile("html/index.html")
	_, e := fmt.Fprint(w, string(body))
	if e != nil {
		fmt.Println(e)
	}
}

func fileExist(w http.ResponseWriter, r *http.Request) { // 顺便实现断点续传
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")
	form := r.URL.Query()
	fileMd5 := form["fileMd5"][0]
	// 查文件是否存在
	f := rop.getFileinfo(fileMd5)
	if len(f) > 0 { // 文件不存在直接返回
		js, _ := json.Marshal(JsonRes{0, 1, "文件存在", f})
		w.Write(js)
		return
	}

	// 文件不存在就检测是否有在上传当中的 断点续传
	mem := rop.getMemWithScore(fileMd5)
	if len(mem) > 0 {
		js, _ := json.Marshal(JsonRes{301, 1, "断点续传", mem})
		w.Write(js)
		return
	}

	js, _ := json.Marshal(JsonRes{404, 1, "文件不存在", nil})
	w.Write(js)

}

func main() {
	rop.init()
	http.HandleFunc("/find", fileExist) //文件检查接口
	http.HandleFunc("/", upload)        //
	http.HandleFunc("/page", index)
	e := http.ListenAndServe(":8080", nil)
	if e != nil {
		fmt.Println(e)
	}
}
