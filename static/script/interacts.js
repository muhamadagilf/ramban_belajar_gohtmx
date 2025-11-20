document.addEventListener('DOMContentLoaded', function() {
  const rightContentCard = document.getElementById("right-content-card");
  const newUserBtn = document.getElementById("new-user-btn");
  const usersBottomSection = document.getElementById("bottom-section");

  if (rightContentCard) {
      rightContentCard.classList.remove("opacity-0")
      rightContentCard.classList.add("opacity-100")
  }

  newUserBtn.addEventListener('click', function() {
      usersBottomSection.classList.toggle("hidden");
  })

})
