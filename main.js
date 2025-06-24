// In main.js
function goToItem(element) { // Renamed 'btn' to 'element' for clarity
    console.log("goToItem called for link:", element.href); // Log the href of the clicked element
    window.location.href = element.href; // Navigate to the URL from the element's href
}

function scrollDown() {
  console.log("scrollDown called");
  const target = document.getElementById("projects"); // This gets the element you want to scroll to
  if (target) {
    target.scrollIntoView({ behavior: "smooth" }); // This tells the browser to scroll smoothly
  } else {
    console.warn("Scroll target element with ID 'projects' not found.");
  }
}