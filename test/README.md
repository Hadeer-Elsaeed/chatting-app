# Chat Application Tests

This directory contains focused unit tests for the Go chatting application, covering the most important authentication and messaging scenarios.

## Test Structure

```
test/
├── go.mod                              # Test module dependencies
├── run_tests.sh                        # Test runner script
├── README.md                           # This file
├── auth_handler_test.go                # Authentication validation tests
└── message_handler_test.go             # Message handling tests
```

## Test Coverage

### 🔐 Authentication Tests (`auth_handler_test.go`)
- **Password validation**: Strong password requirements
- **Email validation**: Proper email format checking  
- **Username validation**: Valid username criteria
- **Input sanitization**: Protection against invalid inputs

### 💬 Message Handler Tests (`message_handler_test.go`)
- **Message validation**: Content length and format checks
- **Message type validation**: Direct vs broadcast message types
- **HTML sanitization**: XSS protection
- **Recipient validation**: Proper recipient handling

## Running Tests

### Option 1: Run All Tests (Linux/Mac)
```bash
cd test
chmod +x run_tests.sh
./run_tests.sh
```

### Option 2: Run Individual Test Files

```bash
# Authentication tests
cd test
go test -v auth_handler_test.go

# Message handler tests  
cd test
go test -v message_handler_test.go

# Run all tests
cd test
go test -v
```

## Important Test Scenarios Covered

### 🚨 Security Tests
- Password strength validation
- Email format validation
- HTML/XSS injection prevention
- Input sanitization

### 📨 Messaging Tests
- Direct message validation
- Broadcast message functionality
- Message content limits
- Media attachment validation

## Test Dependencies

The tests use minimal dependencies:
- **Go standard library**: For most backend tests
- **No external test frameworks**: Keeping it simple and focused

## Adding New Tests

When adding new features, create tests following this pattern:

1. **Create test file**: `feature_name_test.go`
2. **Use table-driven tests**: Test multiple scenarios efficiently
3. **Test edge cases**: Invalid inputs, empty data, limits
4. **Test error handling**: Ensure graceful failure handling
5. **Keep tests focused**: Test one thing at a time

Example test structure:
```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        hasError bool
    }{
        {"Valid Input", validInput, expectedOutput, false},
        {"Invalid Input", invalidInput, errorOutput, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := functionUnderTest(tt.input)
            
            if tt.hasError && err == nil {
                t.Error("Expected error but got none")
            }
            
            if !tt.hasError && err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            
            if result != tt.expected {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

## Test Philosophy

These tests focus on:
- **Critical functionality**: Authentication and messaging validation
- **Security vulnerabilities**: Input validation, XSS prevention
- **User experience**: Proper error handling, validation feedback

The tests are designed to be:
- **Fast**: Quick execution for rapid feedback
- **Focused**: Testing specific, important scenarios
- **Maintainable**: Easy to understand and modify
- **Independent**: Tests don't depend on each other
