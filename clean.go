package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func CleanTask() {
	c := cron.New()
	_, err := c.AddFunc("* * * * *", func() {
		if err := CleanTmpFile(GetGlobalConfig().TmpPath); err != nil {

		}
	})
	if err != nil {
		fmt.Println("Error adding task:", err)
		return
	}
	c.Start()
}

func CleanTmpFile(dir string) error {
	log.Infof("clean path:%s", dir)

	// 打开目录
	dir = fmt.Sprintf(".%s", dir)
	d, err := os.Open(dir)
	if err != nil {
		log.Errorf("open dir err:%s", err)
		return err
	}
	defer func() {
		_ = d.Close()
	}()

	// 读取目录中的文件和子目录
	entries, err := d.Readdir(-1)
	if err != nil {
		log.Errorf("read dir err:%s", err)
		return err
	}

	for _, entry := range entries {
		// 构建完整路径
		path := filepath.Join(dir, entry.Name())
		// 如果是文件，则删除
		if !entry.IsDir() {
			if err := os.Remove(path); err != nil {
				log.Errorf("remove file err:%s", err)
				return err
			}
			log.Infof("delete file:%s", path)
		} else {
			log.Infof("skip dir: %s", path)
		}
	}
	return nil
}
