package qqwry

// origin: https://github.com/freshcn/qqwry/blob/master/download.go

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
)

const (
	// KeyURL is url to download key
	KeyURL = "http://update.cz88.net/ip/copywrite.rar"

	// DbURL is url to download db
	DbURL = "http://update.cz88.net/ip/qqwry.rar"
)

// @ref https://zhangzifan.com/update-qqwry-dat.html

// GetDownloadKey reads newest key from key url
func GetDownloadKey() (uint32, error) {
	resp, err := http.Get(KeyURL)
	if err != nil {
		return 0, fmt.Errorf("could open key url: %w", err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 24)
	if _, err := io.ReadAtLeast(resp.Body, buf, len(buf)); err != nil {
		return 0, fmt.Errorf("could not read from key url: %w", err)
	}

	// @see https://stackoverflow.com/questions/34078427/how-to-read-packed-binary-data-in-go
	return binary.LittleEndian.Uint32(buf[20:]), nil
}

// Downloader is a reader with decryption
type downloader struct {
	r    io.Reader
	resp *http.Response
}

// Read from decrypted header and http body
func (dr *downloader) Read(buf []byte) (int, error) {
	return dr.r.Read(buf)
}

// Close http body
func (dr *downloader) Close() error {
	return dr.resp.Body.Close()
}

// Download create a io.ReadCloser from db url with provided key
func Download(key uint32) (int64, io.ReadCloser, error) {
	resp, err := http.Get(DbURL)
	if err != nil {
		return 0, nil, fmt.Errorf("could open db url: %w", err)
	}

	// decrypt first 512 bytes by key
	buf := make([]byte, 512)

	if _, err := io.ReadAtLeast(resp.Body, buf, len(buf)); err != nil {
		return 0, nil, fmt.Errorf("could not read from key url: %w", err)
	}

	for i, b := range buf {
		key = ((key * 0x805) + 1) & 0xff
		buf[i] = byte(uint32(b) ^ key)
	}

	dr := &downloader{
		// concat decrypted 512 bytes and body
		r: io.MultiReader(bytes.NewReader(buf), resp.Body),

		resp: resp,
	}
	return resp.ContentLength, dr, nil
}
