# Quiz Game WebSocket Test Script
# Requirements: 
# - Install wscat: npm install -g wscat
# - PowerShell v5.1+

# Config
$apiUrl = "http://localhost:8080/api"
$wsUrl = "ws://localhost:8080/ws"
$tempFile = ".\ws_temp.json"

Write-Host "======= QUIZ GAME WEBSOCKET TEST =======" -ForegroundColor Cyan

# Helper function to display messages
function Write-Step {
    param($message)
    Write-Host "`n>>> $message" -ForegroundColor Yellow
}

# Helper function to save message to temp file for wscat
function Send-WSMessage {
    param($message)
    $message | Out-File -FilePath $tempFile -Encoding utf8
    Get-Content $tempFile | wscat -c $wsUrl
}

# Step 1: Test creating a room via HTTP
Write-Step "Creating a quiz room via HTTP API"
try {
    $response = Invoke-RestMethod -Method POST -Uri "$apiUrl/rooms"
    $roomCode = $response.code
    Write-Host "Room created with code: $roomCode" -ForegroundColor Green
} catch {
    Write-Host "Failed to create room: $_" -ForegroundColor Red
    exit
}

# Step 2: Test joining as Player 1
Write-Step "Joining as Player 1 (in new terminal window)"
Write-Host "In a new terminal window, run:"
$joinMsg1 = @{
    type = "join_room"
    data = @{
        room_code = $roomCode
        username = "Player1"
    }
} | ConvertTo-Json

Write-Host "wscat -c $wsUrl" -ForegroundColor Magenta
Write-Host "Then paste this message:" -ForegroundColor Magenta
Write-Host $joinMsg1 -ForegroundColor Cyan
Write-Host "`nPress Enter after Player 1 has joined..."
Read-Host

# Step 3: Test joining as Player 2
Write-Step "Joining as Player 2 (in another terminal window)"
Write-Host "In another terminal window, run:"
$joinMsg2 = @{
    type = "join_room"
    data = @{
        room_code = $roomCode
        username = "Player2"
    }
} | ConvertTo-Json

Write-Host "wscat -c $wsUrl" -ForegroundColor Magenta
Write-Host "Then paste this message:" -ForegroundColor Magenta
Write-Host $joinMsg2 -ForegroundColor Cyan
Write-Host "`nPress Enter after Player 2 has joined..."
Read-Host

# Step 4: Start the game from Player 1
Write-Step "Starting the game (send from Player 1's connection)"
$startMsg = @{
    type = "start_game"
} | ConvertTo-Json

Write-Host "In Player 1's terminal, paste this message:" -ForegroundColor Magenta
Write-Host $startMsg -ForegroundColor Cyan
Write-Host "`nPress Enter after starting the game..."
Read-Host

# Step 5: Submit answers for both players
Write-Step "Submitting answers"
$answer1 = @{
    type = "submit_answer"
    data = @{
        answer = "Player 1's answer"
    }
} | ConvertTo-Json

$answer2 = @{
    type = "submit_answer"
    data = @{
        answer = "Player 2's answer"
    }
} | ConvertTo-Json

Write-Host "In Player 1's terminal, paste this message:" -ForegroundColor Magenta
Write-Host $answer1 -ForegroundColor Cyan
Write-Host "`nIn Player 2's terminal, paste this message:" -ForegroundColor Magenta
Write-Host $answer2 -ForegroundColor Cyan
Write-Host "`nPress Enter after submitting answers..."
Read-Host

# Step 6: Play Again after game completes
Write-Step "After game completes, test 'Play Again' feature"
$playAgainMsg = @{
    type = "play_again"
    data = @{
        max_rounds = 3
        round_time = 20
    }
} | ConvertTo-Json

Write-Host "After the game ends, in Player 1's terminal, paste this message:" -ForegroundColor Magenta
Write-Host $playAgainMsg -ForegroundColor Cyan
Write-Host "`nPress Enter after starting a new game..."
Read-Host

# Step 7: Test disconnection and reconnection
Write-Step "Testing disconnection and reconnection"
Write-Host "1. Close Player 1's terminal (Ctrl+C in wscat)" -ForegroundColor Magenta
Write-Host "2. Open a new terminal and run: wscat -c $wsUrl" -ForegroundColor Magenta

$reconnectMsg = @{
    type = "reconnect"
    data = @{
        room_code = $roomCode
        player_id = "<PLAYER_ID_FROM_EARLIER_CONNECTION>"
        username = "Player1"
    }
} | ConvertTo-Json

Write-Host "3. Then paste this message (update the player_id):" -ForegroundColor Magenta
Write-Host $reconnectMsg -ForegroundColor Cyan
Write-Host "`nPress Enter after testing reconnection..."
Read-Host

# Step 8: Test room cleanup
Write-Step "Testing room cleanup"
Write-Host "To test room cleanup:"
Write-Host "1. Create a new room" -ForegroundColor Magenta
Write-Host "2. Join with at least one player" -ForegroundColor Magenta
Write-Host "3. Disconnect all players" -ForegroundColor Magenta
Write-Host "4. Wait 10+ minutes (cleanup interval)" -ForegroundColor Magenta
Write-Host "5. Try to join the room again - should receive 'room not found' error" -ForegroundColor Magenta

# Cleanup
if (Test-Path $tempFile) {
    Remove-Item $tempFile
}

Write-Host "`n======= TEST COMPLETE =======" -ForegroundColor Green
Write-Host "Room code used: $roomCode" 