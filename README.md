# Final Forum Project

This project is a simple full-stack forum application built with a Go backend and a React frontend.  
It is developed for learning and coursework purposes.

The backend provides RESTful APIs with authentication and permission control.  
The frontend provides a simple user interface built with React and TypeScript.

---

## Project Structure

```
final-forum/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── api/
│   │   ├── database/
│   │   ├── handlers/
│   │   ├── models/
│   │   └── router/
│   ├── schema.sql
│   ├── forum.db
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── api.ts
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── types.ts
│   ├── index.html
│   ├── package.json
│   └── vite.config.ts
│
└── README.md
```

---

## Backend

### Backend Features

- User registration and login
- Session-based authentication using cookies
- Topics, posts, and comments
- Permission control:
  - Users can only edit or delete their own content
- RESTful API design
- SQLite database

### How to Run Backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

The backend server runs at:

```
http://localhost:8080
```

---

## Frontend

### Frontend Features

- Built with React and TypeScript
- Uses Vite as the build tool
- Communicates with backend APIs
- Simple component-based structure
- Supports login and forum operations

### How to Run Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend runs at:

```
http://localhost:5173
```

## Deployed Application

You can access to the application through:

```
https://final-forum.vercel.app/
```

---

## API Overview (Backend)

### Authentication

- POST /auth/register
- POST /auth/login
- POST /auth/logout
- GET  /auth/me

### Topics

- GET    /api/topics
- POST   /api/topics
- PUT    /api/topics/{topicID}
- DELETE /api/topics/{topicID}

### Posts

- GET    /api/topics/{topicID}/posts
- POST   /api/topics/{topicID}/posts
- GET    /api/posts/{postID}
- PUT    /api/posts/{postID}
- DELETE /api/posts/{postID}

### Comments

- GET    /api/posts/{postID}/comments
- POST   /api/posts/{postID}/comments
- PUT    /api/comments/{commentID}
- DELETE /api/comments/{commentID}

---

## Notes

- Backend and frontend run as separate processes
- Frontend communicates with backend via HTTP
- Authentication is handled using cookies
- Non-GET API routes require login
- Only content owners can modify or delete their data

---

## Purpose

This project is intended to practice:

- Go backend development
- REST API design
- Authentication and authorization
- React frontend development
- Full-stack integration

---

## AI Usage Declaration

This project made limited use of AI tools as a learning and support resource.

Specifically, AI assistance was used to:

- Clarify concepts related to Go web development and HTTP request handling.

- Suggest possible code structures and appropriate design patterns.

- Provide guidance on debugging and improving code readability.

- Help refine documentation and written explanations.

All core system design, implementation, and final integration were completed by the author.
AI tools were not used to automatically generate the complete solution, but were used to support understanding, problem-solving, and iterative improvement throughout the development process.
