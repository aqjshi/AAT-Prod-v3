// File: main.go
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Headers/Footer structures mirror your Python HEADERS/FOOTER dictionaries.
var HEADERS = map[string]map[string]string{
	"home":       {"title": "Home", "background": "static/images/home_background.jpg"},
	"service":    {"title": "Service"},
	"consulting": {"title": "Consulting"},
	"research":   {"title": "Research & Development"},
	"training":   {"title": "Training"},
	"product":    {"title": "product"},
	"contact":    {"title": "Contact"},
}
var FOOTER = map[string]map[string]string{
	"terms_of_use": {"title": "tou"},
	"privacy":      {"title": "Privacy"},
}
var base *template.Template

// IMAGE_EXTS is the list of file extensions to try when resolving “home_{id}_0.*”
var IMAGE_EXTS = []string{"jpg", "jpeg", "png"}

var tmpl *template.Template

// Item represents one entry from data/items.json
type Item struct {
	ID                   int      `json:"id"`
	KeywordTitle         string   `json:"keyword_title"`
	Texts                []string `json:"texts,omitempty"`
	NextpageTexts        []string `json:"nextpage_texts,omitempty"`
	NextpageImagePaths   []string `json:"nextpage_image_paths,omitempty"`
	NextpageImageCredits []string `json:"nextpage_image_credits,omitempty"`
	ImagePaths           []string `json:"image_paths,omitempty"`
	ImageCredits         []string `json:"image_credits,omitempty"`

	ResolvedImage string // e.g. "images/home_1_0.jpg"
	InlineStyle   string // full CSS snippet, set in loadItems()
}

var items []Item

// configureLogging sets up a rotating log writer (optional)
func configureLogging() io.Writer {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		if err := os.Mkdir("logs", 0755); err != nil {
			fmt.Fprintln(os.Stderr, "Cannot create logs directory:", err)
		}
	}

	return &lumberjack.Logger{
		Filename:   "logs/goapp.log",
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     30,   // days
		Compress:   true, // optional
	}
}

func loadItems() {
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	filePath := filepath.Join(currDir, "data", "items.json")

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", filePath, err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&items); err != nil {
		log.Fatalf("Failed to decode items.json: %v", err)
	}

	for i := range items {
		if len(items[i].ImagePaths) > 0 {
			raw := items[i].ImagePaths[0]
			norm := strings.ReplaceAll(raw, "\\", "/")
			norm = strings.TrimPrefix(norm, "static/")
			items[i].ResolvedImage = norm
		} else {
			items[i].ResolvedImage = filepath.ToSlash(
				filepath.Join("images", fmt.Sprintf("home_%d_0.jpg", items[i].ID)),
			)
		}

		// ◀︎ Use padding-bottom:75% instead of aspect-ratio
		items[i].InlineStyle = fmt.Sprintf(
			"background:url('/%s') no-repeat center center;",
			items[i].ResolvedImage,
		)

		fmt.Printf("Item %d image URL: %s\n", items[i].ID, items[i].ResolvedImage)
	}

	fmt.Printf("Loaded %d items\n", len(items))
}

// filterItemsWithImage returns only those items whose ResolvedImage != ""
func filterItemsWithImage() []Item {
	res := make([]Item, 0, len(items))
	for _, it := range items {
		if it.ResolvedImage != "" {
			res = append(res, it)
		}
	}
	return res
}

// findItemByID searches the global slice
func findItemByID(id int) *Item {
	for _, it := range items {
		if it.ID == id {
			return &it
		}
	}
	return nil
}

// --------------------------
// 4. MIDDLEWARE & HANDLERS
// --------------------------

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// wrapWithLogging is a simple middleware that logs method/path + timestamp.
func wrapWithLogging(next HandlerFunc, logger io.Writer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log a line: [timestamp] METHOD PATH
		fmt.Fprintf(logger, "%s %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		next(w, r)
	}
}

func serveTextHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /text/<filename>
	// We serve from "./text"
	filename := filepath.Base(r.URL.Path[len("/text/"):]) // sanitize
	http.ServeFile(w, r, filepath.Join("text", filename))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "American Advanced Technology",
		"Items": items,
	}
	if err := tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func consultingHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Items":  filterItemsWithImage(),
		"Title":  HEADERS["consulting"]["title"],
		"Footer": FOOTER,
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  HEADERS["service"]["title"],
		"Footer": FOOTER,
		// If you want to pass service_categories, you can add them here as in Flask.
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func researchHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  HEADERS["research"]["title"],
		"Footer": FOOTER,
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func trainingHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  HEADERS["training"]["title"],
		"Footer": FOOTER,
	}

	// Clone base, then parse ONLY training.html into it
	tmpl := template.Must(base.Clone())
	template.Must(tmpl.ParseFiles("templates/training.html"))

	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  HEADERS["product"]["title"],
		"Footer": FOOTER,
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  HEADERS["contact"]["title"],
		"Footer": FOOTER,
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func privacyHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Privacy",
		"Footer": FOOTER,
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func touHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":  "Terms of Use",
		"Footer": FOOTER,
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func itemHandler(w http.ResponseWriter, r *http.Request) {
	// Expect query parameter “id”
	q := r.URL.Query().Get("id")
	if q == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(q)
	if err != nil {
		http.Error(w, "Invalid id parameter", http.StatusBadRequest)
		return
	}
	matched := findItemByID(id)
	if matched == nil {
		http.Error(w, fmt.Sprintf("No item with id=%d", id), http.StatusNotFound)
		return
	}
	data := map[string]interface{}{
		"Item": matched,
	}
	if err := tmpl.ExecuteTemplate(w, "/base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func submitInquiryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	// In Flask you did request.form.to_dict(flat=False). Here:
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	// Log posted form data:
	fmt.Fprintf(os.Stdout, "New inquiry submitted: %v\n", r.PostForm) // or use logger
	// In Flask you flash+redirect; here we’ll just redirect
	http.Redirect(w, r, "/product", http.StatusSeeOther)
}

func main() {
	// 1) Load and resolve items
	loadItems()

	// Parse templates...
	// Parse all HTML templates in the templates directory
	var err error
	tmpl, err = template.ParseGlob("*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	// Serve index.html at root

	// 2) Dynamic handler for the home page:
	http.HandleFunc("/", homeHandler)

	// 3) Serve everything under ./static/ at URL path /static/
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "home.html")
	})

	http.HandleFunc("/consulting", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "consulting.html")
	})
	http.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "service.html")
	})
	http.HandleFunc("/research", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "research.html")
	})
	http.HandleFunc("/training", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "training.html")
	})
	http.HandleFunc("/product", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "product.html")
	})

	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "contact.html")
	})
	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "privacy.html")
	})
	http.HandleFunc("/icp", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "icp.html")
	})

	http.HandleFunc("/tou", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "tou.html")
	})

	// Serve the CSS file at /styles.css
	http.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "styles.css")
	})
	// Serve the video file at /aerial.mp4
	http.HandleFunc("/aerial.mp4", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/video/aerial.mp4")
	})
	http.HandleFunc("/home_background.jpg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_background.jpg")
	})

	http.HandleFunc("/product_BMS.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/product_BMS.png")
	})
	http.HandleFunc("/product_AQMS.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/products_AQMS.png")
	})
	http.HandleFunc("/product_MDS.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/products_MDS.png")
	})
	http.HandleFunc("/product_NCAM.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/products_NCAM.png")
	})
	http.HandleFunc("/product_TT.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/products_TT.png")
	})

	http.HandleFunc("/main.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "main.js")
	})
	http.HandleFunc("/items.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./data/items.json")
	})
	http.HandleFunc("/images/home_1_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_1_0.png")
	})
	http.HandleFunc("/images/home_2_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_2_0.png")
	})
	http.HandleFunc("/images/home_3_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_3_0.png")
	})
	http.HandleFunc("/images/home_4_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_4_0.png")
	})
	http.HandleFunc("/images/home_5_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_5_0.png")
	})
	http.HandleFunc("/images/home_6_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_6_0.png")
	})
	http.HandleFunc("/images/home_7_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_7_0.png")
	})
	http.HandleFunc("/images/home_8_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_8_0.png")
	})
	http.HandleFunc("/images/home_9_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_9_0.png")
	})
	http.HandleFunc("/images/home_10_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_10_0.png")
	})
	http.HandleFunc("/images/home_11_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_11_0.png")
	})
	http.HandleFunc("/images/home_12_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_12_0.png")
	})
	http.HandleFunc("/images/home_13_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_13_0.png")
	})
	http.HandleFunc("/images/home_14_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_14_0.png")
	})
	http.HandleFunc("/images/home_15_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_15_0.png")
	})
	http.HandleFunc("/images/home_16_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_16_0.png")
	})
	http.HandleFunc("/images/home_17_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_17_0.png")
	})
	http.HandleFunc("/images/home_18_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_18_0.png")
	})
	http.HandleFunc("/images/home_19_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_19_0.png")
	})
	http.HandleFunc("/images/home_20_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_20_0.png")
	})
	http.HandleFunc("/images/home_21_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_21_0.png")
	})
	http.HandleFunc("/images/home_22_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_22_0.png")
	})
	http.HandleFunc("/images/home_23_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_23_0.png")
	})
	http.HandleFunc("/images/home_24_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_24_0.png")
	})
	http.HandleFunc("/images/home_25_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_25_0.png")
	})
	http.HandleFunc("/images/home_26_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_26_0.png")
	})
	http.HandleFunc("/images/home_27_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_27_0.png")
	})
	http.HandleFunc("/images/home_28_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_28_0.png")
	})
	http.HandleFunc("/images/home_29_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_29_0.png")
	})
	http.HandleFunc("/images/home_30_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_30_0.png")
	})
	http.HandleFunc("/images/home_31_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_31_0.png")
	})
	http.HandleFunc("/images/home_32_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_32_0.png")
	})
	http.HandleFunc("/images/home_33_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_33_0.png")
	})
	http.HandleFunc("/images/home_34_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_34_0.png")
	})
	http.HandleFunc("/images/home_35_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_35_0.png")
	})
	http.HandleFunc("/images/home_36_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_36_0.png")
	})
	http.HandleFunc("/images/home_37_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_37_0.png")
	})
	http.HandleFunc("/images/home_38_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_38_0.png")
	})
	http.HandleFunc("/images/home_39_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_39_0.png")
	})
	http.HandleFunc("/images/home_40_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_40_0.png")
	})
	http.HandleFunc("/images/home_41_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_41_0.png")
	})
	http.HandleFunc("/images/home_42_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_42_0.png")
	})
	http.HandleFunc("/images/home_43_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_43_0.png")
	})
	http.HandleFunc("/images/home_44_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_44_0.png")
	})
	http.HandleFunc("/images/home_45_0.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/images/home_45_0.png")
	})

	http.HandleFunc("/training_inquiry_form.pdf", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/files/training_inquiry_form.pdf")
	})
	http.HandleFunc("/static/files/product_inquiry_form.pdf", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/files/product_inquiry_form.pdf")
	})
	log.Println("Listening on http://localhost:8080 …")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
