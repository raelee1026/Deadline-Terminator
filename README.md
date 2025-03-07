# Deadline-Terminator

This is a simple task management application built with a Go backend and an HTML/CSS/JavaScript frontend.

---

## How to Compile and Run

### 1. Clone the Repository

```bash
git clone https://github.com/raelee1026/Deadline-Terminator.git
cd deadline-terminator
```

### 2. Add .env file

- Add .env  `/backend/.env`
- Add the semester. (e.g., ``1131``, ``1132``)
```bash
FILTER_PREFIX=1132
```
### 3. Obtain Google OAuth2 Credentials (credentials.json)

1.Go to Google Cloud Console and create a new project.

2.Enable Gmail API for your project.

3.Create OAuth2 credentials:
  - Go to ``APIs & Services`` → ``Credentials``
  - Click ``Create Credentials`` → ``OAuth Client ID``
  - Choose Web application and add http://localhost:8080/oauth2/callback as a redirect URI.

4.Download the credentials:
  - Click ``Download JSON`` and save it as ``/backend/config/credentials.json``

### 4. Start the Application with Docker

Run the following command to build and start all services:

```bash
docker-compose up --build -d
```

### 5. Authenticate with Google (Required Step)

After starting the containers, you must authenticate with Google to enable Gmail task synchronization.

Follow these steps:

1.Run the following command to check the backend logs:

  ```bash
  docker-compose logs backend
  ```
2.Look for the authorization URL in the logs:

  ```bash
  Please visit the following URL to complete the authorization: 
  ```
3.Click the URL (or copy & paste it into your browser).

### 6. Managing Docker Containers
1.Stop all containers

  ```bash
  docker-compose down
  ```
2.Restart containers (after code changes)

  ```bash
  docker-compose up --build -d
  ````
3.Check running containers

  ```bash
  docker ps
  ```
4.View backend logs (for troubleshooting)

  ```bash
  docker-compose logs backend
  ```
## Demo Video
https://drive.google.com/drive/folders/1uuOu4ukJdv7FZx6AQe0kAuTO3LLZNPfF?usp=sharing