const openModal = document.getElementById("open-modal");
const closeModal = document.getElementById("close-modal");
const taskModal = document.getElementById("task-modal");

// Open modal
openModal.addEventListener("click", () => {
    taskModal.style.display = "flex";
});

// Close modal
closeModal.addEventListener("click", () => {
    taskModal.style.display = "none";
});

// Optional: Close modal when clicking outside the modal content
taskModal.addEventListener("click", (e) => {
    if (e.target === taskModal) {
        taskModal.style.display = "none";
    }
});
