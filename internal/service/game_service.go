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

    // Create new round explicitly
    round := &models.GameRound{
        RoomID:      room.ID,
        QuestionID:  question.ID,
        StartTime:   time.Now(),
        EndTime:     time.Now().Add(time.Duration(room.RoundTime) * time.Second),
        RoundNumber: room.CurrentRound + 1,
        State:      "active",
    }

    // Save round to database
    if err := s.roundRepo.CreateRound(round); err != nil {
        log.Printf("Failed to create round: %v", err)
        return nil, errors.New("failed to create round")
    }

    // Update room's current round
    if err := s.roomRepo.UpdateCurrentRound(room.ID.String()); err != nil {
        log.Printf("Failed to update current round: %v", err)
        return nil, errors.New("failed to update room")
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
        log.Printf("Failed to get current round for room %s: %v", roomCode, err)
        return nil, errors.New("no active round")
    }

    if round.State != "active" {
        return nil, errors.New("round not active")
    }

    // Get the question
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

        // Calculate score based on answer order
        score := s.calculateScore(round.AnswerCount + 1)  // +1 because we just incremented

        // Save answer
        playerAnswer := &models.PlayerAnswer{
            RoundID:     round.ID,
            PlayerID:    playerID,
            Answer:      answer,
            Score:       score,
            AnswerOrder: round.AnswerCount + 1,
            AnsweredAt:  time.Now(),
        }

        if err := s.roundRepo.SaveAnswer(playerAnswer); err != nil {
            return nil, err
        }

        log.Printf("Player %s submitted correct answer (order: %d, score: %d)", 
            playerID, round.AnswerCount+1, score)

        return &RoundResult{
            Correct: true,
            Score:   score,
            Order:   round.AnswerCount + 1,
        }, nil
    }

    // Save incorrect answer with 0 score
    playerAnswer := &models.PlayerAnswer{
        RoundID:     round.ID,
        PlayerID:    playerID,
        Answer:      answer,
        Score:       0,
        AnswerOrder: 0,
        AnsweredAt:  time.Now(),
    }
    s.roundRepo.SaveAnswer(playerAnswer)

    log.Printf("Player %s submitted incorrect answer", playerID)
    return &RoundResult{Correct: false}, nil
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

type RoundResult struct {
    Correct bool `json:"correct"`
    Score   int  `json:"score"`
    Order   int  `json:"order"`
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