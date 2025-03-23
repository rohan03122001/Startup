# Test Create Room API (POST /api/rooms)
Write-Host "Testing Create Room API..."
try {
    $createResponse = Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/rooms"
    Write-Host "Create Room Response:"
    $createResponse | ConvertTo-Json
    $roomCode = $createResponse.code
    Write-Host "Room created with code: $roomCode"
    
    # Test Get Active Rooms API (GET /api/rooms)
    Write-Host "`nTesting Get Active Rooms API..."
    $roomsResponse = Invoke-RestMethod -Method GET -Uri "http://localhost:8080/api/rooms"
    Write-Host "Active Rooms Response:"
    $roomsResponse | ConvertTo-Json
    
    # Test Validate Room API (POST /api/rooms/validate)
    if ($roomCode) {
        Write-Host "`nTesting Validate Room API..."
        $validateBody = @{
            room_code = $roomCode
        } | ConvertTo-Json
        
        $validateResponse = Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/rooms/validate" `
            -Body $validateBody -ContentType "application/json"
        Write-Host "Validate Room Response:"
        $validateResponse | ConvertTo-Json
    }
} catch {
    Write-Host "Error testing API: $_"
    Write-Host "Status Code: $($_.Exception.Response.StatusCode.value__)"
    Write-Host "Response Body: $($_.ErrorDetails.Message)"
} 