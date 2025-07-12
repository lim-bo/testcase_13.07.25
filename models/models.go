package models

type TaskStatus uint8

const (
	Staged TaskStatus = iota
	Completed
)

type Task struct {
	TaskID string
	Files  []*FileRequest
	Status TaskStatus
}

type Extension string

var (
	JPG Extension = ".jpg"
	PDF Extension = ".pdf"
)

type FileRequest struct {
	Name string    `json:"name"`
	Ext  Extension `json:"ext"`
	Link string    `json:"link"`
}
