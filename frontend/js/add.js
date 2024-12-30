const openModal = document.getElementById("open-modal");
const closeModal = document.getElementById("close-modal");
const taskModal = document.getElementById("task-modal");
const syncModal = document.getElementById('sync-modal');
const syncButtonText = syncModal.textContent; 

const overlay = document.createElement('div');
overlay.classList.add('overlay');
document.body.appendChild(overlay); 

syncModal.addEventListener('click', function() {
    // 顯示 loading 並禁用按鈕
    syncModal.disabled = true;
    syncModal.textContent = "Syncing..."; // 顯示 loading 文字
    
    // 顯示 loading 圖示並啟動遮罩層
    const loadingSpinner = document.createElement('span');
    loadingSpinner.classList.add('loading-spinner'); // 顯示 loading 樣式
    syncModal.appendChild(loadingSpinner);
    overlay.style.display = 'block';  // 顯示遮罩層

    // 發送同步請求
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
    })
    .finally(() => {
        // 請求完成後恢復按鈕狀態
        syncModal.disabled = false;
        syncModal.textContent = syncButtonText; 
        loadingSpinner.remove(); 
        overlay.style.display = 'none'; 
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
