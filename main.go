package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
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

var strs = []string{
	"恭喜你，你是一个真正的游戏大师！",
	"你已经掌握了游戏的精髓，现在是时候去征服其他挑战了。",
	"你的坚持和努力得到了回报，你应该为自己感到骄傲。",
	"通关只是一个新的开始，前面还有更多的挑战和机遇等着你。",
	"你的游戏技巧已经达到了一个新的高度，继续保持，你会更上一层楼。",
	"游戏通关并不是终点，而是一个新的起点，希望你在游戏中找到更多的乐趣。",
	"你的游戏成就证明了你的实力和毅力，希望你在未来的游戏中继续闪耀。",
}

var db map[string]*SuccessRes

const (
	CodeTypeLogo = "logo"
	DIR          = "./tmp"

	RootUrl = "http://yankaka.chat:8080/success"
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

	contentUrl := fmt.Sprintf("%s?name=%s&tc=%d&str=%s", RootUrl, name, num, getStr())
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

	log.Infof("request body :%+v", req) //request body :{FileAddress:p2 Name:yanmingxin Tc:15828}

	//生成二维码
	options := make([]Option, 0)
	options = append(options, WithHalftoneSrcFile(fmt.Sprintf("%s/%s.png", "./static", req.FileAddress)))
	options = append(options, WithLogoWidth(BIG))

	contentUrl := fmt.Sprintf("%s?name=%s&tc=%d&str=%s", RootUrl, req.Name, req.Tc, getStr())
	qrCode, err := NewQuCodeGen(contentUrl, options...).GenQrCode()
	if err != nil {
		log.Errorf("gen qr code err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof("resp qrcode:%s", qrCode)

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

	log.Printf("upload info name:%s, tc:%v", name, tc)

}

func runHttp() {
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/qrcode/gen", uploadFileHandlerV2)
	mux.HandleFunc("/success", success)
	mux.Handle("/static/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	_ = http.Serve(listen, mux)
}

func getStr() string {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(7)
	return strs[randomNumber]
}

func main() {
	db = make(map[string]*SuccessRes)
	_ = os.Mkdir(DIR, os.ModePerm)
	runHttp()
}
