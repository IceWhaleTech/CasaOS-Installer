package internal

import (
	"io"

	"github.com/hashicorp/go-getter"
)

type readCloser struct {
	rc         io.ReadCloser
	src        string
	downloaded int64
	totalSize  int64
	callback   func(downloaded, totalSize int64)
}

func (r *readCloser) Read(p []byte) (n int, err error) {
	n, err = r.rc.Read(p)

	r.downloaded += int64(n)

	if r.callback != nil {
		r.callback(r.downloaded, r.totalSize)
	}

	return
}

func (r *readCloser) Close() error {
	return r.rc.Close()
}

type tracker struct {
	callback func(downloaded, totalSize int64)
}

func (t *tracker) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	return &readCloser{
		rc:        stream,
		src:       src,
		totalSize: totalSize,
		callback:  t.callback,
	}
}

func NewTracker(callback func(downladed, totalSize int64)) getter.ProgressTracker {
	return &tracker{
		callback: callback,
	}
}
