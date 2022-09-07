package image

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	"github.com/y1015860449/gotoolkit/utils"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"os"
)

const (
	DefaultMergeImagePx = 360
)

// LoadUrlImage 丛网页获取图片
func LoadUrlImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	if img, _, err := image.Decode(resp.Body); err != nil {
		return nil, err
	} else {
		return img, nil
	}
}

// LoadLocalImage 本地获取图片
func LoadLocalImage(path string) (image.Image, error) {
	// 获取文件类型
	fImgType, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fImgType.Close()
	buff := make([]byte, 512)
	_, err = fImgType.Read(buff)
	if err != nil {
		return nil, err
	}
	imgType := http.DetectContentType(buff)

	// 读取图片
	fImg, err := os.Open(path)
	defer fImg.Close()
	var img image.Image
	switch imgType {
	case "image/jpeg", "image/jpg":
		img, err = jpeg.Decode(fImg)
		if err != nil {
			return nil, err
		}
	case "image/gif":
		img, err = gif.Decode(fImg)
		if err != nil {
			return nil, err
		}
	case "image/png":
		img, err = png.Decode(fImg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, err
	}
	return img, nil
}

// MergeImageByUrl 拼接图片
func MergeImageByUrl(urls []string, maxPx, spacePx int) (*image.RGBA, error) {
	if len(urls) < 2 {
		return nil, errors.New("param is exception")
	}
	if maxPx <= 0 {
		maxPx = DefaultMergeImagePx
	}
	withPx := (maxPx - spacePx) / 2
	var childRect []image.Rectangle
	// 微信和钉钉的头像处理
	switch len(urls) {
	case 2:
		paddingPx := (maxPx - withPx) / 2
		rect1 := image.Rect(0, paddingPx, withPx, paddingPx+withPx)
		rect2 := image.Rect(rect1.Max.X+spacePx, paddingPx, maxPx, paddingPx+withPx)
		childRect = append(childRect, rect1, rect2)
	case 3:
		paddingPx := (maxPx - withPx) / 2
		rect1 := image.Rect(paddingPx, 0, paddingPx+withPx, withPx)
		rect2 := image.Rect(0, withPx+spacePx, withPx, maxPx)
		rect3 := image.Rect(withPx+spacePx, withPx+spacePx, maxPx, maxPx)
		childRect = append(childRect, rect1, rect2, rect3)
	case 4:
		rect1 := image.Rect(0, 0, withPx, withPx)
		rect2 := image.Rect(withPx+spacePx, 0, maxPx, rect1.Max.Y)
		rect3 := image.Rect(0, withPx+spacePx, withPx, maxPx)
		rect4 := image.Rect(withPx+spacePx, withPx+spacePx, maxPx, maxPx)
		childRect = append(childRect, rect1, rect2, rect3, rect4)
	}

	destImg := image.NewRGBA(image.Rect(0, 0, maxPx, maxPx))
	// 添加白色背景
	//draw.Draw(destImg, destImg.Bounds(), image.White, image.ZP, draw.Src)
	for i, url := range urls {
		img, err := LoadUrlImage(url)
		if err != nil {
			fmt.Errorf("GetUrlImage err(%+v)", err)
			return nil, err
		}
		with := childRect[i].Max.X - childRect[i].Min.X
		height := childRect[i].Max.Y - childRect[i].Min.Y
		tmpImg := ImageShrink(img, with, height)
		draw.Draw(destImg, childRect[i], tmpImg, tmpImg.Bounds().Min, draw.Over)
	}
	return destImg, nil
}

// RandomImageName 生成图片名称
func RandomImageName() string {
	name := utils.GetUUID()
	return fmt.Sprintf("%s.jpg", name)
}

// ImageShrink 图片压缩
func ImageShrink(img image.Image, width, height int) image.Image {
	b := img.Bounds()
	imgWidth := b.Max.X
	imgHeight := b.Max.Y
	ratioW := float64(width) / float64(imgWidth)
	ratioH := float64(height) / float64(imgHeight)
	tmpImg := resize.Resize(uint(math.Ceil(float64(imgWidth)*ratioW)), uint(math.Ceil(float64(imgHeight)*ratioH)), img, resize.Lanczos3)
	return tmpImg
}

// SaveImage 保存图片
func SaveImage(targetPath string, img image.Image) error {
	fSave, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer fSave.Close()
	err = jpeg.Encode(fSave, img, nil)
	if err != nil {
		return err
	}
	return nil
}

// GetThumbData 获取缩略图
func GetThumbData(filename string, width, height int) ([]byte, error) {
	image, err := imaging.Open(filename)
	if err != nil {
		return nil, err
	}
	thumbImage := imaging.Resize(image, width, height, imaging.Lanczos)
	var buf bytes.Buffer
	if err = imaging.Encode(&buf, thumbImage, imaging.JPEG); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GetImageInfo 获取图片信息
func GetImageInfo(filename string) (width int, height int, length int64, isGif bool, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, 0, 0, false, nil
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return 0, 0, 0, false, nil
	}
	img, err := imaging.Open(filename)
	if err != nil {
		return 0, 0, 0, false, err
	}
	_, isGifErr := gif.DecodeConfig(file)
	return img.Bounds().Max.X, img.Bounds().Max.Y, stat.Size(), isGifErr == nil, nil
}
