CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50),
    email VARCHAR(255),
    ssn VARCHAR(20)
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    card_number VARCHAR(20),
    notes TEXT
);

-- Tag some columns correctly
COMMENT ON COLUMN users.email IS '[hornfels: pii=true] User email address';
COMMENT ON COLUMN payments.card_number IS '[hornfels: pii=true]';

-- Tag one incorrectly as false (to trigger the data scanner)
COMMENT ON COLUMN users.ssn IS '[hornfels: pii=false] Not an SSN';

-- Insert fake data to trigger the heuristic engine
INSERT INTO users (first_name, email, ssn) VALUES 
('Alice', 'alice@example.com', '123-45-6789'),
('Bob', 'bob@example.com', '987-65-4321');

-- 4111111111111111 is a valid mod-10 Luhn check
INSERT INTO payments (card_number, notes) VALUES 
('4111 1111 1111 1111', 'Payment for services');
