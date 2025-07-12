package models

type TaskStatus string

const (
	Staged    TaskStatus = "staged"
	Completed TaskStatus = "completed"
)

type Task struct {
	TaskID string         `json:"id"`
	Files  []*FileRequest `json:"files"`
	Status TaskStatus     `json:"status"`
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
