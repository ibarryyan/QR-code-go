package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"testing"
)

func TestImagesWidth(t *testing.T) {
	width := 10000
	src, err := imaging.Open("my.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	fmt.Println(w, h)

	ww, hh := width/w, width/h
	fmt.Println(ww, hh)

	i := width / 6
	fmt.Println(i)
}

func TestFileType(t *testing.T) {
	fileType := GetFileType("hello.s")
	fmt.Println(fileType)

	fileName := GetFileName("hello.png")
	fmt.Println(fileName)
}
