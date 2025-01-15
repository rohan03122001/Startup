// internal/service/game_service.go

package service

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/rohan03122001/quizzing/internal/models"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/websocket"
)

type GameService struct {
    roomRepo     *repository.RoomRepository
    questionRepo *repository.QuestionRepository
    roundRepo    *repository.GameRoundRepository
    hub          *websocket.Hub
    roundTimers  map[string]*time.Timer  // tracks room timers
    timerMutex   sync.RWMutex           // protects roundTimers map
}

type RoundResult struct {
    Correct bool `json:"correct"`
    Score   int  `json:"score"`
    Order   int  `json:"order"`
}

func NewGameService(
    roomRepo *repository.RoomRepository,
    questionRepo *repository.QuestionRepository,
    roundRepo *repository.GameRoundRepository,
    hub *websocket.Hub,
) *GameService {
    return &GameService{
        roomRepo:     roomRepo,
        questionRepo: questionRepo,
        roundRepo:    roundRepo,
        hub:         hub,
        roundTimers: make(map[string]*time.Timer),
    }
}

// InitializeGame sets up a new game
func (s *GameService) InitializeGame(roomCode string) error {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        log.Printf("Failed to get room %s: %v", roomCode, err)
        return errors.New("room not found")
    }

    room.CurrentRound = 0
    room.Status = "playing"

    if err := s.roomRepo.UpdateStatus(room.ID.String(), "playing"); err != nil {
        log.Printf("Failed to update room status: %v", err)
        return err
    }

    log.Printf("Game initialized in room %s", roomCode)
    return nil
}

// StartRound begins a new round for a room
func (s *GameService) StartRound(roomCode string) (*models.Question, error) {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        log.Printf("Failed to get room %s: %v", roomCode, err)
        return nil, errors.New("room not found")
    }

    if room.Status != "playing" {
        return nil, errors.New("game not in progress")
    }

    // Get random question
    question, err := s.questionRepo.GetRandom()
    if err != nil {
        log.Printf("Failed to get question: %v", err)
        return nil, errors.New("failed to get question")
    }

    // Create new round
    round := &models.GameRound{
        RoomID:      room.ID,
        QuestionID:  question.ID,
        StartTime:   time.Now(),
        EndTime:     time.Now().Add(time.Duration(room.RoundTime) * time.Second),
        RoundNumber: room.CurrentRound + 1,
        State:       "active",
    }

    if err := s.roundRepo.CreateRound(round); err != nil {
        log.Printf("Failed to create round: %v", err)
        return nil, err
    }

    // Start round timer
    s.startRoundTimer(roomCode, room.RoundTime)

    // Update room's current round
    if err := s.roomRepo.UpdateCurrentRound(room.ID.String()); err != nil {
        log.Printf("Failed to update current round: %v", err)
        return nil, err
    }

    log.Printf("Started round %d in room %s with question ID %s", 
        round.RoundNumber, roomCode, question.ID)
    return question, nil
}

// ProcessAnswer handles a player's answer submission
func (s *GameService) ProcessAnswer(roomCode string, playerID string, answer string) (*RoundResult, error) {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return nil, errors.New("room not found")
    }

    round, err := s.roundRepo.GetCurrentRound(room.ID.String())
    if err != nil {
        log.Printf("No active round found for room %s: %v", roomCode, err)
        return nil, errors.New("no active round")
    }

    if round.State != "active" {
        return nil, errors.New("round not active")
    }

    question, err := s.questionRepo.GetByID(round.QuestionID.String())
    if err != nil {
        return nil, errors.New("question not found")
    }

    // Check answer (case-insensitive)
    isCorrect := strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(question.Answer))
    
    if isCorrect {
        // Increment answer count
        if err := s.roundRepo.UpdateAnswerCount(round.ID.String()); err != nil {
            return nil, err
        }
        round.AnswerCount++

        // Calculate score based on answer order
        score := s.calculateScore(round.AnswerCount)

        // Save answer
        playerAnswer := &models.PlayerAnswer{
            RoundID:     round.ID,
            PlayerID:    playerID,
            Answer:      answer,
            Score:       score,
            AnswerOrder: round.AnswerCount,
            AnsweredAt:  time.Now(),
        }

        if err := s.roundRepo.SaveAnswer(playerAnswer); err != nil {
            return nil, err
        }

        log.Printf("Player %s submitted correct answer in room %s (order: %d, score: %d)", 
            playerID, roomCode, round.AnswerCount, score)

        // Check if all players have answered
        playerCount := s.hub.GetPlayerCount(roomCode)
        if round.AnswerCount >= playerCount {
            // Cancel timer as everyone has answered
            s.timerMutex.Lock()
            if timer, exists := s.roundTimers[roomCode]; exists {
                timer.Stop()
                delete(s.roundTimers, roomCode)
            }
            s.timerMutex.Unlock()
            
            // Handle round end
            go s.handleRoundEnd(roomCode)
        }

        return &RoundResult{
            Correct: true,
            Score:   score,
            Order:   round.AnswerCount,
        }, nil
    }

    // Log incorrect answer
    log.Printf("Player %s submitted incorrect answer in room %s", playerID, roomCode)

    // Save incorrect answer
    playerAnswer := &models.PlayerAnswer{
        RoundID:     round.ID,
        PlayerID:    playerID,
        Answer:      answer,
        Score:       0,
        AnswerOrder: 0,
        AnsweredAt:  time.Now(),
    }
    s.roundRepo.SaveAnswer(playerAnswer)

    return &RoundResult{
        Correct: false,
        Score:   0,
        Order:   0,
    }, nil
}

