package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

const pathToPdfs = "store/pdfs"

func getImg(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Print(err.Error())
		return "", err
	}
	defer response.Body.Close()

	dotIdx := strings.LastIndex(url, ".")
	if dotIdx == -1 {
		return "", errors.New("bad url")
	}

	path := "tmp/" + strconv.Itoa(time.Now().Nanosecond()) + "_" + strconv.Itoa(len(url)) + "." + url[dotIdx+1:]
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return path, nil
}

func sentPdfToSrever(pdfPath string, msg *QueueMsg) {
	file, err := os.Open(pdfPath)
	defer file.Close()
	defer os.Remove(pdfPath)

	if err != nil {
		fmt.Print(err)
		return
	}

	body := &bytes.Buffer{}
	_, err = io.Copy(body, file)

	if err != nil {
		fmt.Print(err)
		return
	}

	_, err = http.Post("http://zalupa:8080/players/"+msg.ID+"/stats/"+msg.Filename, "application/pdf", body)
}

func createPdf(msg *QueueMsg) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 25)
	pdf.MoveTo(75, 10)
	pdf.Cell(50, 15, msg.Username+"'s profile")

	pdf.SetFont("Arial", "", 20)
	pdf.SetFontSize(20)
	pdf.Ln(40)
	pdf.Cell(40, 10, "Sex: "+msg.Sex)
	pdf.Ln(20)
	pdf.Cell(40, 10, "Email: "+msg.Email)
	pdf.Ln(20)

	pdf.Cell(40, 10, "Session count:"+strconv.Itoa(msg.LossCount+msg.WinCount))
	pdf.Ln(20)
	pdf.Cell(40, 10, "Win count:"+strconv.Itoa(msg.WinCount))
	pdf.Ln(20)
	pdf.Cell(40, 10, "Loss count:"+strconv.Itoa(msg.LossCount))
	pdf.Ln(20)
	pdf.Cell(40, 10, "Duration:"+strconv.Itoa(msg.Duration)+" minutes")

	pdf.SetFont("Arial", "", 10)
	path, imgErr := getImg(msg.Avatar)
	if imgErr == nil {
		pdf.ImageOptions(path,
			128, 32,
			60, 60,
			false,
			fpdf.ImageOptions{ImageType: "JPG", ReadDpi: true},
			0,
			"",
		)
		pdf.SetFontSize(20)
		pdf.MoveTo(140, 95)
		pdf.Cell(40, 10, msg.Username)

		os.Remove(path)
	} else {
		pdf.MoveTo(140, 60)
		pdf.Cell(40, 10, "Cant find avatar!")
		log.Print(imgErr.Error())
	}

	path = pathToPdfs + "/" + msg.ID

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}

	pdfPath := path + "/" + msg.Filename
	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		return err
	}

	sentPdfToSrever(pdfPath, msg)

	return nil
}
