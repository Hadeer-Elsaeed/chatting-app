-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create messages table for direct and broadcast messages
CREATE TABLE IF NOT EXISTS messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    sender_id INT NOT NULL,
    content TEXT NOT NULL,
    message_type ENUM('direct', 'broadcast') DEFAULT 'direct',
    media_url VARCHAR(500) NULL,
    media_type VARCHAR(50) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_sender_created (sender_id, created_at),
    INDEX idx_created (created_at)
);

-- Create message_recipients table for direct messages and broadcast recipients
CREATE TABLE IF NOT EXISTS message_recipients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    message_id INT NOT NULL,
    recipient_id INT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_message_recipient (message_id, recipient_id),
    INDEX idx_recipient_created (recipient_id, created_at)
);

-- Insert sample users with bcrypt hashed passwords (password: "password123")
INSERT INTO users (username, email, password_hash) VALUES 
('john_doe', 'john@example.com', '$2a$10$rVmq6G7tQXNOJR5Zr5rGH.yDGjLqXQE3RjKx6zXOQ4yKJ5VGQz5Vm'),
('jane_smith', 'jane@example.com', '$2a$10$rVmq6G7tQXNOJR5Zr5rGH.yDGjLqXQE3RjKx6zXOQ4yKJ5VGQz5Vm'),
('bob_wilson', 'bob@example.com', '$2a$10$rVmq6G7tQXNOJR5Zr5rGH.yDGjLqXQE3RjKx6zXOQ4yKJ5VGQz5Vm'),
('alice_brown', 'alice@example.com', '$2a$10$rVmq6G7tQXNOJR5Zr5rGH.yDGjLqXQE3RjKx6zXOQ4yKJ5VGQz5Vm');

-- Insert sample direct messages
INSERT INTO messages (sender_id, content, message_type) VALUES 
(1, 'Hello Jane! How are you doing?', 'direct'),
(2, 'Hi John! I am doing great, thanks for asking!', 'direct'),
(1, 'That is wonderful to hear!', 'direct');

-- Insert sample broadcast message
INSERT INTO messages (sender_id, content, message_type) VALUES 
(3, 'Hello everyone! Welcome to our chat application!', 'broadcast');

-- Insert message recipients for direct messages
INSERT INTO message_recipients (message_id, recipient_id) VALUES 
(1, 2), -- John to Jane
(2, 1), -- Jane to John
(3, 2); -- John to Jane

-- Insert broadcast recipients (to all users except sender)
INSERT INTO message_recipients (message_id, recipient_id) VALUES 
(4, 1), -- Broadcast to John
(4, 2), -- Broadcast to Jane
(4, 4); -- Broadcast to Alice
