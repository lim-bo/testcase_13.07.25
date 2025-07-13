package archives

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync"
	"testcase/models"
	"time"

	"github.com/google/uuid"
)

func generateID() string {
	return uuid.New().String()[:8]
}

func getRawArchive(files []*models.FileRequest) ([]byte, error) {
	cli := http.Client{
		Timeout: time.Second * 30,
	}
	buf := bytes.NewBuffer(nil)
	zipWriter := zip.NewWriter(buf)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		allErrors []error
	)

	for _, file := range files {
		wg.Add(1)
		go func(f *models.FileRequest) {
			defer wg.Done()
			resp, err := cli.Get(f.Link)
			if err != nil || resp.StatusCode != http.StatusOK {
				allErrors = append(allErrors, err)
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			header := zip.FileHeader{
				Name:   f.Name + string(f.Ext),
				Method: zip.Deflate,
			}
			writer, err := zipWriter.CreateHeader(&header)
			if err != nil {
				allErrors = append(allErrors, err)
				return
			}
			mu.Unlock()

			mu.Lock()
			_, err = io.Copy(writer, resp.Body)
			if err != nil {
				allErrors = append(allErrors, err)
			}
			mu.Unlock()
		}(file)
	}
	wg.Wait()

	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}

	if len(allErrors) > 0 {
		return nil, errors.Join(allErrors...)
	}

	return buf.Bytes(), nil
}
