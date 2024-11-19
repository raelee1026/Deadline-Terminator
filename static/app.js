document.addEventListener("DOMContentLoaded", () => {
    const taskList = document.getElementById("task-list");
    const taskForm = document.getElementById("task-form");

    // 載入任務列表
    function loadTasks() {
        fetch("/api/tasks")
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then((tasks) => {
                taskList.innerHTML = ""; // 清空任務列表
                tasks.forEach((task) => {
                    const taskDiv = document.createElement("div");
                    taskDiv.className = "task";
                    taskDiv.innerHTML = `
                        <h3>${task.title}</h3>
                        <p><strong>Deadline:</strong> ${new Date(task.deadline).toLocaleString()}</p>
                        <p>${task.description}</p>
                    `;
                    taskList.appendChild(taskDiv);
                });
            })
            .catch((error) => {
                console.error("Error fetching tasks:", error);
                taskList.innerHTML = `<p style="color: red;">Failed to load tasks. Please try again later.</p>`;
            });
    }

    // 新增任務
    taskForm.addEventListener("submit", (event) => {
        event.preventDefault(); // 防止表單默認提交

        const title = document.getElementById("title").value.trim();
        const deadline = document.getElementById("deadline").value;
        const description = document.getElementById("description").value.trim();

        if (!title || !deadline) {
            alert("Title and Deadline are required!");
            return;
        }

        fetch("/api/tasks", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                title: title,
                deadline: new Date(deadline).toISOString(), // 確保日期格式正確
                description: description,
            }),
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then(() => {
                taskForm.reset(); // 清空表單
                loadTasks(); // 重新載入任務列表
            })
            .catch((error) => {
                console.error("Error adding task:", error);
                alert("Failed to add task. Please try again.");
            });
    });

    // 初次載入任務列表
    loadTasks();
});