// startRoundTimer starts the timer for a round
func (s *GameService) startRoundTimer(roomCode string, duration int) {
    s.timerMutex.Lock()
    // Cancel existing timer if any
    if timer, exists := s.roundTimers[roomCode]; exists {
        timer.Stop()
    }
    
    // Create new timer
    timer := time.NewTimer(time.Duration(duration) * time.Second)
    s.roundTimers[roomCode] = timer
    s.timerMutex.Unlock()

    // Handle timer expiration
    go func() {
        <-timer.C
        s.handleRoundEnd(roomCode)
    }()

    log.Printf("Started %d second timer for room %s", duration, roomCode)
}

// handleRoundEnd processes the end of a round
func (s *GameService) handleRoundEnd(roomCode string) {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        log.Printf("Error getting room %s: %v", roomCode, err)
        return
    }

    round, err := s.roundRepo.GetCurrentRound(room.ID.String())
    if err != nil {
        log.Printf("Error getting current round: %v", err)
        return
    }

    // Get round results
    answers, err := s.roundRepo.GetRoundAnswers(round.ID.String())
    if err != nil {
        log.Printf("Error getting round answers: %v", err)
        return
    }

    // Update round state
    if err := s.roundRepo.UpdateRoundState(round.ID.String(), "finished"); err != nil {
        log.Printf("Error updating round state: %v", err)
        return
    }

    // Get question for results
    question, _ := s.questionRepo.GetByID(round.QuestionID.String())

    // Broadcast round results
    s.hub.BroadcastToRoom(roomCode, websocket.GameEvent{
        Type: "round_result",
        Data: map[string]interface{}{
            "round_number":    round.RoundNumber,
            "answers":         answers,
            "question":        question,
            "correct_answer":  question.Answer,
        },
    })

    log.Printf("Round %d ended in room %s", round.RoundNumber, roomCode)

    // Check if game should end
    if round.RoundNumber >= room.MaxRounds {
        s.endGame(roomCode)
        return
    }

    // Start next round after delay
    time.Sleep(5 * time.Second)
    question, err = s.StartRound(roomCode)
    if err != nil {
        log.Printf("Error starting next round: %v", err)
        return
    }

    s.hub.BroadcastToRoom(roomCode, websocket.GameEvent{
        Type: "round_started",
        Data: map[string]interface{}{
            "question":     question,
            "round_number": round.RoundNumber + 1,
            "time_limit":   room.RoundTime,
        },
    })
}

// calculateScore determines points based on answer order
func (s *GameService) calculateScore(answerOrder int) int {
    switch answerOrder {
    case 1:
        return 1000 // First correct answer
    case 2:
        return 750  // Second correct answer
    case 3:
        return 500  // Third correct answer
    default:
        return 250  // All subsequent correct answers
    }
}

// endGame handles game completion
func (s *GameService) endGame(roomCode string) {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        log.Printf("Error getting room for game end: %v", err)
        return
    }

    if err := s.roomRepo.UpdateStatus(room.ID.String(), "finished"); err != nil {
        log.Printf("Error updating room status: %v", err)
        return
    }

    // Cancel any existing timer
    s.timerMutex.Lock()
    if timer, exists := s.roundTimers[roomCode]; exists {
        timer.Stop()
        delete(s.roundTimers, roomCode)
    }
    s.timerMutex.Unlock()

    log.Printf("Game ended in room %s", roomCode)

    s.hub.BroadcastToRoom(roomCode, websocket.GameEvent{
        Type: "game_end",
        Data: map[string]interface{}{
            "message": "Game completed!",
        },
    })
}

// ShouldEndRound determines if the round should end
func (s *GameService) ShouldEndRound(roomID string) bool {
    round, err := s.roundRepo.GetCurrentRound(roomID)
    if err != nil {
        return false
    }

    playerCount := s.hub.GetPlayerCount(roomID)
    timeExpired := time.Now().After(round.EndTime)
    allAnswered := round.AnswerCount >= playerCount

    return timeExpired || allAnswered
}

// EndRound finishes the current round
func (s *GameService) EndRound(roomID string) error {
    round, err := s.roundRepo.GetCurrentRound(roomID)
    if err != nil {
        return err
    }

    if err := s.roundRepo.UpdateRoundState(round.ID.String(), "finished"); err != nil {
        return err
    }

    // Get and broadcast results
    answers, err := s.roundRepo.GetRoundAnswers(round.ID.String())
    if err != nil {
        return err
    }

    s.hub.BroadcastToRoom(roomID, websocket.GameEvent{
        Type: "round_result",
        Data: answers,
    })

    log.Printf("Ended round in room %s", roomID)
    return nil
}