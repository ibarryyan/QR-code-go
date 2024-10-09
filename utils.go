package main

import (
	"bytes"
	"errors"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetFileType(filename string) FileType {
	ext := filepath.Ext(filename)
	if ext != "" {
		switch strings.TrimPrefix(ext, ".") {
		case "jpg":
			return JPG
		case "png":
			return PNG
		}
	}
	return NOT
}

func GetFileName(filename string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return filename[:len(filename)-len(ext)]
	}
	return filename
}

// 根据地址获取图片内容
func GetDataByUrl(url string) (image.Image, error) {
	var img image.Image
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body := res.Body
	defer func() {
		_ = body.Close()
	}()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(url, ".jpg") && !strings.HasSuffix(url, ".jpeg") && !strings.HasSuffix(url, ".png") {
		return nil, errors.New("image type is not support")
	}

	reader := bytes.NewReader(data)
	if strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".jpeg") {
		if img, err = jpeg.Decode(reader); err != nil {
			if img, err = png.Decode(bytes.NewReader(data)); err != nil {
				return nil, err
			}
		}
	}

	if strings.HasSuffix(url, ".png") {
		if img, err = png.Decode(reader); err != nil {
			return nil, err
		}
	}
	return img, nil
}

func SaveFile(fileName string, pic *image.RGBA) error {
	dstFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()
	if err = png.Encode(dstFile, pic); err != nil {
		return err
	}
	return nil
}

func LoadFont(path string) (*truetype.Font, error) {
	fontBytes, err := os.ReadFile(path) // 读取字体文件
	if err != nil {
		return nil, err
	}
	font, err := freetype.ParseFont(fontBytes) // 解析字体文件
	if err != nil {
		return nil, err
	}
	return font, nil
}
