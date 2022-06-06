package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Code-Hex/pget"
	xutils "github.com/azhai/xgen/utils"
)

// Md5Sum 计算文件的md5哈希值
func Md5Sum(fname string) (string, error) {
	fp, size, err := xutils.OpenFile(fname, true, false)
	if err != nil || size <= 0 {
		return "", err
	}
	defer fp.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, fp); err != nil {
		return "", err
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	return sum, nil
}

// Md5Image 计算图片的md5，可能需要先下载
func Md5Image(imgUrl, saveDir string) (string, error) {
	if !strings.HasPrefix(strings.ToLower(imgUrl), "http") {
		return Md5Sum(imgUrl)
	}
	saveName := filepath.Base(imgUrl)
	saveName = SimplifyFilename(saveName)
	savePath := filepath.Join(saveDir, saveName)
	if size, _ := xutils.FileSize(savePath); size > 0 {
		return Md5Sum(savePath)
	}
	err := Download(imgUrl, saveName, saveDir, 1)
	if err != nil {
		return "", err
	}
	return Md5Sum(savePath)
}

// SimplifyFilename 简化图片文件扩展名
func SimplifyFilename(fname string) string {
	extname := filepath.Ext(fname)
	if idx := strings.IndexRune(extname, '_'); idx > 0 {
		size := len(fname) - len(extname) + idx
		fname = fname[:size]
	}
	return fname
}

// Download 下载文件到指定位置
func Download(fileUrl, fileName, fileDir string, num int) error {
	if num <= 0 {
		num = runtime.NumCPU()
	}
	conf := &pget.DownloadConfig{
		Client:   http.DefaultClient,
		Dirname:  fileDir,
		Filename: fileName,
		Procs:    num,
		URLs:     []string{fileUrl},
	}
	resp, err := conf.Client.Head(fileUrl)
	if err != nil {
		return err
	}
	conf.ContentLength = resp.ContentLength
	return pget.Download(context.Background(), conf)
}
