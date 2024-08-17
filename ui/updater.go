package ui

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"strings"
	"time"

	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/minio/selfupdate"

	"github.com/jezek/xgbutil/xgraphics"
	"github.com/jezek/xgbutil/xwindow"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

var (
	logoSize   int = 256 // Size of logo image
	logoMargin int = 15  // Margin of logo image
)

var (
	win *xwindow.Window  // Overlay progress window
	cv  *xgraphics.Image // Overlay progress canvas
)

func UpdateBinary(ws *desktop.Workspace, asset common.Info, fun func()) error {

	// Show progress window
	showProgress(ws)

	// Download gzip release file
	updateProgress("download release")
	gzipBody, err := downloadRelease(asset.Url)
	if err != nil {
		return err
	}

	// Download and validate checksum
	updateProgress("validate checksum")
	checksums, err := downloadChecksum(asset.Extra.Url)
	if err != nil {
		return err
	}
	err = validateChecksum(gzipBody, checksums, asset.Name)
	if err != nil {
		return err
	}

	// Extract and update binary
	updateProgress("update binary")
	tarReader, err := extractFile(gzipBody, common.Build.Name)
	if err != nil {
		return err
	}
	err = applyUpdate(tarReader, common.Process.Path)
	if err != nil {
		return err
	}

	// Update progress window with restart info
	updateProgress(fmt.Sprintf("start %s v%s", common.Build.Name, common.Source.Releases[0].Name))

	// Close progress window and make success callback
	closeProgress(3000, fun)

	return nil
}

func CheckPermissions(filepath string) (*selfupdate.Options, error) {
	file, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	// Check file update permissions
	options := selfupdate.Options{
		TargetPath: filepath,
		TargetMode: file.Mode(),
	}
	err = options.CheckPermissions()
	if err != nil {
		return nil, err
	}

	return &options, nil
}

func downloadRelease(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, failure(err)
	}
	defer response.Body.Close()

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, failure(err)
	}

	return body, nil
}

func downloadChecksum(url string) (map[string]string, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, failure(err)
	}
	defer response.Body.Close()

	// Parse checksums file
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "  ")
		if len(parts) == 2 {
			checksums[parts[1]] = parts[0]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, failure(err)
	}

	return checksums, nil
}

func validateChecksum(body []byte, checksums map[string]string, filename string) error {

	// Calculate checksum of response body
	hash := sha256.New()
	if _, err := io.Copy(hash, bytes.NewBuffer(body)); err != nil {
		return failure(err)
	}
	bodyChecksum := hex.EncodeToString(hash.Sum(nil))

	// Check if checksum exists for filename
	fileChecksum, exists := checksums[filename]
	if !exists {
		msg := fmt.Sprintf("Checksum for %s not found", filename)
		return failure(errors.New(msg))
	}

	// Check if checksums match
	if bodyChecksum != fileChecksum {
		msg := fmt.Sprintf("Checksum mismatch for %s (%s != %s)", filename, fileChecksum, bodyChecksum)
		return failure(errors.New(msg))
	}

	return nil
}

func extractFile(body []byte, filename string) (*tar.Reader, error) {
	gzipReader, err := gzip.NewReader(bytes.NewBuffer(body))
	if err != nil {
		return nil, failure(err)
	}
	defer gzipReader.Close()

	// Check files in tar archive
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, failure(err)
		}

		// Return searched file
		if header.Name == filename {
			return tarReader, nil
		}
	}

	msg := fmt.Sprintf("File %s not found", filename)
	return nil, failure(errors.New(msg))
}

func applyUpdate(reader io.Reader, filepath string) error {
	options, err := CheckPermissions(filepath)
	if err != nil {
		return failure(err)
	}

	// Apply binary update
	err = selfupdate.Apply(reader, *options)
	if err != nil {
		return failure(err)
	}

	return nil
}

func showProgress(ws *desktop.Workspace) {

	// Calculate window dimensions
	w, h := logoSize+logoMargin*2, logoSize+logoMargin*2

	// Create an empty canvas image
	bg := bgra("gui_background")
	cv = xgraphics.New(store.X, image.Rect(0, 0, w+2*rectMargin, h+fontSize+2*fontMargin+2*rectMargin))
	cv.For(func(x int, y int) xgraphics.BGRA { return bg })

	// Show the canvas graphics
	win = showGraphics(cv, ws, 0.0)
}

func updateProgress(txt string) {
	if win == nil || cv == nil {
		return
	}

	// Calculate window dimensions
	size := cv.Rect.Size()
	x, y, w, h := 0, 0, size.X, size.Y

	// Draw background onto canvas
	color := bgra("gui_client_slave")
	drawImage(cv, &image.Uniform{color}, color, x+rectMargin, y+rectMargin, x+w-rectMargin, y+h-rectMargin)

	// Draw logo onto canvas
	logo, _, _ := image.Decode(bytes.NewBuffer(common.File.Logo))
	drawImage(cv, xgraphics.NewConvert(store.X, logo), color, x+rectMargin+logoMargin, y+rectMargin+logoMargin, x+w-rectMargin, y+h-rectMargin)

	// Draw text onto canvas
	drawText(cv, txt, bgra("gui_text"), cv.Rect.Dx()/2, cv.Rect.Dy()-2*fontMargin-rectMargin-logoMargin/2, fontSize)

	// Update canvas
	cv.XDraw()
	cv.XPaint(win.Id)
}

func closeProgress(after time.Duration, fun func()) {
	time.AfterFunc(after*time.Millisecond, func() {
		if win != nil {
			win.Destroy()
			win = nil
		}
		if fun != nil {
			fun()
		}
	})
}

func failure(err error) error {
	log.Error("Error updating binary: ", err)

	// Show error message
	updateProgress("error updating binary")

	// Close progress window
	closeProgress(3000, nil)

	return err
}
