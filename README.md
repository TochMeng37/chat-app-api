# Golang Fiber Chat Server

This project is a real-time chat application backend built with **Golang**, **Fiber**, **PostgreSQL**, **JWT authentication**, and **WebSocket** support. It allows users to register, login, and chat in real-time (broadcast or private messages).

---

## Features

- User registration and login with **JWT authentication**
- Passwords hashed with **bcrypt**
- PostgreSQL database for users and messages
- Real-time chat using **WebSocket**
- Broadcast messages to all connected users
- Optional private messaging
- Chat history endpoint

---

## Requirements

- Go 1.20+
- PostgreSQL 12+
- Node.js (for frontend/testing, optional)
- `.env` file with environment variables

---

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/golang-fiber-chat.git
cd golang-fiber-chat
