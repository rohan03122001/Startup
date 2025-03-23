// internal/service/game_service.go

package service

import (
	"errors"
	"log"
	"sort"
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

// Add this new struct for final results
type PlayerResult struct {
    PlayerID   string         `json:"player_id"`
    Username   string         `json:"username"`
    TotalScore int           `json:"total_score"`
    Rank       int           `json:"rank"`
    Rounds     []RoundResult `json:"rounds"`
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

    // Add debugging for answer comparison
    submitted := strings.TrimSpace(answer)
    correct := strings.TrimSpace(question.Answer)
    
    log.Printf("Answer comparison - Submitted: '%s' (len: %d), Correct: '%s' (len: %d)", 
        submitted, len(submitted), correct, len(correct))
    
    // Try multiple comparison strategies
    isCorrect := false
    
    // 1. Case-insensitive with trimming (original method)
    if strings.EqualFold(submitted, correct) {
        isCorrect = true
        log.Printf("Answer matched using EqualFold")
    }
    
    // 2. Normalized comparison
    if !isCorrect {
        // Convert to lowercase and trim
        submittedNorm := strings.ToLower(submitted)
        correctNorm := strings.ToLower(correct)
        if submittedNorm == correctNorm {
            isCorrect = true
            log.Printf("Answer matched using lowercase normalization")
        }
    }
    
    // 3. Fuzzy matching (more lenient)
    if !isCorrect {
        // Remove punctuation, extra spaces, and compare
        submittedClean := cleanStringForComparison(submitted)
        correctClean := cleanStringForComparison(correct)
        
        log.Printf("Cleaned for comparison - Submitted: '%s', Correct: '%s'", 
            submittedClean, correctClean)
            
        if submittedClean == correctClean {
            isCorrect = true
            log.Printf("Answer matched using cleaned comparison")
        }
    }
    
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
            // Stop the timer before handling round end
            s.stopRoundTimer(roomCode)
            
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

// Helper function to clean strings for comparison
func cleanStringForComparison(s string) string {
    // Convert to lowercase
    s = strings.ToLower(s)
    
    // Remove punctuation
    punctuation := []string{",", ".", ";", ":", "\"", "'", "!", "?", "(", ")"}
    for _, p := range punctuation {
        s = strings.ReplaceAll(s, p, "")
    }
    
    // Replace multiple spaces with a single space
    for strings.Contains(s, "  ") {
        s = strings.ReplaceAll(s, "  ", " ")
    }
    
    // Trim spaces
    return strings.TrimSpace(s)
}

// Add new method to safely stop timer
func (s *GameService) stopRoundTimer(roomCode string) {
    s.timerMutex.Lock()
    defer s.timerMutex.Unlock()
    
    if timer, exists := s.roundTimers[roomCode]; exists {
        timer.Stop()
        delete(s.roundTimers, roomCode)
        log.Printf("Stopped timer for room %s", roomCode)
    }
}

// Update startRoundTimer to be more robust
func (s *GameService) startRoundTimer(roomCode string, duration int) {
    s.timerMutex.Lock()
    // Stop any existing timer first
    if existingTimer, exists := s.roundTimers[roomCode]; exists {
        existingTimer.Stop()
        delete(s.roundTimers, roomCode)
        log.Printf("Stopped existing timer for room %s", roomCode)
    }
    
    // Create new timer
    timer := time.NewTimer(time.Duration(duration) * time.Second)
    s.roundTimers[roomCode] = timer
    s.timerMutex.Unlock()

    log.Printf("Started new timer for room %s with duration %d seconds", roomCode, duration)

    // Start time update goroutine
    go func() {
        remaining := duration
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-timer.C:
                s.stopRoundTimer(roomCode)  // Ensure timer is cleaned up
                s.handleRoundEnd(roomCode)
                return
            case <-ticker.C:
                s.timerMutex.RLock()
                _, timerExists := s.roundTimers[roomCode]
                s.timerMutex.RUnlock()
                
                if !timerExists {
                    return // Exit if timer was stopped
                }
                
                remaining--
                if remaining >= 0 {
                    s.hub.BroadcastToRoom(roomCode, websocket.GameEvent{
                        Type: "timer_update",
                        Data: map[string]interface{}{
                            "remaining": remaining,
                            "warning":   remaining <= 5,
                        },
                    })
                }
            }
        }
    }()
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

    // Get all rounds for this room
    rounds, err := s.roundRepo.GetRoomRounds(room.ID.String())
    if err != nil {
        log.Printf("Error getting room rounds: %v", err)
        return
    }

    // Get all players in the room
    players := s.hub.GetPlayersInRoom(roomCode)
    playerResults := make(map[string]*PlayerResult)

    // Initialize results for all players
    for _, player := range players {
        playerResults[player["id"]] = &PlayerResult{
            PlayerID:   player["id"],
            Username:   player["username"],
            TotalScore: 0,
            Rounds:     make([]RoundResult, len(rounds)),
        }
    }

    // Process each round
    for i, round := range rounds {
        // Get answers for this round
        answers, err := s.roundRepo.GetRoundAnswers(round.ID.String())
        if err != nil {
            log.Printf("Error getting round answers: %v", err)
            continue
        }

        // Process answers for this round
        answeredPlayers := make(map[string]bool)
        for _, answer := range answers {
            if result, exists := playerResults[answer.PlayerID]; exists {
                result.TotalScore += answer.Score
                result.Rounds[i] = RoundResult{
                    Correct: answer.Score > 0,
                    Score:   answer.Score,
                    Order:   answer.AnswerOrder,
                }
                answeredPlayers[answer.PlayerID] = true
            }
        }

        // Handle players who didn't answer
        for playerID := range playerResults {
            if !answeredPlayers[playerID] {
                playerResults[playerID].Rounds[i] = RoundResult{
                    Correct: false,
                    Score:   0,
                    Order:   0,
                }
            }
        }
    }

    // Convert map to slice and sort by total score
    finalResults := make([]*PlayerResult, 0, len(playerResults))
    for _, result := range playerResults {
        finalResults = append(finalResults, result)
    }

    // Sort by total score (descending)
    sort.Slice(finalResults, func(i, j int) bool {
        return finalResults[i].TotalScore > finalResults[j].TotalScore
    })

    // Assign ranks (handle ties)
    currentRank := 1
    previousScore := -1
    for i, result := range finalResults {
        if result.TotalScore != previousScore {
            currentRank = i + 1
        }
        result.Rank = currentRank
        previousScore = result.TotalScore
    }

    // Update room status
    if err := s.roomRepo.UpdateStatus(room.ID.String(), "finished"); err != nil {
        log.Printf("Error updating room status: %v", err)
    }

    // Cancel any existing timer
    s.timerMutex.Lock()
    if timer, exists := s.roundTimers[roomCode]; exists {
        timer.Stop()
        delete(s.roundTimers, roomCode)
    }
    s.timerMutex.Unlock()

    // Broadcast final results
    s.hub.BroadcastToRoom(roomCode, websocket.GameEvent{
        Type: "game_end",
        Data: map[string]interface{}{
            "final_results": finalResults,
            "total_rounds": len(rounds),
            "room_code":    roomCode,
        },
    })

    log.Printf("Game ended in room %s with %d players", roomCode, len(players))
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

// RestartGame resets the game with the same players
func (s *GameService) RestartGame(roomCode string, settings *models.GameSettings) error {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return errors.New("room not found")
    }

    // Update room settings if provided
    if settings != nil {
        room.MaxRounds = settings.MaxRounds
        room.RoundTime = settings.RoundTime
    }

    // Reset room state
    room.CurrentRound = 0
    room.Status = "waiting"

    if err := s.roomRepo.UpdateRoom(room); err != nil {
        return err
    }

    // Broadcast restart event
    s.hub.BroadcastToRoom(roomCode, websocket.GameEvent{
        Type: "game_restart",
        Data: map[string]interface{}{
            "settings": map[string]interface{}{
                "max_rounds":  room.MaxRounds,
                "round_time": room.RoundTime,
            },
        },
    })

    log.Printf("Game restarted in room %s with %d rounds, %d seconds per round", 
        roomCode, room.MaxRounds, room.RoundTime)
    return nil
}

