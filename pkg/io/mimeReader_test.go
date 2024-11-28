package indexer

import (
	"bytes"
	"golang.org/x/exp/rand"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"
)

func TestMimeReader(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			var col color.Color
			if rand.Int()%2 == 0 {
				col = image.White
			} else {
				col = image.Black
			}
			img.Set(x, y, col)
		}
	}
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}
	data := buf.Bytes()
	mr, err := NewMimeReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to create mime reader: %v", err)
	}
	data2 := bytes.NewBuffer(nil)
	if n, err := io.Copy(data2, mr); err != nil {
		t.Fatalf("failed to copy data: %v", err)
	} else if n != int64(len(data)) {
		t.Fatalf("failed to copy all data: %d != %d", n, len(data))
	}
	if !bytes.Equal(data, data2.Bytes()) {
		t.Fatalf("data differs")
	}
	contentType, err := mr.DetectContentType()
	if err != nil {
		t.Fatalf("failed to detect content type: %v", err)
	}
	if contentType != "image/png" {
		t.Fatalf("wrong mime type: %s", contentType)
	}
}
