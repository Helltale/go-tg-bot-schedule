package fileutils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func MustOpenJPG(filename string) *os.File {
	filePath := filepath.Join("D:\\", "projects", "golang", "tg-program", "client-tg", "other", "imgs", filename+".jpg")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Ошибка: файл не найден по пути: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", filePath, err)
	}
	return file
}

func MustOpenDOCX(filename string) *os.File {
	filePath := filepath.Join("D:\\", "projects", "golang", "tg-program", "client-tg", "other", "docs", filename+".docx")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Ошибка: файл не найден по пути: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", filePath, err)
	}
	return file
}

func MustOpenPDF(filename string) *os.File {
	filePath := filepath.Join("D:\\", "projects", "golang", "tg-program", "client-tg", "other", "docs", filename+".pdf")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("Ошибка: файл не найден по пути: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Ошибка при открытии файла %s: %v", filePath, err)
	}
	return file
}

func GetFileInfo(document string) string {
	var message string

	switch document {
	case "doc1":
		fileInfo, err := os.Stat(filepath.Join("d:\\", "projects", "golang", "tg-program", "client-tg", "other", "docs", document+".docx"))
		if err != nil {
			fmt.Println("Ошибка:", err)
			return message
		}
		message += fmt.Sprintf("Название: %s\n", fileInfo.Name())
		message += fmt.Sprintf("Размер: %d б\n", fileInfo.Size())
		message += fmt.Sprintf("Дата обновления: %s\n", fileInfo.ModTime().Format(time.DateOnly))

		return message

	default:
		return message
	}
}