// GetGameState returns the current game state
func (s *GameService) GetGameState(roomCode string, playerID string) (map[string]interface{}, error) {
    room, err := s.roomRepo.GetByCode(roomCode)
    if err != nil {
        return nil, errors.New("room not found")
    }

    // Get current round if game is in progress
    var currentRound *models.GameRound
    var currentQuestion *models.Question
    if room.Status == "playing" {
        currentRound, err = s.roundRepo.GetCurrentRound(room.ID.String())
        if err == nil && currentRound != nil {
            currentQuestion, _ = s.questionRepo.GetByID(currentRound.QuestionID.String())
        }
    }

    // Get player's answers and scores
    playerAnswers, err := s.roundRepo.GetPlayerAnswers(room.ID.String(), playerID)
    if err != nil {
        log.Printf("Error getting player answers: %v", err)
    }

    // Get all players in room
    players := s.hub.GetPlayersInRoom(roomCode)

    gameState := map[string]interface{}{
        "room_code":     roomCode,
        "game_status":   room.Status,
        "current_round": room.CurrentRound,
        "max_rounds":    room.MaxRounds,
        "round_time":    room.RoundTime,
        "players":       players,
        "your_answers":  playerAnswers,
    }

    // Include current question if game is in progress
    if currentQuestion != nil {
        gameState["current_question"] = currentQuestion
        gameState["round_end_time"] = currentRound.EndTime
    }

    return gameState, nil
}