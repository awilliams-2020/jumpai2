# Advisor Scheduling Platform

A Calendly-like scheduling tool for advisors to meet with their clients, built with React, Golang, and PostgreSQL.

## Features

- Google Calendar integration with OAuth
- Multiple Google Calendar account support
- HubSpot CRM integration
- Customizable scheduling windows
- Configurable scheduling links with:
  - Usage limits
  - Expiration dates
  - Custom forms
  - Meeting duration settings
  - Advance scheduling limits
- LinkedIn profile scraping
- AI-powered context augmentation for meeting notes
- Email notifications

## Tech Stack

- Frontend: React with TypeScript
- Backend: Golang
- Database: PostgreSQL
- Authentication: Google OAuth
- External Integrations: Google Calendar API, HubSpot API, LinkedIn

## Project Structure

```
.
├── frontend/           # React frontend application
├── backend/           # Golang backend application
├── docker/           # Docker configuration files
└── docs/             # Documentation
```

## Setup Instructions

### Prerequisites

- Node.js 18+
- Go 1.21+
- PostgreSQL 15+
- Docker and Docker Compose

### Development Setup

1. Clone the repository
2. Set up environment variables:
   - Copy `.env.example` to `.env` in both frontend and backend directories
   - Fill in the required API keys and configuration

3. Start the development environment:
   ```bash
   docker-compose up -d
   ```

4. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   ```

5. Install backend dependencies:
   ```bash
   cd backend
   go mod download
   ```

6. Start the development servers:
   - Frontend: `npm run dev`
   - Backend: `go run main.go`

## API Documentation

API documentation is available at `/api/docs` when running the backend server.

## License

MIT 