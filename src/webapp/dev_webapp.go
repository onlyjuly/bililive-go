//go:build dev

package webapp

import (
	"net/http"
	"os"
)

func FS() (http.FileSystem, error) {
	buildPath := "src/webapp/build"
	return http.FS(os.DirFS(buildPath)), nil
}
