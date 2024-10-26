package main

import (
	"fmt"
	"image"
	"time"

	"github.com/disintegration/imaging"
	log "github.com/sirupsen/logrus"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

type LogoWidth int

type FileType string

type OutputSize int

type Option func(*QrCodeGen)

const (
	MINI   LogoWidth = 4
	MEDIUM LogoWidth = 3
	BIG    LogoWidth = 2

	NOT FileType = ""
	JPG FileType = "jpg"
	PNG FileType = "png"

	OutputMini   OutputSize = 200
	OutputMedium OutputSize = 500
	OutputBig    OutputSize = 1000

	DefaultLogoWidth  = MEDIUM
	DefaultOutputSize = OutputMedium
	DefaultFileType   = JPG
)

type QrCodeGen struct {
	Name            string
	Content         string
	LogoFile        string
	LogoWidth       LogoWidth
	HalftoneSrcFile string
	Width           OutputSize
	OutputFileType  FileType
	Path            string
}

func NewQuCodeGen(content string, opts ...Option) *QrCodeGen {
	gen := &QrCodeGen{
		Content:        content,
		Width:          DefaultOutputSize,
		OutputFileType: DefaultFileType,
		LogoWidth:      DefaultLogoWidth,
	}
	for _, opt := range opts {
		opt(gen)
	}
	return gen
}

func WithLogoFile(fileName string) Option {
	return func(c *QrCodeGen) {
		c.LogoFile = fileName
	}
}

func WithLogoWidth(width LogoWidth) Option {
	return func(c *QrCodeGen) {
		c.LogoWidth = width
	}
}

func WithHalftoneSrcFile(fileName string) Option {
	return func(c *QrCodeGen) {
		c.HalftoneSrcFile = fileName
	}
}

func WithName(name string) Option {
	return func(c *QrCodeGen) {
		c.Name = name
	}
}

func WithOutputFileType(fileType FileType) Option {
	return func(c *QrCodeGen) {
		c.OutputFileType = fileType
	}
}

func WithOutputFileSize(size OutputSize) Option {
	return func(c *QrCodeGen) {
		c.Width = size
	}
}

func WithPath(path string) Option {
	return func(c *QrCodeGen) {
		c.Path = path
	}
}

func (g *QrCodeGen) GenQrCode() (string, error) {
	// 确认文件名称
	qrFileName := fmt.Sprintf("%d.%s", time.Now().UnixMilli(), g.OutputFileType)
	if g.Name != "" {
		qrFileName = fmt.Sprintf("%s.%s", g.Name, g.OutputFileType)
	}

	// 内容
	qrc, err := qrcode.NewWith(g.Content,
		qrcode.WithEncodingMode(qrcode.EncModeByte),
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium),
	)
	if err != nil {
		log.Errorf("qrcode.NewWith error: %v", err)
		return "", err
	}

	// 基本内容
	imageOptions := make([]standard.ImageOption, 0)
	imageOptions = append(imageOptions, standard.WithQRWidth(uint8(g.Width/10)))

	if g.LogoFile != "" {
		var resizeImg *image.NRGBA
		logoSrc, err := imaging.Open(g.LogoFile)
		if err != nil {
			log.Errorf("imaging.Open error: %v", err)
			return "", err
		}
		logoWidth, logoHeight := logoSrc.Bounds().Dx(), logoSrc.Bounds().Dy()

		log.Infof("logofile size width: %d ,height: %d", logoWidth, logoHeight)

		if g.LogoWidth > 0 {
			switch g.LogoWidth {
			case MINI:
				resizeImg = imaging.Resize(logoSrc, int(g.Width)/int(MINI), int(g.Width)/int(MINI), imaging.Lanczos)
			case MEDIUM:
				resizeImg = imaging.Resize(logoSrc, int(g.Width)/int(MEDIUM), int(g.Width)/int(MEDIUM), imaging.Lanczos)
			case BIG:
				resizeImg = imaging.Resize(logoSrc, int(g.Width)/int(BIG), int(g.Width)/int(BIG), imaging.Lanczos)
			}
		} else {
			resizeImg = imaging.Resize(logoSrc, int(g.Width)/int(MEDIUM), int(g.Width)/int(MEDIUM), imaging.Lanczos)
		}

		g.LogoFile = fmt.Sprintf("%s_tmp.%s", GetFileName(g.LogoFile), JPG)
		if err = imaging.Save(resizeImg, g.LogoFile); err != nil {
			log.Errorf("imaging.Save: %v", err)
			return "", err
		}
		imageOptions = append(imageOptions, standard.WithLogoImageFileJPEG(g.LogoFile))
		imageOptions = append(imageOptions, standard.WithLogoSizeMultiplier(int(g.LogoWidth)))
	}

	// Halftone
	if g.HalftoneSrcFile != "" {
		imageOptions = append(imageOptions, []standard.ImageOption{
			standard.WithHalftone(g.HalftoneSrcFile),
			standard.WithBgTransparent(),
		}...)
	}

	w, err := standard.New(fmt.Sprintf("%s/%s", fmt.Sprintf(".%s", g.Path), qrFileName), imageOptions...)
	if err != nil {
		log.Errorf("qrcode.NewWith error: %v", err)
		return "", err
	}
	if err = qrc.Save(w); err != nil {
		log.Errorf("qrc.Save: %v", err)
		return "", err
	}
	return qrFileName, nil
}
