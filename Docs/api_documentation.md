# Quiz Application API Documentation

## Table of Contents

1. [HTTP Endpoints](#http-endpoints)
2. [WebSocket Events](#websocket-events)
3. [Data Models](#data-models)
4. [Error Handling](#error-handling)

## HTTP Endpoints

### Create Room

Creates a new quiz room and returns the room code.

**Endpoint:** `POST /api/rooms`

**Response:**

```json
{
  "room_code": "ABC123",
  "max_players": 10,
  "round_time": 30
}
```

**Status Codes:**

- 201: Room created successfully
- 500: Internal server error

### Get Active Rooms

Retrieves a list of all active rooms waiting for players.

**Endpoint:** `GET /api/rooms`

**Response:**

```json
[
  {
    "id": "uuid",
    "code": "ABC123",
    "status": "waiting",
    "max_players": 10,
    "round_time": 30,
    "max_rounds": 5,
    "current_round": 0,
    "created_at": "2024-01-16T10:00:00Z"
  }
]
```

**Status Codes:**

- 200: Success
- 500: Internal server error

## WebSocket Events

### Connection

**Endpoint:** `ws://localhost:8080/ws`

### Client -> Server Events

#### 1. Join Room

Sent when a player wants to join a room.

```json
{
  "type": "join_room",
  "data": {
    "room_code": "ABC123",
    "username": "Player1"
  }
}
```

#### 2. Start Game

Sent by any player to start the game (requires minimum 2 players).

```json
{
  "type": "start_game"
}
```

#### 3. Submit Answer

Sent when a player submits their answer.

```json
{
  "type": "submit_answer",
  "data": {
    "answer": "answer text"
  }
}
```

#### 4. Play Again

Sent to restart the game with the same players.

```json
{
  "type": "play_again",
  "data": {
    "max_rounds": 5,
    "round_time": 30
  }
}
```

#### 5. Reconnect

Sent when a player tries to reconnect to an existing game.

```json
{
  "type": "reconnect",
  "data": {
    "room_code": "ABC123",
    "player_id": "uuid",
    "username": "Player1"
  }
}
```

### Server -> Client Events

#### 1. Player Joined

```json
{
  "type": "player_joined",
  "data": {
    "player_id": "uuid",
    "username": "Player1",
    "total_players": 2
  }
}
```

#### 2. Room Joined

```json
{
  "type": "room_joined",
  "data": {
    "room_code": "ABC123",
    "players": [
      {
        "id": "uuid",
        "username": "Player1"
      }
    ],
    "settings": {
      "max_players": 10,
      "round_time": 30,
      "max_rounds": 5
    }
  }
}
```

#### 3. Round Started

```json
{
  "type": "round_started",
  "data": {
    "question": {
      "id": "uuid",
      "content": "Question text"
    },
    "round_number": 1,
    "time_limit": 30
  }
}
```

#### 4. Timer Update

```json
{
  "type": "timer_update",
  "data": {
    "remaining": 25,
    "warning": false
  }
}
```

#### 5. Answer Result

```json
{
  "type": "answer_result",
  "data": {
    "correct": true,
    "score": 1000,
    "order": 1
  }
}
```

#### 6. Round Result

```json
{
  "type": "round_result",
  "data": {
    "round_number": 1,
    "answers": [
      {
        "player_id": "uuid",
        "answer": "answer text",
        "score": 1000,
        "answer_order": 1
      }
    ],
    "question": {
      "id": "uuid",
      "content": "Question text"
    },
    "correct_answer": "correct answer"
  }
}
```

#### 7. Game End

```json
{
  "type": "game_end",
  "data": {
    "final_results": [
      {
        "player_id": "uuid",
        "username": "Player1",
        "total_score": 3500,
        "rank": 1,
        "rounds": [
          {
            "correct": true,
            "score": 1000,
            "order": 1
          }
        ]
      }
    ],
    "total_rounds": 5,
    "room_code": "ABC123"
  }
}
```

## Data Models

### Room

```sql
CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    code VARCHAR(6) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL,
    max_players INT DEFAULT 10,
    round_time INT DEFAULT 30,
    max_rounds INT DEFAULT 5,
    current_round INT DEFAULT 0,
    created_at TIMESTAMP,
    ended_at TIMESTAMP,
    last_activity TIMESTAMP NOT NULL
);
```

### Question

```sql
CREATE TABLE questions (
    id UUID PRIMARY KEY,
    content TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMP
);
```

### GameRound

```sql
CREATE TABLE game_rounds (
    id UUID PRIMARY KEY,
    room_id UUID REFERENCES rooms(id),
    question_id UUID REFERENCES questions(id),
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    round_number INT NOT NULL,
    state VARCHAR(20) NOT NULL,
    answer_count INT DEFAULT 0
);
```

### PlayerAnswer

```sql
CREATE TABLE player_answers (
    id UUID PRIMARY KEY,
    round_id UUID REFERENCES game_rounds(id),
    player_id VARCHAR NOT NULL,
    answer TEXT NOT NULL,
    score INT DEFAULT 0,
    answer_order INT NOT NULL,
    answered_at TIMESTAMP
);
```

## Error Handling

### Common Error Responses

```json
{
  "type": "error",
  "data": {
    "message": "Error description"
  }
}
```

### Common Error Types

1. Room not found
2. Game already in progress
3. Room is full
4. Need at least 2 players to start
5. Round not active
6. Invalid answer format
7. No active round
8. Question not found

### HTTP Status Codes

- 200: Success
- 201: Created
- 400: Bad Request
- 404: Not Found
- 500: Internal Server Error
