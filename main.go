package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func generateRandomFileName() string {
	rand.Seed(time.Now().UnixNano())
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 5)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	if _, err := os.Stat("./dump"); os.IsNotExist(err) {
		err := os.Mkdir("./dump", 0755)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Failed to create dump directory", http.StatusInternalServerError)
			return
		}
	}

	r.ParseMultipartForm(10 << 20) // Максимальный размер файла: 10MB
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Генерируем случайное имя файла из 5 символов
	randomFileName := generateRandomFileName()
	filePath := fmt.Sprintf("./dump/%s.%v", randomFileName, "zip")

	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("./dump")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<div style='display: flex;align-items: center;justify-content: flex-start;flex-direction: row;flex-wrap: wrap;'>")
	for _, file := range files {
		fileName := file.Name()
		filePath := fmt.Sprintf("/download/%s", fileName)
		fileStat, _ := file.Info()
		fileSizeBytes := fileStat.Size()
		fileSizeMB := float64(fileSizeBytes) / (1024 * 1024)
		fileModTime := fileStat.ModTime().Format("02.01.2006 15:04")

		fmt.Fprintf(w, "<div style='display: flex;padding: 2px 12px;background: #ebebeb;font-size: 14pt;font-weight: 600;border-radius: 12px;flex-direction: column;flex-wrap: nowrap;align-items: center;margin: 6px;'>")
		fmt.Fprintf(w, "<p>%s</p>", fileName)
		fmt.Fprintf(w, "<p style='margin: 6px; font-size: 11pt; width: 100px; text-align: end;'>%.2f MB</p>", fileSizeMB)
		fmt.Fprintf(w, "<p style='font-size: 10pt;font-weight: 500;color: #858585;margin: 2px 0;width: 100px;text-align: end;'>%s</p>", fileModTime)
		fmt.Fprintf(w, "<a href=\"%s\" download=\"%s\"><button style='width: 100px; height: 36px; margin: 12px  0 0 0; border-radius: 12px; background: aquamarine; cursor: pointer;'>Скачать</button></a><br>", filePath, fileName)
		fmt.Fprintf(w, "</div>")

	}
	fmt.Fprintf(w, "</div>")
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[len("/download/"):]
	filePath := fmt.Sprintf("./dump/%s", fileName)
	http.ServeFile(w, r, filePath)
}

func main() {

	if _, err := os.Stat("./dump"); os.IsNotExist(err) {
		err := os.Mkdir("./dump", 0755)
		if err != nil {
			log.Fatal("Failed to create dump directory: ", err)
		}
	}

	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/dumps", listFiles)
	http.HandleFunc("/download/", downloadFile)
	err := http.ListenAndServe(":18300", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
