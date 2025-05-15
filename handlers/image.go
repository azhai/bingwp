package handlers

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	db "github.com/azhai/bingwp/models/default"
	"github.com/azhai/gozzo/cryptogy"
	"github.com/azhai/gozzo/filesystem"
	"github.com/azhai/gozzo/transfer"
	xq "github.com/azhai/xgen/xquery"
)

var (
	imageSaveDir    = ""
	DailyIdYear2020 = 4018
	NoPhotoWidth    = 80
	NoPhotoSize     = 1192
	NoPhotoMd5      = "f0f7d2c575a576fcbe5904900906e27a"
)

// SetImageSaveDir 设置图片保存目录
func SetImageSaveDir(dir string) {
	dir, _ = filepath.Abs(dir)
	imageSaveDir = dir
}

func ImagePath(dt time.Time) string {
	date := dt.Format("20060102")
	return fmt.Sprintf("image/%s/%s.jpg", date[:6], date)
}

func ThumbPath(dt time.Time) string {
	date := dt.Format("20060102")
	return fmt.Sprintf("thumb/%s/%s.jpg", date[:4], date)
}

func NewWallImages(id int64, img string) *db.WallImage {
	obj := new(db.WallImage)
	obj.DailyId, obj.FileName = id, img
	return obj
}

func UpdateDailyImages(wp *db.WallDaily) (dims string, err error) {
	thumbFile, imageFile := ThumbPath(wp.BingDate), ImagePath(wp.BingDate)
	if err = FetchImages(wp.BingSku, false, thumbFile, imageFile); err != nil {
		return
	}
	thumb := NewWallImages(wp.Id, thumbFile)
	thumb.Id = thumb.DailyId*2 - 1
	if thumb, err = LoadImageRow(thumb); err == nil {
		_, err = UpdateImageInfo(thumb)
	}
	image := NewWallImages(wp.Id, imageFile)
	image.Id = image.DailyId * 2
	if image, err = LoadImageRow(image); err == nil {
		dims, err = UpdateImageInfo(image)
	}
	return
}

func FetchImages(sku string, force bool, filenames ...string) error {
	spec, down := "", transfer.NewDownloader(imageSaveDir, 1)
	for _, imgFile := range filenames {
		if strings.HasPrefix(imgFile, "thumb") {
			spec = "_400x240"
		} else {
			spec = "_UHD"
		}
		url := fmt.Sprintf("%s%s%s.jpg", BingThumbUrl, sku, spec)
		if _, err := down.Download(url, imgFile, force); err != nil {
			return err
		}
	}
	return nil
}

func LoadImageRow(img *db.WallImage) (*db.WallImage, error) {
	where := xq.WithWhere("daily_id = ? AND file_name = ?", img.DailyId, img.FileName)
	if img.Id > 0 {
		where = xq.WithWhere("id = ?", img.Id)
	}
	if exists, err := img.Load(where); !exists {
		err = img.Save(nil)
		return img, err
	}
	return img, nil
}

func UpdateImageInfo(img *db.WallImage) (dims string, err error) {
	filename := filepath.Join(imageSaveDir, img.FileName)
	fh := filesystem.File(filename)
	if !fh.IsExist() {
		err = fh.Error()
		return
	}
	changes, size := make(map[string]any), int64(0)
	if size = fh.Size(); size <= 0 {
		err = fh.Error()
		return
	}
	changes["img_size"] = int(size)
	width, weight := fh.GetDims()
	if width > 0 && weight > 0 {
		dims = fmt.Sprintf("%dx%d", width, weight)
	}
	changes["img_width"], changes["img_height"] = width, weight
	md5, _ := cryptogy.Md5File(filename)
	changes["img_md5"] = md5
	err = img.Save(changes)
	return
}
