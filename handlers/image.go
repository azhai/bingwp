package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/azhai/bingwp/geohash"
	db "github.com/azhai/bingwp/models/default"
	"github.com/azhai/bingwp/utils"

	xutils "github.com/azhai/xgen/utils"
	xq "github.com/azhai/xgen/xquery"
)

func FetchImage(row *db.WallDaily) (dims string, err error) {
	tName := row.BingDate.Format("20060102") + "t.jpg"
	hdName := row.BingDate.Format("20060102") + "hd.jpg"
	saveDir := "/data/bingwp/" + row.BingDate.Format("200601")
	url := fmt.Sprintf("%s%s_UHD.jpg", BingThumbUrl, row.BingSku)
	if err = os.MkdirAll(saveDir, 0o755); err == nil {
		fname := filepath.Join(saveDir, hdName)
		if size, _ := xutils.FileSize(fname); size <= 0 {
			err = utils.Download(url, hdName, saveDir, 0)
		}
		url = strings.Replace(url, "_UHD", "_400x240", 1)
		fname = filepath.Join(saveDir, tName)
		if size, _ := xutils.FileSize(fname); size <= 0 {
			err = utils.Download(url, tName, saveDir, 1)
		}
	}
	if err != nil {
		return
	}
	thumb := &db.WallImage{DailyId: row.Id, Id: row.Id*2 - 1, SaveDir: saveDir, FileName: tName}
	image := &db.WallImage{DailyId: row.Id, Id: row.Id * 2, SaveDir: saveDir, FileName: hdName}
	table := (db.WallImage{}).TableName()
	err = db.InsertBatch(table, thumb, image)
	_, err = UpdateImageInfo(thumb)
	dims, err = UpdateImageInfo(image)
	return
}

func UpdateImageInfo(img *db.WallImage) (dims string, err error) {
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
	if width > 0 && weight > 0 {
		dims = fmt.Sprintf("%dx%d", width, weight)
	}
	return
}

func RepairImage() (err error) {
	where := xq.WithWhere("max_dpi = '' OR max_dpi = ?", "400x240")
	var rows []*db.WallDaily
	err = db.Query(where).Desc("id").Find(&rows)
	for _, wp := range rows {
		img, dims := new(db.WallImage), ""
		img.Load(xq.WithWhere("id = ?", wp.Id*2))
		dims, err = UpdateImageInfo(img)
		if dims != "" {
			wp.Save(map[string]any{"max_dpi": dims})
		}
	}
	return
}

func UpdateGeo() {
	coord := geohash.NewCoordinate(0)
	var rows []*db.WallLocation
	db.Query().Where("geohash = '' AND latitude <> 0").Find(&rows)
	for _, row := range rows {
		hash := coord.Encode(row.Latitude, row.Longitude)
		if hash != "" {
			row.Save(map[string]any{"geohash": hash})
		}
	}
}
