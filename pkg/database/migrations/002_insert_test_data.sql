-- Insert test messages
INSERT INTO messages ("to", content) VALUES 
    ('+905551111111', 'Test message 1 - Please check your account balance'),
    ('+905552222222', 'Test message 2 - Your package has been delivered'),
    ('+905553333333', 'Test message 3 - Your appointment is confirmed for tomorrow'),
    ('+905554444444', 'Test message 4 - Your password has been reset successfully'),
    ('+905555555555', 'Test message 5 - Thank you for your purchase')
ON CONFLICT DO NOTHING;
