package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	CodeTypeLogo = "logo"
	DIR          = "./tmp"
)

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// 设置最大上传大小  32 MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Errorf("file size over")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Errorf("form file err")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	// 获取其他表单字段
	name, tc, codeType := r.FormValue("name"), r.FormValue("tc"), r.FormValue("codeType")
	log.Infof("upload info name:%s , tc:%v , codeType:%v", name, tc, codeType)

	// 处理文件上传
	dst, err := os.Create(filepath.Join(DIR, handler.Filename))
	if err != nil {
		log.Errorf("create file err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = dst.Close()
	}()

	// 复制文件内容
	if _, err = io.Copy(dst, file); err != nil {
		log.Errorf("copy file err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//生成二维码
	options := make([]Option, 0)
	if codeType == CodeTypeLogo {
		options = append(options, WithLogoFile(fmt.Sprintf("%s/%s", DIR, handler.Filename)))
	} else {
		options = append(options, WithHalftoneSrcFile(fmt.Sprintf("%s/%s", DIR, handler.Filename)))
	}
	options = append(options, WithLogoWidth(BIG))
	qrCode, err := NewQuCodeGen(name, options...).GenQrCode()
	if err != nil {
		log.Errorf("gen qr code err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]interface{}{
		"code": 200,
		"data": qrCode,
	})
	if err != nil {
		log.Errorf("json marshal err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(resp)
	return
}

func runHttp() {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/qrcode/gen", uploadFileHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("."))))
	_ = http.Serve(listen, mux)
}

func main() {
	_ = os.Mkdir(DIR, os.ModePerm)
	runHttp()
}
