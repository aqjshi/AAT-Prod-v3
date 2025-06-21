// File: main.go
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type StringOrSlice []string

// UnmarshalJSON is the custom decoder for our StringOrSlice type.
func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a single string.
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = []string{str}
		return nil
	}

	// If not a string, try as a slice of strings.
	var slice []string
	if err := json.Unmarshal(data, &slice); err == nil {
		*s = slice
		return nil
	}

	return fmt.Errorf("cannot unmarshal JSON value %s into StringOrSlice", string(data))
}

// Item represents one entry from data/items.json
type Item struct {
	ID                   int      `json:"id"`
	KeywordTitle         string   `json:"keyword_title"`
	Texts                []string `json:"texts,omitempty"`
	NextpageTexts        []string `json:"nextpage_texts,omitempty"`
	NextpageImagePaths   []string `json:"nextpage_image_paths,omitempty"`
	NextpageImageCredits []string `json:"nextpage_image_credits,omitempty"`

	ImagePaths    StringOrSlice `json:"image_paths"`
	ImageCredits  []string      `json:"image_credits,omitempty"`
	ResolvedImage string        // e.g. "images/home_1_0.jpg"
	InlineStyle   string        // full CSS snippet, set in loadItems()
}

// ProjectFromJSON maps directly to the structure of each object in your projects.json file.
type ProjectFromJSON struct {
	ID            int      `json:"id"`
	KeywordTitle  string   `json:"keyword_title"`
	ImagePaths    string   `json:"image_paths"` // Main image, we can use this later if needed
	NextPageTexts []string `json:"next_page_texts"`
}

// Project is the final, clean data structure that we pass to the HTML template.
// This struct does not need to change.
type Project struct {
	ID         int
	Title      string
	Content    string
	ImagePaths []string
}

var items []Item
var projectItems []Item

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
			// norm := strings.ReplaceAll(raw, "\\", "/")
			// norm = strings.TrimPrefix(norm, "static/")
			items[i].ResolvedImage = raw
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

func loadProjectItems() {
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	filePath := filepath.Join(currDir, "data", "projects.json")

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", filePath, err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&projectItems); err != nil {
		log.Fatalf("Failed to decode projectItems.json: %v", err)
	}

	for i := range projectItems {
		if len(projectItems[i].ImagePaths) > 0 {
			raw := projectItems[i].ImagePaths[0]
			// norm := strings.ReplaceAll(raw, "\\", "/")
			// norm = strings.TrimPrefix(norm, "static/")
			projectItems[i].ResolvedImage = raw
		} else {
			projectItems[i].ResolvedImage = filepath.ToSlash(
				filepath.Join("images", fmt.Sprintf("project_%d.png", projectItems[i].ID)),
			)
		}

		// ◀︎ Use padding-bottom:75% instead of aspect-ratio
		projectItems[i].InlineStyle = fmt.Sprintf(
			"background:url('/%s') no-repeat center center;",
			projectItems[i].ResolvedImage,
		)

		fmt.Printf("Item %d image URL: %s\n", projectItems[i].ID, projectItems[i].ResolvedImage)
	}

	fmt.Printf("Loaded %d projectItems\n", len(projectItems))
}

// findImagesForProject scans the filesystem for all images belonging to a specific project ID.
func findImagesForProject(projectID int) ([]string, error) {
	var imagePaths []string
	// Construct the directory path, e.g., "static/images/1"
	dirPath := filepath.Join("static", "images", fmt.Sprintf("%d", projectID))

	files, err := os.ReadDir(dirPath)
	if err != nil {
		// It's okay if a directory doesn't exist, it just means no images.
		// We log it for debugging but don't stop the whole program.
		log.Printf("Info: No image directory found for project ID %d at '%s'", projectID, dirPath)
		return imagePaths, nil // Return empty slice
	}

	for _, file := range files {
		if file.IsDir() {
			continue // Skip subdirectories
		}
		fileName := file.Name()
		ext := strings.ToLower(filepath.Ext(fileName))

		// Check for valid image extensions
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
			// Construct the web-accessible path directly using fileName
			// This will result in paths like "static/images/1/image_name.jpg"
			webPath := filepath.ToSlash(filepath.Join(dirPath, fileName))
			imagePaths = append(imagePaths, webPath)
		}
	}
	return imagePaths, nil
}

