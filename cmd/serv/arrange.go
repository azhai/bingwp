package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	db "github.com/azhai/bingwp/models/default"
	"github.com/azhai/bingwp/utils"

	xutils "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
)

func ArrangeImages() (err error) {
	var rows []*db.WallDaily
	err = db.Query().Desc("id").Limit(92).Find(&rows) //一个季度
	for _, row := range rows {
		img := new(db.WallImage)
		img.Load(xq.WithWhere("id = ?", row.Id*2))
		if img.Id <= 0 || img.ImgSize > 2500 {
			continue
		}
		os.MkdirAll(img.SaveDir, 0o755)
		modes := []string{"_UHD", "_1920x1080", "_1366x768", ""}
		fname := filepath.Join(img.SaveDir, img.FileName)
		i, date := 0, row.BingDate.Format("2006/01/02")
		size, _ := xutils.FileSize(fname)
		for size <= 2500 && i < len(modes) {
			url := fmt.Sprintf("%s/%s/%s%s.jpg", SiteThumbUrl, date, row.BingSku, modes[i])
			err = utils.Download(url, img.FileName, img.SaveDir, 2)
			size, _ = xutils.FileSize(fname)
			i++
		}
		fmt.Println(row.Id, img.FileName)
		if err != nil || size <= 2500 {
			continue
		}
		if err = UpdateImageInfo(img); err != nil {
			panic(err)
		}
		dpi := fmt.Sprintf("%dx%d", img.ImgWidth, img.ImgHeight)
		err = row.Save(map[string]any{"max_dpi": dpi})
	}
	return
}

func FetchImage(row *db.WallDaily) (err error) {
	date := row.BingDate.Format("2006/01/02")
	url := fmt.Sprintf("%s/%s/%s_400x240.jpg", SiteThumbUrl, date, row.BingSku)
	tName := row.BingDate.Format("20060102") + "t.jpg"
	hdName := row.BingDate.Format("20060102") + "hd.jpg"
	saveDir := "/data/bingwp/" + row.BingDate.Format("200601")
	if err = os.MkdirAll(saveDir, 0o755); err == nil {
		err = utils.Download(url, tName, saveDir, 1)
	}
	if err != nil {
		return
	}
	thumb := &db.WallImage{DailyId: row.Id, Id: row.Id*2 - 1, SaveDir: saveDir, FileName: tName}
	image := &db.WallImage{DailyId: row.Id, Id: row.Id * 2, SaveDir: saveDir, FileName: hdName}
	table := (db.WallImage{}).TableName()
	err = db.InsertBatch(table, thumb, image)
	err = UpdateImageInfo(thumb)
	return
}

func UpdateImageInfo(img *db.WallImage) (err error) {
	fname := filepath.Join(img.SaveDir, img.FileName)
	changes := make(map[string]any)
	size, _ := xutils.FileSize(fname)
	if size <= 0 {
		return
	}
	changes["img_size"] = int(size)
	width, weight := utils.GetImageDims(fname)
	changes["img_width"], changes["img_height"] = width, weight
	md5, _ := utils.Md5Sum(fname)
	changes["img_md5"] = md5
	err = img.Save(changes)
	return
}

type FileDict struct {
	NameHash map[string][]string
	PathMap  map[string]string
}

func NewFileDict() *FileDict {
	return &FileDict{
		NameHash: make(map[string][]string),
		PathMap:  make(map[string]string),
	}
}

func (d *FileDict) AddFiles(dir, ext string) error {
	files, err := xutils.FindFiles(dir, ext)
	if err != nil {
		return err
	}
	for fullName := range files {
		baseName := filepath.Base(fullName)
		if idx := strings.LastIndex(baseName, "_"); idx > 0 {
			baseName = baseName[:idx]
		}
		if len(baseName) < 3 {
			continue
		}
		label := baseName[:3]
		d.NameHash[label] = append(d.NameHash[label], baseName)
		d.PathMap[baseName] = fullName
	}
	return nil
}

func (d FileDict) MatchPrefix(fname string) (baseName, fullName string) {
	var lst []string
	label, ok := fname[:3], false
	if lst, ok = d.NameHash[label]; !ok {
		return
	}
	for _, baseName = range lst {
		if strings.HasPrefix(baseName, fname) {
			fullName = d.PathMap[baseName]
			return
		}
	}
	return
}
