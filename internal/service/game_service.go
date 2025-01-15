// internal/service/game_service.go

package service

import (
	"errors"
	"log"
	"strings"
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
    }
}

func (s *GameService) StartGame(roomCode string) error {
    log.Printf("Starting game for room code: %s", roomCode)
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        log.Printf("Room not found: %s", roomCode)
        return errors.New("room not found")
    }

    currentPlayers := s.hub.GetPlayerCount(room.ID.String())
    log.Printf("Current players in room: %d", currentPlayers)

    if currentPlayers < 2 {
        return errors.New("need at least 2 players to start")
    }

    if err := s.roomRepo.UpdateStatus(room.ID.String(), "playing"); err != nil {
        log.Printf("Failed to update room status: %v", err)
        return err
    }

    // Start first round
    question, err := s.StartRound(roomCode)
    if err != nil {
        return err
    }

    // Broadcast first question
    s.hub.BroadcastToRoom(room.ID.String(), websocket.GameEvent{
        Type: "round_started",
        Data: map[string]interface{}{
            "question": question,
            "round_number": 1,
        },
    })

    return nil
}

func (s *GameService) StartRound(roomCode string) (*models.Question, error) {
    log.Printf("Starting new round for room code: %s", roomCode)
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return nil, errors.New("room not found")
    }

    if room.Status != "playing" {
        return nil, errors.New("game not in progress")
    }

    // Get random question
    question, err := s.questionRepo.GetRandom()
    if err != nil {
        log.Printf("Failed to get random question: %v", err)
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

    // Update room's current round
    if err := s.roomRepo.UpdateCurrentRound(room.ID.String()); err != nil {
        log.Printf("Failed to update current round: %v", err)
        return nil, err
    }

    return question, nil
}

// ProcessAnswer handles a player's answer submission
func (s *GameService) ProcessAnswer(roomID string, playerID string, answer string) (*RoundResult, error) {
    round, err := s.roundRepo.GetCurrentRound(roomID)
    if err != nil {
        return nil, errors.New("no active round found")
    }

    if round.State != "active" {
        return nil, errors.New("round is not active")
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
            playerID, roomID, round.AnswerCount, score)

        return &RoundResult{
            Correct: true,
            Score:   score,
            Order:   round.AnswerCount,
        }, nil
    }

    log.Printf("Player %s submitted incorrect answer in room %s", playerID, roomID)
    return &RoundResult{
        Correct: false,
        Score:   0,
        Order:   0,
    }, nil
}

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

    log.Printf("Ending round in room %s", roomID)
    
    // Update round state
    round.State = "finished"
    if err := s.roundRepo.UpdateRoundState(round.ID.String(), "finished"); err != nil {
        return err
    }

    // Get all answers for the round
    answers, err := s.roundRepo.GetRoundAnswers(round.ID.String())
    if err != nil {
        return err
    }

    // Broadcast results
    s.hub.BroadcastToRoom(roomID, websocket.GameEvent{
        Type: "round_result",
        Data: answers,
    })

    return nil
}

type RoundResult struct {
    Correct bool `json:"correct"`
    Score   int  `json:"score"`
    Order   int  `json:"order"`
}

// Add these functions to internal/service/game_service.go

// InitializeGame sets up a new game
func (s *GameService) InitializeGame(roomID string) error {
    room, err := s.roomRepo.GetByCode(roomID)
    if err != nil {
        log.Printf("Failed to get room %s: %v", roomID, err)
        return errors.New("room not found")
    }

    // Reset room state
    room.CurrentRound = 0
    room.Status = "playing"

    if err := s.roomRepo.UpdateStatus(room.ID.String(), "playing"); err != nil {
        log.Printf("Failed to update room status: %v", err)
        return err
    }

    // Notify all players
    s.hub.BroadcastToRoom(room.ID.String(), websocket.GameEvent{
        Type: "game_initialized",
        Data: map[string]interface{}{
            "max_rounds": room.MaxRounds,
            "round_time": room.RoundTime,
        },
    })

    log.Printf("Game initialized in room %s", roomID)
    return nil
}

// ShouldEndGame checks if the game should end (all rounds completed)
func (s *GameService) ShouldEndGame(roomID string) bool {
    room, err := s.roomRepo.GetByCode(roomID)
    if err != nil {
        log.Printf("Failed to get room %s: %v", roomID, err)
        return false
    }

    // Game ends when current round reaches max rounds
    shouldEnd := room.CurrentRound >= room.MaxRounds
    if shouldEnd {
        log.Printf("Game in room %s should end (round %d of %d)", 
            roomID, room.CurrentRound, room.MaxRounds)
    }

    return shouldEnd
}