// loadProjects now reads the JSON file and then scans for corresponding images.
func loadProjects() ([]Project, error) {
	var projectsFromJSON []ProjectFromJSON
	var finalProjects []Project

	// 1. Read and decode the main projects.json file
	jsonPath := "data/projects.json"
	jsonFile, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("could not read json file '%s': %w", jsonPath, err)
	}
	if err := json.Unmarshal(jsonFile, &projectsFromJSON); err != nil {
		return nil, fmt.Errorf("failed to decode json from '%s': %w", jsonPath, err)
	}

	// 2. For each project from JSON, process its data and find its images
	for _, p := range projectsFromJSON {
		// Find all images by scanning the corresponding directory
		images, err := findImagesForProject(p.ID)
		if err != nil {
			// Even with an error, we can still show the project text
			log.Printf("Warning: could not find images for project ID %d: %v", p.ID, err)
		}

		// Combine the data into our final, clean Project struct
		finalProject := Project{
			ID:    p.ID,
			Title: p.KeywordTitle,
			// Join the text snippets into a single block of content
			Content:    strings.Join(p.NextPageTexts, "\n\n"),
			ImagePaths: images,
		}
		finalProjects = append(finalProjects, finalProject)
	}

	return finalProjects, nil
}

// --------------------------
// 4. MIDDLEWARE & HANDLERS
// --------------------------

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "American Advanced Technology",
		"Items": items,
	}
	if err := tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func projectHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "American Advanced Technology",
		"Items": projectItems,
	}
	if err := tmpl.ExecuteTemplate(w, "project.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// project_item_Handler handles requests for individual project pages.
func project_item_Handler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse the project ID from the URL query parameters.
	projectIDStr := r.URL.Query().Get("id")
	if projectIDStr == "" {
		http.Error(w, "Project ID is missing from URL.", http.StatusBadRequest)
		return
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		http.Error(w, "Invalid Project ID format.", http.StatusBadRequest)
		return
	}

	// 2. Load all projects. This should ideally be done once or cached if performance is critical,
	// but for now, re-loading is fine for simplicity given it reads from a local JSON and disk.
	allProjects, err := loadProjects()
	if err != nil {
		http.Error(w, "Failed to load projects data.", http.StatusInternalServerError)
		log.Printf("Error loading projects: %v", err)
		return
	}

	// 3. Find the specific project by ID.
	var foundProject *Project
	for i := range allProjects {
		if allProjects[i].ID == projectID {
			foundProject = &allProjects[i]
			break
		}
	}

	if foundProject == nil {
		http.Error(w, "Project not found.", http.StatusNotFound)
		return
	}

	// 4. Prepare data for the template.
	// We are passing a single Project object to the template.
	data := struct {
		Title   string
		Project *Project
		Headers map[string]map[string]string // Pass headers for navigation/footer
		Footer  map[string]map[string]string // Pass footer for navigation/footer
	}{
		Title:   foundProject.Title, // Use the project's title for the page title
		Project: foundProject,
		Headers: HEADERS,
		Footer:  FOOTER,
	}

	// 5. Execute the template.
	// The `tmpl` variable is already global and initialized with all templates via ParseGlob.
	if err := tmpl.ExecuteTemplate(w, "project_item.html", data); err != nil {
		http.Error(w, "Failed to render project item page.", http.StatusInternalServerError)
		log.Printf("Error rendering project_item.html: %v", err)
	}
}

func main() {
	// 1) Load and resolve items
	loadItems()
	loadProjectItems()
	// Parse templates...
	// Parse all HTML templates in the templates directory
	var err error
	tmpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	// Serve index.html at root

	// 2) Dynamic handler for the home page:
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/project", projectHandler)
	http.HandleFunc("/project_item.html", project_item_Handler) // Handler for individual project details

	// 3) Serve everything under ./static/ at URL path /static/
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/home.html")
	})

	http.HandleFunc("/consulting", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/consulting.html")
	})
	http.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/service.html")
	})
	http.HandleFunc("/research", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/research.html")
	})
	http.HandleFunc("/training", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/training.html")
	})
	http.HandleFunc("/product", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/product.html")
	})

	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/contact.html")
	})
	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/privacy.html")
	})
	http.HandleFunc("/icp", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/icp.html")
	})

	http.HandleFunc("/tou", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/tou.html")
	})
	http.HandleFunc("/non", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/non.html")
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

	http.HandleFunc("/training_inquiry_form.pdf", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/files/training_inquiry_form.pdf")
	})
	http.HandleFunc("/static/files/product_inquiry_form.pdf", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/files/product_inquiry_form.pdf")
	})

	ln, err := net.Listen("tcp4", ":8080")
	if err != nil {
		log.Fatalf("Failed to bind to IPv4: %v", err)
	}
	log.Println("Listening on http://0.0.0.0:8080 …")
	log.Fatal(http.Serve(ln, nil))
}
