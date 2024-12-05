# Deadline-Terminator

This is a simple task management application built with a Go backend and an HTML/CSS/JavaScript frontend.

---

## How to Compile and Run

### 1. Clone the Repository

```bash
git clone https://github.com/raelee1026/Deadline-Terminator.git
cd deadline-terminator
```

### 2. Compile and Run the Server

Run the following command to start the backend server:

```bash
go run main.go
```

The server will start at `http://localhost:8080`.

### 3. Access the Application

Open your browser and go to:

```plaintext
http://localhost:8080
```
You can now use the application.

### 4. Authenticate with Google
After running main.go, open your browser and go to the following URL to authenticate with your NYCU email:
```bash
https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=997285622302-goltvajj196rm1ims0sijhgbvro82cad.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foauth2%2Fcallback&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.readonly&state=state-token
```
---
## Crawler
### Requirements
`pip install requests`

`pip install python-dotenv mysql-connector-python`

### Run
`python main.py 1131`

- 1131 -> 113學年 第一學期    
- 112X -> 112學年 暑期
