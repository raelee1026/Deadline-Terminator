const openModal = document.getElementById("open-modal");
const closeModal = document.getElementById("close-modal");
const taskModal = document.getElementById("task-modal");
const syncModal = document.getElementById('sync-modal');

syncModal.addEventListener('click', function() {
    console.log("click")
    fetch('/api/tasks/catch', {
        method: 'POST',  
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: 'Sync Gmail' })
    })
    .then(response => response.json())
    .then(data => {
        console.log('Gmail sync response:', data);
        alert('Gmail synced successfully!');
    })
    .catch(error => {
        console.error('Error syncing Gmail:', error);
        alert('Failed to sync Gmail.');
    });
});

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
