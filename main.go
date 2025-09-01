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
// Define the struct for a single service category and its items
type ServiceCategory struct {
    Title    string
    Services []string
}

var serviceCategories []ServiceCategory

func loadServiceCategories() {

    serviceCategories = []ServiceCategory{
        {
            Title: "Optics, Imaging and Electronics",
            Services: []string{
                "Photography",
                "3D Laser Scanning and Mapping",
                "LiDAR Technology and Mapping",
                "3D Photogrammetry and Mapping",
                "Engineering Surveying",
                "Digital Image Correlation",
                "HighPrecision 3D Stereo Imaging",
                "Remote Sensing",
                "Computer Vision",
                "Internet of Things (IoT) Sensor",
            },
        },
        {
            Title: "Information and Telecommunication",
            Services: []string{
                "Computer Aided Design and Analysis (CAD)",
                "3D Printing",
                "Building Information Modeling (BIM)",
                "Dashboard Design and Development",
                "Data Visualization",
                "Web Design and Development",
                "App Design and Development",
                "Software Design and Development",
                "Optical Communication",
                "Wireless Communication",
                "Virtual Reality",
                "Augmented Reality",
            },
        },
        {
            Title: "Infrastructure and Facility Asset Management",
            Services: []string{
                "Inspection",
                "Monitoring",
                "Condition Assessment",
            },
        },
        {
            Title: "Project and Construction Management",
            Services: []string{
                "Utility Coordination", // Added based on full text
                "Project Management",
                "Program Management", // Added based on full text
                "Risk Management", // Added based on full text
                "Project Scheduling", // Added based on full text
                "Project Controls", // Added based on full text
                "Cost Estimating", // Added based on full text
                "Construction Inspection, QC Inspection and Quality Management", // Added based on full text
                "Construction Management",
                "Engineering Design Support", // Added based on full text
                "Office Engineering, Document Control and Staff Augmentation", // Added based on full text
                "Communication and Public Outreach", // Added based on full text
            },
        },
        {
            Title: "Artificial Intelligence",
            Services: []string{
                "AI Center Buildup and Operations", // Added based on full text
                "Machine Learning",
                "Artificial Intelligence (AI) Algorithm and Deployment", // Added based on full text
            },
        },
        {
            Title: "Applied Mathematics, Computing and Data Analytics",
            Services: []string{
                "Data Center Buildup and Operations", // Added based on full text
                "Mathematical Modeling",
                "Statistical Science",
                "Optimization",
                "Data Analytics",
                "Data Driven Decision Making",
                "HighPerformance Computing for Partial Differential Equations",
                "Multiscale and Multiphysics Modeling and Simulation",
                "Cloud Computing",
            },
        },
        {
            Title: "Transportation and Logistics",
            Services: []string{
                "Traffic Congestion Prediction and Mitigation",
                "Traffic Simulation",
                "Traffic Signal Design and Optimization",
                "Transportation Systems Management and Operations (TSMO)",
                "Intelligent Transportation System",
                "Connected and Autonomous Vehicles", // Added based on full text
                "Highway Geometric Design",
                "Rail Transportation Design and & Operations", // Corrected '&'
                "Airport Pavement Design, Operations and Maintenance",
                "Smart Traveler Information System for Transit and Subway", // Combined with Fleet from user's list
                "Fleet Management", // Moved here
                "Smart Driving with Optimal Vehicle Routing", // Moved here
            },
        },
        {
            Title: "Energy", // Corrected to "Energy" from "Clean Energy" for consistency with user's full text
            Services: []string{
                "Monitoring, Inspection and Diagnose for Wind, Solar, Nuclear, Natural Gas, Hydroelectric and Geothermal Energy", // Combined from user's full text
                "Battery Management System", // Moved here
                "Geothermal Energy for Cold Region Subgrade", // Moved here
                "Active Suspension and Control of Vehicle", // Moved here
            },
        },
        {
            Title: "Finance and Insurance",
            Services: []string{
                "Budgeting Service for Nonprofit Organization and Small Business", // Corrected "Non‑profit"
                "Accounting Service for Nonprofit Organization and Small Business", // Corrected "Non‑profit"
                "Financial Risk Management",
                "Financial Asset Management",
                "AI Congestion-Based Pricing", // Added based on full text
                "AI Vehicle Insurance", // Added based on full text
                "AI Quantitative Trading", // Added based on full text
                "Community Development", // Added based on full text
            },
        },
        {
            Title: "Environmental and Public Health", // Combined from "Health" in user's full text
            Services: []string{
                "Cognitive, Neurological and Behavioral Driving Safety and Injury Prevention",
                "AI Precision Medicine", // Added based on full text
                "Environmental Monitoring",
                "Environmental Health and Safety",
                "Air Pollution Effect on Public Health", // Corrected "Effect"
                "Smart Device for Healthcare",
                "Neurological Rehabilitation",
            },
        },
        {
            Title: "Operations Safety", // Corrected to "Operations Safety" from just "Safety" for clarity
            Services: []string{
                "Traffic Safety",
                "Transportation Safety",
                "Work Zone Safety",
                "Construction Safety",
                "Bridge, Tunnel and Slope Structure Safety",
                "Hazard and Disaster Safety and Mitigation",
                "Fire Safety",
                "Methane Safety",
                "Battery Safety",
            },
        },
        {
            Title: "Nanotechnology and Sustainable Pavement", // Corrected title
            Services: []string{
                "Nano Composite Material for Asphalt Modification",
                "Performance Engineered Balanced Pavement Mixture and Design", // Corrected from "Performance Engineered Pavement Mixture Design"
                "Solid Waste and Reclaimed Asphalt Pavement for Sustainable and Low Carbon Pavement", // Corrected title
                "Nondestructive Testing of Airport and Road Pavement",
            },
        },
        {
            Title: "Drone Applications",
            Services: []string{
                "Exterior Window and Solar Panel Washing", // Combined
                "Building and Infrastructure Painting",
                "Drone Surveying and Mapping for Building, Infrastructure, Facility, etc.", // Combined
            },
        },
    }
    log.Printf("Loaded %d service categories.", len(serviceCategories))
}
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

