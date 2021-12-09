package mimetypes

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

//go:embed mimetypes.json
var mimetypes []byte

func NewMimeExt() *MimeExt {
	me := &MimeExt{}
	me.Init()
	return me
}

type MimeExt struct {
	ext2mime map[string][]string
	mime2ext map[string][]string
}

func (me *MimeExt) Add(mime, ext string) {
	if _, ok := me.mime2ext[mime]; !ok {
		me.mime2ext[mime] = []string{}
	}
	me.mime2ext[mime] = append(me.mime2ext[mime], ext)

	if _, ok := me.ext2mime[ext]; !ok {
		me.ext2mime[ext] = []string{}
	}
	me.ext2mime[ext] = append(me.ext2mime[ext], mime)
}

func (me *MimeExt) GetExt(mime string) []string {
	exts, ok := me.mime2ext[mime]
	if !ok {
		return []string{"bin"}
	}
	return exts
}

func (me *MimeExt) GetMime(ext string) []string {
	mimes, ok := me.ext2mime[ext]
	if !ok {
		return []string{"application/octet-stream"}
	}
	return mimes
}

func (me *MimeExt) SPrintMime2Ext() string {
	var result string
	for key, vals := range me.mime2ext {
		result += fmt.Sprintf("\"%s\": {\"%s\"},\n", key, strings.TrimRight(strings.Join(vals, "\", \""), "\""))
	}
	return result
}

func (me *MimeExt) SPrintExt2Mime() string {
	var result string
	for key, vals := range me.ext2mime {
		result += fmt.Sprintf("\"%s\": {\"%s\"},\n", key, strings.TrimRight(strings.Join(vals, "\", \""), "\""))
	}
	return result
}

type mimeEntry struct {
	Source       string   `json:"source"`
	Charset      string   `json:"charset,omitempty"`
	Compressible bool     `json:"compressible,omitempty"`
	Extensions   []string `json:"extensions,omitempty"`
}

func (me *MimeExt) Init() {
	me.mime2ext = map[string][]string{}
	me.ext2mime = map[string][]string{}
	mes := map[string]mimeEntry{}
	if err := json.Unmarshal(mimetypes, &mes); err != nil {
		log.Fatalf("cannot unmarshal mimetypes: %v", err)
	}
	for mt, data := range mes {
		for _, ext := range data.Extensions {
			me.Add(mt, ext)
		}
	}
}
