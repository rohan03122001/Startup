# Quiz Application System Documentation

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture](#architecture)
3. [Component Details](#component-details)
4. [Game Flow](#game-flow)
5. [Technical Implementation](#technical-implementation)
6. [Features and Functions](#features-and-functions)

## Architecture

### Project Structure

```
quiz-app/
├── cmd/api/                 # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── models/             # Data models
│   ├── repository/         # Database operations
│   ├── service/            # Business logic
│   ├── websocket/          # Real-time communication
│   └── handlers/           # HTTP and WebSocket handlers
├── config/                 # Configuration files
```

### Technology Stack

- **Backend**: Go (Golang)
- **Database**: PostgreSQL
- **ORM**: GORM
- **Web Framework**: Gin
- **WebSocket**: Gorilla WebSocket
- **Configuration**: Viper

## Component Details

### 1. Models (`internal/models/`)

Contains the core data structures:

- **Room**: Game room information
- **Question**: Quiz questions and answers
- **GameRound**: Individual round data
- **PlayerAnswer**: Player responses and scoring

### 2. Repository Layer (`internal/repository/`)

Database interaction layer:

- **RoomRepository**: Room CRUD operations
- **QuestionRepository**: Question management
- **GameRoundRepository**: Round tracking
- Each repository implements transaction handling and logging

### 3. Service Layer (`internal/service/`)

Business logic implementation:

- **GameService**: Core game mechanics
  - Round management
  - Answer processing
  - Score calculation
  - Game state management
- **RoomService**: Room operations
  - Room creation
  - Player management
  - Game initialization
- **CleanupService**: Maintenance
  - Inactive room detection
  - Resource cleanup
  - Session management

### 4. WebSocket Layer (`internal/websocket/`)

Real-time communication:

- **Hub**: Central WebSocket manager
  - Client tracking
  - Room management
  - Message broadcasting
- **Client**: Individual connection handler
  - Message pumps
  - Connection lifecycle
  - Error handling

### 5. Handlers (`internal/handlers/`)

Request processing:

- **HTTPHandler**: REST endpoints
- **GameHandler**: WebSocket game events
- **WebSocketHandler**: Connection management

## Game Flow

### 1. Room Creation and Setup

1. Player requests room creation via HTTP
2. System generates unique room code
3. Room initialized in waiting state
4. Creator receives room code

### 2. Player Joining

1. Player connects via WebSocket
2. Sends join room request with code
3. System validates room status
4. Player added to room
5. All room players notified

### 3. Game Start

1. Any player initiates game start
2. System verifies minimum players (2)
3. Game state changes to "playing"
4. First round begins

### 4. Round Flow

1. System selects random question
2. Question broadcast to all players
3. Timer starts (default 30 seconds)
4. Players submit answers
5. System processes answers:
   - Validates correctness
   - Assigns scores based on speed
   - Updates round state
6. Round ends when:
   - All players answer
   - Timer expires
7. Results broadcast to all players

### 5. Scoring System

- First correct answer: 1000 points
- Second correct answer: 750 points
- Third correct answer: 500 points
- Subsequent correct answers: 250 points

### 6. Game Completion

1. All rounds completed
2. Final scores calculated
3. Rankings determined
4. Results broadcast
5. Room state updated
6. Players can choose to play again
