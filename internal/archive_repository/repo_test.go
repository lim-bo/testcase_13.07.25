package archives_test

import (
	"os"
	archives "testcase/internal/archive_repository"
	"testcase/models"
	"testing"
)

// Not a production test
func TestFullCycle(t *testing.T) {
	mng := archives.New(3, 3)
	id, err := mng.CreateTask()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("taskID: ", id)
	status, err := mng.GetTaskStatus(id)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("task status: ", status)
	files := []*models.FileRequest{
		{
			Name: "pdf_sample",
			Ext:  models.PDF,
			Link: "https://www.adobe.com/support/products/enterprise/knowledgecenter/media/c4611_sample_explain.pdf",
		},
		{
			Name: "jpg_sample",
			Ext:  models.JPG,
			Link: "https://i.pinimg.com/736x/a3/73/cb/a373cb4e2a6305b3eaa605d92db09732.jpg",
		},
		{
			Name: "another_jpg_sample",
			Ext:  models.JPG,
			Link: "https://i.pinimg.com/736x/60/7c/b4/607cb4dc3592a40e67fd9c029c36f264.jpg",
		},
	}
	for _, f := range files {
		err = mng.AddFile(id, f)
		if err != nil {
			t.Fatal(err)
		}
	}

	status, err = mng.GetTaskStatus(id)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("task status after adding files: ", status)
	data, err := mng.GetArchive(id)
	err = os.WriteFile("test.zip", data, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
}
