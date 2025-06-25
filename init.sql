-- Initialize database for M-Pesa Tinker application
-- This script will run automatically when the MySQL container starts

USE mpesa;

-- Create the stk_requests table for storing M-Pesa STK push requests
CREATE TABLE IF NOT EXISTS stk_requests (
    id INT AUTO_INCREMENT PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    amount VARCHAR(10) NOT NULL,
    status VARCHAR(50) DEFAULT 'initiated',
    checkout_request_id VARCHAR(255),
    mpesa_receipt_number VARCHAR(100),
    transaction_date BIGINT,
    callback_amount VARCHAR(10),
    result_code INT,
    result_desc TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_phone ON stk_requests(phone);
CREATE INDEX idx_status ON stk_requests(status);
CREATE INDEX idx_checkout_request_id ON stk_requests(checkout_request_id);
CREATE INDEX idx_created_at ON stk_requests(created_at);

-- Insert some sample data for testing (optional)
-- Uncomment the lines below if you want sample data
-- INSERT INTO stk_requests (phone, amount, status) VALUES
-- ('254712345678', '100', 'initiated'),
-- ('254798765432', '250', 'completed'),
-- ('254711223344', '500', 'failed');

-- Display confirmation
SELECT 'Database initialization completed successfully!' as message;
