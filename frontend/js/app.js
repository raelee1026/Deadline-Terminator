document.addEventListener("DOMContentLoaded", () => {
    const taskList = document.getElementById("task-list");
    const taskForm = document.getElementById("task-form");

    // 加載任務列表
    function loadTasks() {
        fetch("/api/tasks")
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then((tasks) => {
                taskList.innerHTML = "";
                tasks.forEach((task) => {
                    const taskDiv = document.createElement("div");
                    taskDiv.className = "task";
                    if (task.deleted) {
                        taskDiv.classList.add("task-deleted"); // 添加刪除樣式
                    }

                    //html, css
                    taskDiv.innerHTML = `
                        <h3>${task.title}</h3>
                        <p><strong>Deadline:</strong> ${new Date(task.deadline).toLocaleString()}</p>
                        <p>${task.description}</p>
                        <button class="delete-btn" data-id="${task.id}">Delete</button>
                    `;

                    // 刪除的任務放到最下面
                    if (task.deleted) {
                        taskList.appendChild(taskDiv);
                    } else {
                        taskList.prepend(taskDiv);
                    }

                    // 添加刪除按鈕的事件監聽器
                    taskDiv.querySelector(".delete-btn").addEventListener("click", () => {
                        deleteTask(task.id);
                    });
                });
            })
            .catch((error) => {
                console.error("Error fetching tasks:", error);
                taskList.innerHTML = `<p style="color: red;">Failed to load tasks. Please try again later.</p>`;
            });
    }

    // 刪除任務
    function deleteTask(taskId) {
        fetch("/api/tasks/delete", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ id: taskId }),
        })
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                loadTasks(); // 刷新任務列表
            })
            .catch((error) => {
                console.error("Error deleting task:", error);
            });
    }

    // 提交新任務
    taskForm.addEventListener("submit", (event) => {
        event.preventDefault();

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
                deadline: new Date(deadline).toISOString(),
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
                taskForm.reset();
                loadTasks(); // 重新加載任務列表
            })
            .catch((error) => {
                console.error("Error adding task:", error);
                alert("Failed to add task. Please try again.");
            });
    });

    // 初始化加載任務列表
    loadTasks();
});
