package file

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Md5FileHex 获取文件的md5值
func Md5FileHex(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	digest := md5.New()
	_, err = io.Copy(digest, f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

type FileDetail struct {
	Name   string
	ExName string
	Size   int64
	Md5    string
}

// GetFileDetail 获取文件详细信息
func GetFileDetail(path string) (*FileDetail, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	detail := &FileDetail{Size: info.Size()}
	fName := info.Name()
	pos := strings.LastIndex(fName, ".")
	if pos != -1 {
		detail.Name = fName[:pos]
		detail.ExName = fName[pos+1:]
	}
	digest := md5.New()
	_, err = io.Copy(digest, f)
	if err != nil {
		return nil, err
	}
	detail.Md5 = fmt.Sprintf("%x", digest.Sum(nil))
	return detail, nil
}

// ReadAtFile 读取指定位置和大小的文件数据
func ReadAtFile(path string, offset, limit int64) ([]byte, error) {
	if offset < 0 || limit <= 0 || len(path) <= 0 {
		return nil, errors.New("param invalid")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if offset >= info.Size() {
		return nil, errors.New("param invalid")
	}

	if offset+limit > info.Size() {
		limit = info.Size() - offset
	}
	data := make([]byte, limit, limit)
	if _, err = f.ReadAt(data, offset); err != nil {
		return nil, err
	}
	return data, nil
}

// WriteAtFile 在文件的指定位置写入数据
func WriteAtFile(path string, offset int64, data []byte) error {
	if offset < 0 || len(path) <= 0 || len(data) <= 0 {
		return errors.New("param invalid")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.WriteAt(data, offset); err != nil {
		return err
	}
	return nil
}
