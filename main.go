package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

const DIR = "./tmp"

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// 设置最大上传大小（可选）
	err := r.ParseMultipartForm(32 << 20) // 32 MB
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	// 获取其他表单字段
	name := r.FormValue("name")
	//size := r.FormValue("size")
	codeType := r.FormValue("codeType")

	// 处理文件上传
	dst, err := os.Create(filepath.Join(DIR, handler.Filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = dst.Close()
	}()

	// 复制文件内容
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//生成二维码
	if codeType == "logo" {
		gen := NewQuCodeGen(name, WithLogoFile(fmt.Sprintf("%s/%s", DIR, handler.Filename)), WithLogoWidth(BIG))
		if err := gen.GenQrCode(); err != nil {
			fmt.Println(err)
		}
	} else {
		gen := NewQuCodeGen(name, WithHalftoneSrcFile(fmt.Sprintf("%s/%s", DIR, handler.Filename)))
		if err := gen.GenQrCode(); err != nil {
			fmt.Println(err)
		}
	}

	// 响应客户端
	fmt.Fprintf(w, "File uploaded successfully: %s\n", handler.Filename)
}

func runHttp() {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/gen", uploadFileHandler)
	_ = http.Serve(listen, mux)
}

func main() {
	_ = os.Mkdir(DIR, os.ModePerm)
	runHttp()
}
