package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/labstack/echo/v4"
)

type Client struct { // Our example struct, you can use "-" to ignore a field
	Id      string `csv:"client_id"`
	Name    string `csv:"client_name"`
	Age     string `csv:"client_age"`
	NotUsed string `csv:"-"`
}

const (
	fileName = "clients.csv"
)

func main() {
	if err := local(); err != nil {
		panic(err)
	}

	if err := bin([]Client{
		{Id: "1", Name: "hoge", Age: "12"},
		{Id: "2", Name: "\"f,uga\"", Age: "22"},
		{Id: "3", Name: "foo", Age: "32"},
		{Id: "4", Name: "bar", Age: "42"},
	}); err != nil {
		panic(err)
	}
}

func bin(list []Client) error {
	lines := []string{"client_id,client_name,client_age"}
	for _, v := range list {
		lines = append(lines, fmt.Sprintf(
			"%s,%s,%s",
			v.Id,
			v.Name,
			v.Age,
		))
	}
	body := []byte(strings.Join(lines, "\n"))

	file, err := createFile(fileName, []byte(body))
	if err != nil {
		return err
	}

	clients := []*Client{}

	clientsFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("miss open file. file_name: %s", fileName)
	}
	defer clientsFile.Close()

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		return gocsv.LazyCSVReader(in)
	})

	if err := gocsv.Unmarshal(clientsFile, &clients); err != nil {
		return fmt.Errorf("miss unmarshal file. err: %w", err)
	}

	fmt.Println("output bin file")
	for _, client := range clients {
		fmt.Printf("val: %+v\n", client)
	}
	return nil
}

func local() error {
	clientsFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("miss open file. file_name: %s", fileName)
	}
	defer clientsFile.Close()

	clients := []*Client{}

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		return gocsv.LazyCSVReader(in)
	})

	if err := gocsv.UnmarshalFile(clientsFile, &clients); err != nil {
		return fmt.Errorf("miss unmarshal file. err: %w", err)
	}

	fmt.Println("output local file")
	for _, client := range clients {
		fmt.Printf("val: %+v\n", client)
	}
	return nil
}

func createFile(name string, body []byte) (*multipart.FileHeader, error) {
	buf := new(bytes.Buffer)
	mr := multipart.NewWriter(buf)
	w, err := mr.CreateFormFile("file", name)
	if err != nil {
		return nil, fmt.Errorf("miss create form file. err: %w", err)
	}
	w.Write(body)
	mr.Close()
	req := httptest.NewRequest(echo.POST, "/", buf)
	req.Header.Set(echo.HeaderContentType, mr.FormDataContentType())
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	file, err := c.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("miss convert form file. err: %w", err)
	}
	return file, nil
}
