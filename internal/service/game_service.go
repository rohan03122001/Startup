// internal/service/game_service.go

package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rohan03122001/quizzing/internal/models"
	"github.com/rohan03122001/quizzing/internal/repository"
	"github.com/rohan03122001/quizzing/internal/websockets"
)

type GameService struct {
    roomRepo      *repository.RoomRepository
    questionRepo  *repository.QuestionRepository
    roundRepo     *repository.GameRoundRepository
    hub           *websockets.Hub
}

func NewGameService(
    roomRepo *repository.RoomRepository,
    questionRepo *repository.QuestionRepository,
    roundRepo *repository.GameRoundRepository,
    hub *websockets.Hub,
) *GameService {
    return &GameService{
        roomRepo:     roomRepo,
        questionRepo: questionRepo,
        roundRepo:    roundRepo,
        hub:         hub,
    }
}

// InitializeGame sets up a new game
func (s *GameService) InitializeGame(roomID string) error {
    // Reset room state
    room, err := s.roomRepo.GetRoomByCode(roomID)
    if err != nil {
        return err
    }

    room.CurrentRound = 0
    room.Status = "playing"

    if err := s.roomRepo.UpdateStatus(room.ID.String(), "playing"); err != nil {
        return err
    }

    // Notify all players
    s.hub.BroadcastToRoom(room.ID.String(), websockets.GameEvent{
        Type: "game_initialized",
        Data: map[string]interface{}{
            "max_rounds": room.MaxRounds,
            "round_time": room.RoundTime,
        },
    })

    return nil
}

// StartRound begins a new round for a room
func (s *GameService) StartRound(roomID string) (*models.Question, error) {
    room, err := s.roomRepo.GetRoomByCode(roomID)
    if err != nil {
        return nil, errors.New("room not found")
    }

    if room.Status != "playing" {
        return nil, errors.New("game not in progress")
    }

    // Get random question
    question, err := s.questionRepo.GetRandom()
    if err != nil {
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
        return nil, err
    }

    // Increment room's current round
    if err := s.roomRepo.IncrementRound(room.ID.String()); err != nil {
        return nil, err
    }

    return question, nil
}

// ProcessAnswer handles a player's answer submission
func (s *GameService) ProcessAnswer(roomID string, userID string, answer string) (*RoundResult, error) {
    round, err := s.roundRepo.GetCurrentRound(roomID)
    if err != nil {
        return nil, errors.New("no active round found")
    }

    if round.State != "active" {
        return nil, errors.New("round is not active")
    }

    // Get current question
    question, err := s.questionRepo.GetByID(round.QuestionID.String())
    if err != nil {
        return nil, errors.New("question not found")
    }

    // Check answer
    isCorrect := strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(question.Answer))
    
    if isCorrect {
        // Increment answer count for scoring
        if err := s.roundRepo.UpdateAnswerCount(round.ID.String()); err != nil {
            return nil, err
        }
        round.AnswerCount++

        // Calculate score based on answer order
        score := s.calculateScore(round.AnswerCount)

        // Save answer
        playerAnswer := &models.PlayerAnswer{
            RoundID:     round.ID,
            UserID:      uuid.MustParse(userID),
            Answer:      answer,
            Score:       score,
            AnswerOrder: round.AnswerCount,
            AnsweredAt:  time.Now(),
        }

        if err := s.roundRepo.SaveAnswer(playerAnswer); err != nil {
            return nil, err
        }

        return &RoundResult{
            Correct: true,
            Score:   score,
            Order:   round.AnswerCount,
        }, nil
    }

    return &RoundResult{
        Correct: false,
        Score:   0,
        Order:   0,
    }, nil
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

// EndRound finishes the current round
func (s *GameService) EndRound(roomID string) error {
    round, err := s.roundRepo.GetCurrentRound(roomID)
    if err != nil {
        return err
    }

    round.State = "finished"
    round.EndTime = time.Now()

    if err := s.roundRepo.UpdateRound(round); err != nil {
        return err
    }

    // Get and broadcast results
    answers, err := s.roundRepo.GetRoundAnswers(round.ID.String())
    if err != nil {
        return err
    }

    s.hub.BroadcastToRoom(roomID, websockets.GameEvent{
        Type: "round_result",
        Data: answers,
    })

    return nil
}

//ERROR
// ShouldEndRound determines if current round should end
func (s *GameService) ShouldEndRound(roomID string) bool {
    _, err := s.roomRepo.GetRoomByCode(roomID)
    if err != nil {
        return false
    }

    round, err := s.roundRepo.GetCurrentRound(roomID)
    if err != nil {
        return false
    }

    playerCount := s.hub.GetPlayerCount(roomID)
    timeExpired := time.Now().After(round.EndTime)
    allAnswered := round.AnswerCount >= playerCount

    return timeExpired || allAnswered
}

// ShouldEndGame checks if the game should end
func (s *GameService) ShouldEndGame(roomID string) bool {
    room, err := s.roomRepo.GetRoomByCode(roomID)
    if err != nil {
        return false
    }

    return room.CurrentRound >= room.MaxRounds
}

type RoundResult struct {
    Correct bool `json:"correct"`
    Score   int  `json:"score"`
    Order   int  `json:"order"`
}

