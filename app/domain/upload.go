package domain

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/kataras/iris"

	"github.com/nfnt/resize"
)

const uploadsDir = "./public/uploads/"

var Files = ScanUploads(uploadsDir)

type UploadedFile struct {
	// {name: "", size: } are the dropzone's only requirements.
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type UploadedFiles struct {
	dir   string
	Items []UploadedFile
	mu    sync.RWMutex // slices are safe but RWMutex is a good practise for you.
}

//RefreshUploadFolder refreshes UploadedFiles struct in case someone manualluy delets the uploaded images from the server
func RefreshUploadFolder() *UploadedFiles {
	return ScanUploads(uploadsDir)
}

func Upload(file multipart.File, info *multipart.FileHeader, ctx iris.Context) {
	fname := info.Filename

	out, err := os.OpenFile(uploadsDir+fname,
		os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.Application().Logger().Warnf("Error while preparing the new file: %v", err.Error())
		return
	}
	defer out.Close()

	io.Copy(out, file)

	// optionally, add that file to the list in order to be visible when refresh.
	uploadedFile := Files.add(fname, info.Size)
	go Files.createThumbnail(uploadedFile)
}

func Remove(name string) bool {
	return Files.Delete(name)
}

func ScanUploads(dir string) *UploadedFiles {
	f := new(UploadedFiles)

	lindex := dir[len(dir)-1]
	if lindex != os.PathSeparator && lindex != '/' {
		dir += string(os.PathSeparator)
	}

	// create directories if necessary
	// and if, then return empty uploaded files; skipping the scan.
	if err := os.MkdirAll(dir, os.FileMode(0666)); err != nil {
		return f
	}

	// otherwise scan the given "dir" for files.
	f.scan(dir)
	fmt.Println(reflect.TypeOf(f))
	return f
}

// add the file's Name and Size to the uploadedFiles memory list
func (f *UploadedFiles) add(name string, size int64) UploadedFile {
	f.mu.Lock()
	uf := UploadedFile{
		Name: name,
		Size: size,
	}
	f.Items = append(f.Items, uf)
	f.mu.Unlock()

	return uf
}

// remove, removes the file from the slice
func (f *UploadedFiles) Delete(name string) bool {
	fmt.Println(f.Items)
	f.mu.Lock()
	for i, item := range f.Items {
		if item.Name == name {
			f.Items = append(f.Items[:i], f.Items[i+1:]...)
			os.Remove(uploadsDir + name)
			os.Remove(uploadsDir + "thumbnail_" + name)
			f.mu.Unlock()
			return true
		}
	}
	f.mu.Unlock()
	return false
}

func (f *UploadedFiles) scan(dir string) {
	f.dir = dir
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		// if it's directory or a thumbnail we saved earlier, skip it.
		if info.IsDir() || strings.HasPrefix(info.Name(), "thumbnail_") {
			return nil
		}

		f.add(info.Name(), info.Size())
		return nil
	})
}

// create thumbnail 100x100
// and save that to the ./public/uploads/thumbnail_$FILENAME
func (f *UploadedFiles) createThumbnail(uf UploadedFile) {
	file, err := os.Open(path.Join(f.dir, uf.Name))
	if err != nil {
		return
	}
	defer file.Close()

	name := strings.ToLower(uf.Name)

	out, err := os.OpenFile(f.dir+"thumbnail_"+uf.Name,
		os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer out.Close()

	if strings.HasSuffix(name, ".jpg") {
		// decode jpeg into image.Image
		img, err := jpeg.Decode(file)
		if err != nil {
			return
		}

		// write new image to file
		resized := resize.Thumbnail(180, 180, img, resize.Lanczos3)
		jpeg.Encode(out, resized,
			&jpeg.Options{Quality: jpeg.DefaultQuality})

	} else if strings.HasSuffix(name, ".png") {
		img, err := png.Decode(file)
		if err != nil {
			return
		}

		// write new image to file
		resized := resize.Thumbnail(180, 180, img, resize.Lanczos3) // slower but better res
		png.Encode(out, resized)
	}
	// and so on... you got the point, this code can be simplify, as a practise.
}
