package main

// The directory containing the images, relative to the project root.
const imageDir = "static/images/3"

// Data structure to pass to the HTML template
type PageData struct {
	ImagePaths []string
}

// The HTML template with embedded CSS
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Image Gallery</title>
    <style>
        /* Basic reset and body styling */
        body {
            margin: 0;
            font-family: sans-serif;
            background-color: #f0f0f0;
        }

        /* The main container for the image column */
        .image-container {
            display: flex;
            flex-direction: column;
            align-items: center; /* Center images horizontally */
            gap: 20px; /* Space between images */
            padding: 20px;
            box-sizing: border-box;
			width: 100%;
			height: 100vh; /* Full viewport height */
			overflow-y: scroll; /* Enable vertical scrolling */
        }

        /* Wrapper for each image to enforce aspect ratio and overflow */
        .image-wrapper {
            width: 70%; /* Adjust width as needed */
            max-width: 300px;
            aspect-ratio: 4 / 3;

            border-radius: 8px;
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }

        /* The image itself */
        .image-wrapper img {
            width: 100%;
            height: 100%;
            object-fit: fill; /* Stretches the image to fill the container */
            display: block;
        }
    </style>
</head>
<body>
    <div class="image-container">
        <h1>Image Gallery</h1>
        {{range .ImagePaths}}
            <div class="image-wrapper">
                <img src="{{.}}">
            </div>
        {{end}}
    </div>
</body>
</html>
`

// func main() {
// 	// Create a new template and parse the HTML string
// 	tmpl, err := template.New("gallery").Parse(htmlTemplate)
// 	if err != nil {
// 		log.Fatalf("Failed to parse template: %v", err)
// 	}

// 	// Handler for the root URL ("/")
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		// Get the absolute path to the image directory
// 		absPath, err := filepath.Abs(imageDir)
// 		if err != nil {
// 			http.Error(w, "Failed to get absolute path", http.StatusInternalServerError)
// 			return
// 		}

// 		// Read all file entries in the directory
// 		files, err := os.ReadDir(absPath)
// 		if err != nil {
// 			log.Printf("Could not read image directory '%s': %v", absPath, err)
// 			http.Error(w, "Image directory not found.", http.StatusInternalServerError)
// 			return
// 		}

// 		var imagePaths []string
// 		for _, file := range files {
// 			// Skip directories
// 			if !file.IsDir() {
// 				// Create the web-accessible path for the image
// 				// Note: using filepath.ToSlash to ensure forward slashes for URL
// 				webPath := filepath.ToSlash(filepath.Join("/", imageDir, file.Name()))
// 				imagePaths = append(imagePaths, webPath)
// 			}
// 		}

// 		// Execute the template with the list of image paths
// 		data := PageData{ImagePaths: imagePaths}
// 		err = tmpl.Execute(w, data)
// 		if err != nil {
// 			http.Error(w, "Failed to render template", http.StatusInternalServerError)
// 		}
// 	})

// 	// Handler to serve static files (images)
// 	// This serves the entire 'static' directory.
// 	// The path in the HTML <img src="/static/..."> will be correctly resolved.
// 	staticFileServer := http.FileServer(http.Dir("."))
// 	http.Handle("/static/", staticFileServer)

// 	fmt.Println("Server starting on http://localhost:8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
