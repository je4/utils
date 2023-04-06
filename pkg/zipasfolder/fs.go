package zipasfolder

import (
	"archive/zip"
	"github.com/bluele/gcache"
	"github.com/pkg/errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

func NewFS(baseFS fs.StatFS, cacheSize int) *FS {
	return &FS{
		baseFS: baseFS,
		zipCache: gcache.New(cacheSize).
			LoaderFunc(func(key interface{}) (interface{}, error) {
				zipFilename, ok := key.(string)
				if !ok {
					return nil, errors.Errorf("cannot cast key %v to string", key)
				}
				zipFile, err := baseFS.Open(zipFilename)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot open zip file '%s'", zipFilename)
				}
				stat, err := zipFile.Stat()
				if err != nil {
					return nil, errors.Wrapf(err, "cannot stat zip file '%s'", zipFilename)
				}
				filesize := stat.Size()
				readerAt, ok := zipFile.(io.ReaderAt)
				if !ok {
					zipFile.Close()
					return nil, errors.Errorf("cannot cast file '%s' to io.ReaderAt", zipFilename)
				}
				zipReader, err := zip.NewReader(readerAt, filesize)
				if err != nil {
					zipFile.Close()
					return nil, errors.Wrapf(err, "cannot create zip reader for '%s'", zipFilename)
				}
				zipFS := NewZIPFS(zipReader, zipFile)
				return zipFS, nil
			}).
			EvictedFunc(func(key, value any) {
				zipFS, ok := value.(*ZIPFS)
				if !ok {
					return
				}
				zipFS.Close()
			}).
			LRU().
			Build(),
	}
}

type FS struct {
	baseFS   fs.StatFS
	zipCache gcache.Cache
}

func (fsys *FS) Sub(dir string) (fs.FS, error) {
	//TODO implement me
	panic("implement me")
}

func (fsys *FS) ReadFile(name string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (fsys *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (fsys *FS) Open(name string) (fs.File, error) {
	zipFile, zipPath, handleZip := expandZipFile(name)
	if !handleZip {
		file, err := fsys.baseFS.Open(name)
		if err != nil {
			return file, errors.Wrapf(err, "cannot open file '%s'", name)
		}
		return file, nil
	}

	zipFSCache, err := fsys.zipCache.Get(zipFile)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get zip file '%s'", zipFile)
	}
	zipFS, ok := zipFSCache.(*ZIPFS)
	if !ok {
		return nil, errors.Errorf("cannot cast zip file '%s' to *ZIPFS", zipFile)
	}
	rc, err := zipFS.Open(zipPath)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file '%s' in zip file '%s'", zipPath, zipFile)
	}
	return rc, nil
}

func isZipFile(name string) bool {
	return strings.ToLower(filepath.Ext(name)) == ".zip"
}

func expandZipFile(name string) (string, string, bool) {
	name = filepath.ToSlash(filepath.Clean(name))
	parts := strings.Split(name, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if isZipFile(parts[i]) {
			return strings.Join(parts[:i], "/"), strings.Join(parts[i+1:], "/"), true
		}
	}
	return "", "", false
}

var (
	_ fs.FS         = &FS{}
	_ fs.ReadDirFS  = &FS{}
	_ fs.ReadFileFS = &FS{}
	_ fs.SubFS      = &FS{}
)
