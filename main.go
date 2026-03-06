package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	binaryftp "binary-go/binaryftp/client"
)

const (
	ftpAddr = "localhost:9000"
	tmpDir  = "./tmp"
	downDir = "./downloads"
)

var page = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
<title>BinaryFTP File Server</title>
<style>
body{font-family:Arial;margin:40px}
li{margin:5px 0}
</style>
</head>
<body>

<h2>Upload File</h2>

<form action="/upload" method="post" enctype="multipart/form-data">
<input type="file" name="file">
<button type="submit">Upload</button>
</form>

<h2>Files</h2>

<ul>
{{range .}}
<li>
{{.}}
<a href="/download?file={{.}}">download</a>

<form action="/delete" method="post" style="display:inline">
<input type="hidden" name="file" value="{{.}}">
<button type="submit">delete</button>
</form>

</li>
{{end}}
</ul>

</body>
</html>
`))

var client = binaryftp.New(ftpAddr)

func main() {

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/delete", deleteHandler)

	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	files, err := client.ListFiles()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	page.Execute(w, files)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(100 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	tmpPath := filepath.Join(tmpDir, header.Filename)

	dst, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	err = client.Upload(tmpPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	os.Remove(tmpPath)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {

	name := filepath.Base(r.URL.Query().Get("file"))

	localPath := filepath.Join(downDir, name)

	err := client.Download(name, localPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	http.ServeFile(w, r, localPath)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	http.Error(w, "delete not implemented in binaryftp client", 501)
}
