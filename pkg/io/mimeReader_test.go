package indexer

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"reflect"
	"testing"

	"golang.org/x/exp/rand"
)

func createTestImage(t *testing.T, x int) (*bytes.Buffer, int64) {
	img := image.NewRGBA(image.Rect(0, 0, x, x))
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
	return buf, int64(buf.Len())
}

type bufTest struct {
	create     bool
	dimensions int
	expected   int64
	mime       string
	ffb        []byte
	bytesToUse *bytes.Buffer
}

func iterateTests(yield func(bufTest) bool) {
	tests := []bufTest{
		bufTest{false, 0, 0, "application/octet-stream", []byte{00, 00, 00, 00}, bytes.NewBuffer([]byte{})},
		bufTest{false, 0, 4, "application/octet-stream", []byte{137, 80, 78, 71}, bytes.NewBuffer([]byte{137, 80, 78, 71})},
		bufTest{true, 100, 3104, "image/png", []byte{137, 80, 78, 71}, nil},
		bufTest{true, 10, 130, "application/octet-stream", []byte{137, 80, 78, 71}, nil},
	}
	for _, test := range tests {
		if !yield(test) {
			return
		}
	}
}

func TestMimeReader(t *testing.T) {
	for test := range iterateTests {
		var buf *bytes.Buffer
		var unmodifiedSize int64
		if test.create {
			buf, unmodifiedSize = createTestImage(t, test.dimensions)
		} else {
			buf = test.bytesToUse
			unmodifiedSize = test.expected
		}
		if unmodifiedSize != test.expected {
			t.Fatalf(
				"test image size: '%d' is not as expected: %d",
				unmodifiedSize,
				test.expected,
			)
		}
		imgData := buf.Bytes()
		mr, err := NewMimeReader(bytes.NewReader(imgData))
		if err != nil {
			t.Fatalf("failed to create mime reader: %v", err)
		}
		imgCopy := bytes.NewBuffer(nil)
		if bytesRead, err := io.Copy(imgCopy, mr); err != nil {
			t.Fatalf("failed to copy data: %v", err)
		} else if bytesRead != unmodifiedSize {
			t.Fatalf("failed to copy all data: %d != %d", bytesRead, unmodifiedSize)
		}
		if !bytes.Equal(imgData, imgCopy.Bytes()) {
			t.Fatalf("data differs")
		}
		contentType, err := mr.DetectContentType()
		if err != nil {
			t.Fatalf("failed to detect content type: %v", err)
		}
		if contentType != test.mime {
			t.Fatalf("wrong mime type: %s, expected: %s", contentType, test.mime)
		}
		// Additional belt and braces test to verify state of stream to
		// simulate next read, e.g. via PRONOM/Siegfried.
		ffb := imgCopy.Bytes()[:4]
		if !reflect.DeepEqual(ffb, test.ffb) {
			t.Fatalf("ffb: '%x' not equal to expected: '%x'", ffb, test.ffb)
		}
	}
}
