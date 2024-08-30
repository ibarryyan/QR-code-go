package main

import (
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
