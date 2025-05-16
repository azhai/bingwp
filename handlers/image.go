package handlers

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/azhai/bingwp/services/database"
	"github.com/azhai/gozzo/cryptogy"
	"github.com/azhai/gozzo/filesystem"
	"github.com/azhai/gozzo/transfer"
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

func GetImageInfo(img *database.WallImage) error {
	filename := filepath.Join(imageSaveDir, img.FileName)
	fh := filesystem.File(filename)
	if !fh.IsExist() {
		return fh.Error()
	}
	if img.ImgSize = fh.Size(); img.ImgSize <= 0 {
		return fh.Error()
	}
	var err error
	img.ImgWidth, img.ImgHeight = fh.GetDims()
	img.ImgMd5, err = cryptogy.Md5File(filename)
	return err
}

// SaveDailyImages 保存每日壁纸图片
func SaveDailyImages(wp *database.WallDaily) (dims string, err error) {
	thFile, imFile := ThumbPath(wp.BingDate), ImagePath(wp.BingDate)
	if err = FetchImages(wp.BingSku, false, thFile, imFile); err != nil {
		return
	}
	thumb := &database.WallImage{Id: wp.Id*2 - 1, DailyId: wp.Id}
	image := &database.WallImage{Id: wp.Id * 2, DailyId: wp.Id}
	thumb.FileName, image.FileName = thFile, imFile
	if err = GetImageInfo(thumb); err == nil {
		if _, err = database.UpsertRow(thumb); err != nil {
			return
		}
	}
	if err = GetImageInfo(image); err == nil {
		if _, err = database.UpsertRow(image); err != nil {
			return
		}
	}
	dims = fmt.Sprintf("%dx%d", image.ImgWidth, image.ImgHeight)
	return
}
