package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type SuccessRes struct {
	Name     string
	Tc       int
	Sentence string
}

type Req struct {
	FileAddress string `json:"fileAddress"`
	Name        string `json:"name"`
	Tc          int    `json:"tc"`
}

const (
	CodeTypeLogo = "logo"
	DIR          = "./tmp"
	RootUrl      = "http://yankaka.chat:8081/success"
	StaticPath   = "http://yankaka.chat:8081/static/"
	FontPath     = "./font/hanyiyongzidingshenggaojianti.ttf"
	TemplatePath = "./img/zht.jpeg"
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

	num, err := strconv.Atoi(tc)
	if err != nil {
		log.Errorf("conver int err:%s", err)
		return
	}

	contentUrl := fmt.Sprintf("%s?name=%s&tc=%d", RootUrl, name, num)
	qrCode, err := NewQuCodeGen(contentUrl, options...).GenQrCode()
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

func uploadFileHandlerV2(w http.ResponseWriter, r *http.Request) {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	var req Req
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	log.Infof("request body :%+v", req)

	qrFileName := fmt.Sprintf("%d", time.Now().UnixMilli())
	//生成二维码
	options := make([]Option, 0)
	options = append(options, WithHalftoneSrcFile(fmt.Sprintf("%s/%s.png", "./static", req.FileAddress)))
	options = append(options, WithLogoWidth(BIG))
	options = append(options, WithName(qrFileName))

	contentUrl := fmt.Sprintf("%s?name=%s&tc=%d&img=%s", RootUrl, req.Name, req.Tc, fmt.Sprintf("%s.%s", qrFileName, DefaultFileType))
	qrCode, err := NewQuCodeGen(contentUrl, options...).GenQrCode()
	if err != nil {
		log.Errorf("gen qr code err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("resp qrcode:%s, contentUrl:%s", qrCode, contentUrl)

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

func success(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	query := r.URL.Query()
	// 获取其他表单字段
	name := query.Get("name")
	tc := query.Get("tc")
	sourceImg := query.Get("img")

	log.Printf("upload info name:%s, tc:%v", name, tc)

	img := NewResImg(TemplatePath, []ResImgOption{
		WithFontPath(FontPath),
		WithFontSize(30),
		WithContentImg(ContentImg{
			ImagePath: fmt.Sprintf("%s%s", StaticPath, sourceImg),
			Width:     280,
			Height:    280,
			LineWidth: 2,
			Padding:   10,
			X:         367,
			Y:         410,
		}),
		WithContents([]Content{
			{
				Text: name,
				X:    480,
				Y:    735,
			},
			{
				Text: "祝贺你完成拼图",
				X:    415,
				Y:    780,
			},
			{
				Text: fmt.Sprintf("共计耗时%ss", tc),
				X:    420,
				Y:    825,
			},
			{
				Text:     "公众号:扯编程的淡",
				FontSize: 24,
				X:        690,
				Y:        855,
			},
		}),
	})

	_, fileName, err := img.Gen()
	if err != nil {
		log.Errorf("img gen err:%s", err)
		return
	}
	redirectUrl := fmt.Sprintf("%s%s", StaticPath, fileName)

	log.Infof("gen img fileName:%s , redirectUrl：%s", fileName, redirectUrl)

	http.Redirect(w, r, redirectUrl, http.StatusFound)
}

func runHttp() {
	listen, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/qrcode/gen", uploadFileHandlerV2)
	mux.HandleFunc("/success", success)
	mux.Handle("/static/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	_ = http.Serve(listen, mux)
}

func main() {
	_ = os.Mkdir(DIR, os.ModePerm)
	runHttp()
}
