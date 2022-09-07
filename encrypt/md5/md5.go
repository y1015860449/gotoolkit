package md5

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func StringMd5(txt string) (string, error) {
	m := md5.New()
	_, err := io.WriteString(m, txt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(m.Sum(nil)), nil
}

func FileMd5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		fmt.Println("Copy", err)
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func ByteMd5(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

func ByteMd5Hex(data []byte) string {
	return hex.EncodeToString(ByteMd5(data))
}
