-- test_questions.sql

INSERT INTO questions (id, content, answer, created_at) VALUES
    (gen_random_uuid(), 'What is the capital of France?', 'Paris', NOW()),
    (gen_random_uuid(), 'What is 2 + 2?', '4', NOW()),
    (gen_random_uuid(), 'Who painted the Mona Lisa?', 'Leonardo da Vinci', NOW()),
    (gen_random_uuid(), 'What planet is known as the Red Planet?', 'Mars', NOW()),
    (gen_random_uuid(), 'What is the largest mammal in the world?', 'Blue Whale', NOW()),
    (gen_random_uuid(), 'What is the chemical symbol for Gold?', 'Au', NOW()),
    (gen_random_uuid(), 'Which programming language has a snake as its logo?', 'Python', NOW()),
    (gen_random_uuid(), 'What year did World War II end?', '1945', NOW()),
    (gen_random_uuid(), 'What is the square root of 64?', '8', NOW()),
    (gen_random_uuid(), 'Who wrote Romeo and Juliet?', 'William Shakespeare', NOW());