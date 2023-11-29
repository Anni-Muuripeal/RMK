package main

import (
	"crypto/tls"
	"encoding/xml"
	"html/template"
	"log"
	"net/http"
)

const (
	xmlURL = "https://juhendid.rmk.ee/app-structure.xml"
	port   = ":8080"
)

var tmpl *template.Template

type Document struct {
	Label string `xml:"label,attr"`
	Order string `xml:"order,attr"`
	Type  string `xml:"type,attr"`
	URL   string `xml:",innerxml"`
}

type Classification struct {
	Label     string     `xml:"label,attr"`
	Desc      string     `xml:"desc,attr"`
	Order     string     `xml:"order,attr"`
	Documents []Document `xml:"document"`
}

type AndDocuments struct {
	Classifications []Classification `xml:"classification"`
}

func init() {
	var err error
	tmpl, err = template.New("htmlTemplate").Parse(htmlTemplate)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/", homeHandler)

	log.Printf("Starting server at port %s\nOpen http://localhost%s\nUse Ctrl+C to close the server\n", port, port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// this serves static files to avoid MIME type errors
func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

// fetche and parse XML data, renders the HTML template
func homeHandler(w http.ResponseWriter, r *http.Request) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Get(xmlURL)
	if err != nil {
		log.Println("Error fetching XML:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var result AndDocuments
	err = xml.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println("Error parsing XML:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, result)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<link rel="stylesheet" type="text/css" href="/static/styles.css">
	<title>RMK</title>
</head>
<body>
	{{range .Classifications}}
		<h2>{{.Label}}</h2>
		<p>{{.Desc}}</p>
		<ul>
		{{range .Documents}}
			<li><a href="{{.URL}}" target="_blank">{{.Label}}</a></li>
		{{end}}
		</ul>
	{{end}}
</body>
</html>
`
