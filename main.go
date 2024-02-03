package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
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

	r.ParseMultipartForm(100 << 20) // Максимальный размер файла: 10MB
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	fileExt := filepath.Ext(header.Filename)

	randomFileName := generateRandomFileName()
	filePath := fmt.Sprintf("./dump/%s%v", randomFileName, fileExt)

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

	fmt.Fprint(w, 
		`<script>
			function deleteAndReload(fileName) {
				if (confirm('Вы уверены, что хотите удалить файл ' + fileName + '?')) {
					fetch('/delete/' + fileName, {
						method: 'POST'
					})
					.then(response => {
						if (response.ok) {
							location.reload();
						} else {
							alert('Не удалось удалить файл ' + fileName);
						}
					})
					.catch(error => {
						console.error('Ошибка:', error);
					});
				}
			}
		</script>`,
	)

	fmt.Fprintf(w, "<div style='display: flex;align-items: center;justify-content: flex-start;flex-direction: row;flex-wrap: wrap;'>")
	for _, file := range files {
		fileName := file.Name()
		filePath := fmt.Sprintf("/download/%s", fileName)
		fileStat, _ := file.Info()
		fileSizeBytes := fileStat.Size()
		fileSizeMB := float64(fileSizeBytes) / (1024 * 1024)
		fileModTime := fileStat.ModTime().Format("02.01.2006 15:04")

		fmt.Fprintf(w, "<div style='display: flex;padding: 2px 12px;background: #ebebeb;font-size: 14pt;font-weight: 600;border-radius: 12px;flex-direction: column;flex-wrap: nowrap;align-items: center;margin: 6px;border: 2px dashed #8d8d8d;'>")
		fmt.Fprintf(w, "<p>%s</p>", fileName)
		fmt.Fprintf(w, "<p style='margin: 6px; font-size: 11pt; width: 100px; text-align: end;'>%.2f MB</p>", fileSizeMB)
		fmt.Fprintf(w, "<p style='font-size: 10pt;font-weight: 500;color: #858585;margin: 2px 0;width: 100px;text-align: end;'>%s</p>", fileModTime)
		fmt.Fprintf(w, "<a href=\"%s\" download=\"%s\"><button style='width: 100px;height: 36px;margin: 12px  0 8px 0;border-radius: 12px;border: none;background: #878787;font-weight: 600;color: white;cursor: pointer;'>Скачать</button></a>", filePath, fileName)
		
		fmt.Fprintf(w, "<button onclick=\"deleteAndReload('%s')\" style='width: 80px;height: 28px;border-radius: 12px;background: #252525;cursor: pointer;color: white;font-weight: 600;border: none;margin: 0 0 12px 0;'>Удалить</button>", fileName)

		fmt.Fprintf(w, "</div>")

	}
	fmt.Fprintf(w, "</div>")
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[len("/download/"):]
	filePath := fmt.Sprintf("./dump/%s", fileName)
	http.ServeFile(w, r, filePath)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
  fileName := r.URL.Path[len("/delete/"):]

  filePath := fmt.Sprintf("./dump/%s", fileName)
  err := os.Remove(filePath)
  if err != nil {
    http.Error(w, "Failed to delete file", http.StatusInternalServerError)
    return
  }

  fmt.Fprintf(w, "File %s has been deleted", fileName)
}

func basicAuth(handler http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	 user, pass, ok := r.BasicAuth()
 
	 if !ok || user != username || pass != password {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	 }
	 handler(w, r)
	}
 }

func main() {

	username := os.Args[1]
	password := os.Args[2]

	if _, err := os.Stat("./dump"); os.IsNotExist(err) {
		err := os.Mkdir("./dump", 0755)
		if err != nil {
			log.Fatal("Failed to create dump directory: ", err)
		}
	}

	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/list", basicAuth(listFiles, username, password))
	http.HandleFunc("/download/", basicAuth(downloadFile, username, password))
	http.HandleFunc("/delete/", basicAuth(deleteFile, username, password))
	err := http.ListenAndServe(":18300", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
