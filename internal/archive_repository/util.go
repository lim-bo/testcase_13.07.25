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
	buf.Reset()
	zipWriter := zip.NewWriter(buf)

	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)
	filesMutex := sync.Mutex{}

	for _, file := range files {
		wg.Add(1)
		go func(f *models.FileRequest) {
			defer wg.Done()
			resp, err := cli.Get(f.Link)
			if err != nil || resp.StatusCode != http.StatusOK {
				errChan <- err
				return
			}
			defer resp.Body.Close()
			header := zip.FileHeader{
				Name:   f.Name + string(f.Ext),
				Method: zip.Store,
			}
			writer, err := zipWriter.CreateHeader(&header)
			if err != nil {
				errChan <- err
				return
			}
			filesMutex.Lock()
			defer filesMutex.Unlock()
			_, err = io.Copy(writer, resp.Body)
			if err != nil {
				errChan <- err
			}
		}(file)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()

	var resultErr error = nil
	for err := range errChan {
		resultErr = errors.Join(resultErr, err)
	}

	if resultErr != nil {
		return nil, resultErr
	}

	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