var IMAGE_EXTS = []string{"jpg", "jpeg", "png"}

var tmpl *template.Template

type StringOrSlice []string


func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a single string.
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = []string{str}
		return nil
	}

	var slice []string
	if err := json.Unmarshal(data, &slice); err == nil {
		*s = slice
		return nil
	}

	return fmt.Errorf("cannot unmarshal JSON value %s into StringOrSlice", string(data))
}
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
	Link          string   // **NEW**: This will hold the URL for each item
}

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

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "About Us",

	}
	if err := tmpl.ExecuteTemplate(w, "about.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}



func consultingHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Consulting",

	}
	if err := tmpl.ExecuteTemplate(w, "consulting.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func contactHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Contact",
	}
	if err := tmpl.ExecuteTemplate(w, "contact.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func emptyHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Empty",

	}
	if err := tmpl.ExecuteTemplate(w, "empty.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func icpHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Indirect Cost Policy",

	}
	if err := tmpl.ExecuteTemplate(w, "icp.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}





func nonHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Nondiscrimination",

	}
	if err := tmpl.ExecuteTemplate(w, "non.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func privacyHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Privacy",

	}
	if err := tmpl.ExecuteTemplate(w, "privacy.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func productHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Products",

	}
	if err := tmpl.ExecuteTemplate(w, "product.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func researchHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Research",

	}
	if err := tmpl.ExecuteTemplate(w, "research.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}



func touHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Terms of Use",

	}
	if err := tmpl.ExecuteTemplate(w, "tou.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func trainingHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Training",

	}
	if err := tmpl.ExecuteTemplate(w, "training.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


func projectHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Projects",
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
func serviceHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "Title":      "Service", // This is for the page's <title> and <h1>Service</h1>
        "Categories": serviceCategories, // Pass the new service data
        "Headers":    HEADERS,          // Ensure common data for header/footer is passed
        "Footer":     FOOTER,           // Ensure common data for header/footer is passed
    }
    if err := tmpl.ExecuteTemplate(w, "service.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func main() {
	// 1) Load and resolve items
	loadItems()
	loadProjectItems()
	loadServiceCategories() // *** CALL THIS NEW FUNCTION HERE ***

	// Parse templates...
	// Parse all HTML templates in the templates directory
	var err error

	tmpl, err = template.ParseFiles(
		"templates/header.html",
		"templates/footer.html",
		"templates/home.html",
		"templates/about.html",
		"templates/consulting.html",
		"templates/contact.html",
		"templates/empty.html",
		"templates/icp.html",
		"templates/non.html",
		"templates/privacy.html",
		"templates/product.html",
		"templates/project_item.html", 
		"templates/project.html",
		"templates/research.html", 
		"templates/service.html", 
		"templates/tou.html", 
		"templates/training.html", 
	)
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}



	// 2) Dynamic handler for the home page:
	http.HandleFunc("/", homeHandler)
	// 3) Serve everything under ./static/ at URL path /static/
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/main.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "main.js")
	})



	http.HandleFunc("/project", projectHandler)
	http.HandleFunc("/project_item.html", project_item_Handler) // Handler for individual project details

	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/consulting", consultingHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/empty", emptyHandler)
	http.HandleFunc("/icp", icpHandler)
	http.HandleFunc("/non", nonHandler)
	http.HandleFunc("/privacy", privacyHandler)
	http.HandleFunc("/product", productHandler)
	http.HandleFunc("/research", researchHandler)
	http.HandleFunc("/service", serviceHandler)
	http.HandleFunc("/tou", touHandler)
	http.HandleFunc("/training", trainingHandler)


	
	// Serve the CSS file at /styles.css
	http.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "styles.css")
	})



	// http.HandleFunc("/research", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/research.html")
	// })
	// http.HandleFunc("/training", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/training.html")
	// })
	// http.HandleFunc("/product", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/product.html")
	// })

	// http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/contact.html")
	// })
	// http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/privacy.html")
	// })
	// http.HandleFunc("/icp", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/icp.html")
	// })

	// http.HandleFunc("/tou", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/tou.html")
	// })
	// http.HandleFunc("/non", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/non.html")
	// })
	// 	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "templates/about.html")
	// })

	// Serve the video file at /aerial.mp4
	http.HandleFunc("/aerial.mp4", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/video/aerial.mp4")
	})
	http.HandleFunc("/home_background", func(w http.ResponseWriter, r *http.Request) {
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
