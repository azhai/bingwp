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

func UpdateDailyImages(wp *db.WallDaily) (dims string, err error) {
	thumbFile, imageFile := ThumbPath(wp.BingDate), ImagePath(wp.BingDate)
	if err = FetchImages(wp.BingSku, false, thumbFile, imageFile); err != nil {
		return
	}
	thumb := &db.WallImage{DailyId: wp.Id, FileName: thumbFile}
	thumb.Id = thumb.DailyId*2 - 1
	if thumb, err = LoadImageRow(thumb); err == nil {
		_, err = UpdateImageInfo(thumb)
	}
	image := &db.WallImage{DailyId: wp.Id, FileName: imageFile}
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

func RepairLostImages() (err error) {
	var rows []*db.WallImage
	sql := "daily_id >= ? AND (img_width <= ? OR img_size <= ? AND img_md5 = ?)"
	where := xq.WithWhere(sql, DailyIdYear2020, NoPhotoWidth, NoPhotoSize, NoPhotoMd5)
	err = db.Query(where).Desc("id").Find(&rows)
	for _, row := range rows {
		err = RepairImage(row, true)
	}
	return
}

func RepairDailyImages(date string) (err error) {
	wp, ok := new(db.WallDaily), false
	if ok, err = wp.Load(xq.WithWhere("bing_date = ?", date)); !ok {
		return
	}
	var rows []*db.WallImage
	err = db.Query(xq.WithWhere("daily_id = ?", wp.Id)).Find(&rows)
	for _, row := range rows {
		err = RepairImage(row, false)
	}
	return
}

func RepairImage(img *db.WallImage, force bool) (err error) {
	wp, dims := new(db.WallDaily), ""
	ok, _ := wp.Load(xq.WithWhere("id = ?", img.DailyId))
	if ok && FetchImages(wp.BingSku, force, img.FileName) == nil {
		dims, err = UpdateImageInfo(img)
	}
	if dims != "" && dims != "80x80" && dims != "400x240" {
		err = wp.Save(map[string]any{"max_dpi": dims})
	}
	return
}

func RebuildImageRecords() (err error) {
	dt := time.Now()
	stop := MustParseDate(dateFirst)
	for dt.After(stop) {
		id := GetOffsetDay(dt)
		thumbFile, imageFile := ThumbPath(dt), ImagePath(dt)
		thumb := &db.WallImage{DailyId: id, FileName: thumbFile}
		thumb.Id = thumb.DailyId*2 - 1
		if err = thumb.Save(nil); err == nil {
			_, err = UpdateImageInfo(thumb)
			if err != nil {
				fmt.Println(err)
			}
		}
		image := &db.WallImage{DailyId: id, FileName: imageFile}
		image.Id = image.DailyId * 2
		if err = image.Save(nil); err == nil {
			_, err = UpdateImageInfo(image)
			if err != nil {
				fmt.Println(err)
			}
		}
		dt = dt.Add(-24 * time.Hour)
	}
	return
}
