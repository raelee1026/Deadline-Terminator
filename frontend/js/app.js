document.addEventListener("DOMContentLoaded", () => {
    const taskList = document.getElementById("task-list");
    const taskForm = document.getElementById("task-form");

    window.loadTasks = function () {
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
                        taskDiv.classList.add("task-deleted");
                    }

                    const descriptionWithBreaks = task.description.replace(/\\n/g, '\n');
                    console.log(descriptionWithBreaks);
                    taskDiv.innerHTML = `
                        <h3>${task.title}</h3>
                        <p><strong>Deadline:</strong> ${new Date(task.deadline).toLocaleString()}</p>
                        <pre>${descriptionWithBreaks}</pre>
                        <button class="delete-btn" data-id="${task.id}">Delete</button>
                    `;

                    if (task.deleted) {
                        taskList.appendChild(taskDiv);
                    } else {
                        taskList.prepend(taskDiv);
                    }
                    taskDiv.querySelector(".delete-btn").addEventListener("click", () => {
                        deleteTask(task.id);
                    });
                });
            })
            .catch((error) => {
                console.error("Error fetching tasks:", error);
                taskList.innerHTML = `<p style="color: #666; text-align: center;">📌 No tasks yet. Click the "+" button to add one!</p>`;
            });
    }

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
                loadTasks();
            })
            .catch((error) => {
                console.error("Error deleting task:", error);
            });
    }


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
                loadTasks();
            })
            .catch((error) => {
                console.error("Error adding task:", error);
                alert("Failed to add task. Please try again.");
            });
    });

    loadTasks();
});
