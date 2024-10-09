package main

import (
	"fmt"
	"testing"
)

func TestResImg(t *testing.T) {
	img := NewResImg("./img/zht.jpeg", []ResImgOption{
		WithFontPath("./font/hanyiyongzidingshenggaojianti.ttf"),
		WithFontSize(30),
		WithContentImg(ContentImg{
			ImagePath: "http://yankaka.chat:8080/static/1726547875181.jpg",
			Width:     280,
			Height:    280,
			LineWidth: 2,
			Padding:   10,
			X:         367,
			Y:         410,
		}),
		WithContents([]Content{
			{
				Text: "张三",
				X:    480,
				Y:    735,
			},
			{
				Text: "祝贺你完成拼图",
				X:    415,
				Y:    780,
			},
			{
				Text: "共计耗时100s",
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

	gen, err := img.Gen()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(gen)
}
