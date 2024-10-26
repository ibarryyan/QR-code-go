package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

type ResImgOption func(*ResImg)

const (
	DefaultFontSize    = 48
	DefaultDPI         = 70
	DefaultDstFilePath = "./static/"
	DefaultFontPath    = "hanyiyongzidingshenggaojianti.ttf"
)

type ResImg struct {
	FontPath    string
	TemplateImg string
	FontSize    int
	DPI         int
	Contents    []Content
	DstFilePath string
	ContentImg  ContentImg
}

type Content struct {
	Text     string
	X, Y     int
	Color    *color.RGBA
	Font     *truetype.Font
	FontSize int
}

type ContentImg struct {
	ImagePath     string
	Width, Height uint
	LineWidth     int
	Padding       int
	X, Y          int
}

func NewResImg(templatePath string, opts []ResImgOption) *ResImg {
	r := &ResImg{
		Contents:    make([]Content, 0),
		FontPath:    DefaultFontPath,
		FontSize:    DefaultFontSize,
		DPI:         DefaultDPI,
		DstFilePath: DefaultDstFilePath,
	}

	if templatePath != "" {
		r.TemplateImg = templatePath
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithFontPath(path string) ResImgOption {
	return func(img *ResImg) {
		img.FontPath = path
	}
}

func WithFontSize(size int) ResImgOption {
	return func(img *ResImg) {
		img.FontSize = size
	}
}

func WithContents(contents []Content) ResImgOption {
	return func(img *ResImg) {
		img.Contents = append(img.Contents, contents...)
	}
}

func WithDPI(dpi int) ResImgOption {
	return func(img *ResImg) {
		img.DPI = dpi
	}
}

func WithContentImg(contentImg ContentImg) ResImgOption {
	return func(img *ResImg) {
		img.ContentImg = contentImg
	}
}

func WithDstPath(path string) ResImgOption {
	return func(img *ResImg) {
		img.DstFilePath = path
	}
}

func (i *ResImg) Gen() (string, string, error) {
	// 根据路径打开模板文件
	templateFile, err := os.Open(i.TemplateImg)
	if err != nil {
		log.Errorf("os open file err:%s", err)
		return "", "", err
	}
	defer func() {
		_ = templateFile.Close()
	}()

	// 解码
	templateFileImage, err := jpeg.Decode(templateFile)
	if err != nil {
		log.Errorf("png decode err:%s", err)
		return "", "", err
	}
	// 新建一张和模板文件一样大小的画布
	newTemplateImage := image.NewRGBA(templateFileImage.Bounds())
	// 将模板图片画到新建的画布上
	draw.Draw(newTemplateImage, templateFileImage.Bounds(), templateFileImage, templateFileImage.Bounds().Min, draw.Over)

	// 加载字体文件  这里我们加载两种字体文件
	font, err := LoadFont(i.FontPath)
	if err != nil {
		log.Errorf("load font err:%s", err)
		return "", "", err
	}

	// 向图片中写入文字
	if err := i.writeWord2Pic(font, newTemplateImage, i.Contents); err != nil {
		log.Errorf("write word err:%s", err)
		return "", "", err
	}

	if err := i.writeImg2Pic(newTemplateImage); err != nil {
		log.Errorf("write image err:%s", err)
		return "", "", err
	}

	fileName := fmt.Sprintf("%d.png", time.Now().Unix())
	filePath := fmt.Sprintf("%s%s", i.DstFilePath, fileName)
	if err = SaveFile(filePath, newTemplateImage); err != nil {
		log.Errorf("save file err:%s", err)
		return "", "", err
	}
	return filePath, fileName, nil
}

func (i *ResImg) writeWord2Pic(font *truetype.Font, newTemplateImage *image.RGBA, contents []Content) error {
	initContentSetting := func(c *freetype.Context) {
		c.SetSrc(image.Black)
		c.SetDPI(float64(i.DPI))
		c.SetFontSize(float64(i.FontSize))
		c.SetFont(font)
	}

	content := freetype.NewContext()
	content.SetClip(newTemplateImage.Bounds())
	content.SetDst(newTemplateImage)
	initContentSetting(content)

	for _, c := range contents {
		if c.FontSize != 0 {
			content.SetFontSize(float64(c.FontSize))
		}
		if c.Font != nil {
			content.SetFont(c.Font)
		}
		if _, err := content.DrawString(c.Text, freetype.Pt(c.X, c.Y)); err != nil {
			log.Errorf("draw string err:%s", err)
			continue
		}
		initContentSetting(content)
	}
	return nil
}

func (i *ResImg) writeImg2Pic(newTemplateImage *image.RGBA) error {
	img := i.ContentImg

	imageData, err := GetDataByUrl(img.ImagePath) // 根据地址获取图片内容
	if err != nil {
		log.Errorf("get data by url err:%s", err)
		return err
	}

	// 图片层
	imageData = resize.Resize(img.Width, img.Height, imageData, resize.Lanczos3)
	ix, iy := imageData.Bounds().Dx(), imageData.Bounds().Dy()

	// 新建一个透明图层
	transparentImg := image.NewRGBA(image.Rect(0, 0, ix+img.Padding, iy+img.Padding))
	tx, ty := transparentImg.Bounds().Dx(), transparentImg.Bounds().Dy()

	// 将缩略图放到透明图层上
	draw.Draw(transparentImg, image.Rect(img.Padding/2, img.Padding/2, tx, ty), imageData, image.Point{}, draw.Over)

	// 图片周围画线
	gc := draw2dimg.NewGraphicContext(transparentImg)
	gc.SetStrokeColor(color.RGBA{R: uint8(36), G: uint8(106), B: uint8(96), A: 0xff})
	gc.SetFillColor(color.RGBA{})
	gc.SetLineWidth(float64(img.LineWidth)) // 线框宽度
	gc.BeginPath()
	gc.MoveTo(0, 0)
	gc.LineTo(float64(tx), 0)
	gc.LineTo(float64(tx), float64(ty))
	gc.LineTo(0, float64(ty))
	gc.LineTo(0, 0)
	gc.Close()
	gc.FillStroke()

	//粘贴缩略图
	draw.Draw(newTemplateImage, transparentImg.Bounds().Add(image.Pt(img.X, img.Y)), transparentImg, transparentImg.Bounds().Min, draw.Over)
	return nil
}
