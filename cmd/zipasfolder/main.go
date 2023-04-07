package main

import (
	"flag"
	"fmt"
	"github.com/je4/utils/v2/pkg/zipasfolder"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

var basedir = flag.String("basedir", "", "The base directory to use for the zip file. (default: current directory)")

func recurseDir(fsys fs.FS, name string) {
	files, err := fs.ReadDir(fsys, name)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("[d] %s/%s\n", name, file.Name())
			recurseDir(fsys, filepath.ToSlash(filepath.Join(name, file.Name())))
		} else {
			if filepath.Ext(file.Name()) == ".xml" {
				fp, err := fsys.Open(filepath.ToSlash(filepath.Join(name, file.Name())))
				if err != nil {
					panic(err)
				}
				io.Copy(os.Stdout, fp)
				fp.Close()
			}
			fmt.Printf("[f] %s/%s\n", name, file.Name())
		}
	}
}

func main() {
	flag.Parse()

	dirFS := os.DirFS(*basedir)
	baseDir, ok := dirFS.(fs.StatFS)
	if !ok {
		panic("cannot cast dirFS to BaseFS")
	}
	newFS := zipasfolder.NewFS(baseDir, 20)
	defer newFS.Close()

	recurseDir(newFS, "")

	time.Sleep(2 * time.Minute)
}
