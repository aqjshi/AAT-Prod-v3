function goToItem(btn) {
    const id = btn.dataset.id;
    window.location.href = `/item?id=${id}`;
  }




function scrollDown() {
  console.log("scrollDown called");
  const target = document.getElementById("services");
  if (target) {
    target.scrollIntoView({ behavior: "smooth" });
  }
}
