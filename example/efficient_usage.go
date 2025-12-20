package example

import (
	"context"
	"github.com/Lunar-Chipter/mire/core"
	"github.com/Lunar-Chipter/mire/logger"
)

// Example of unified API usage
func efficientLoggingExample() {
	log := logger.NewDefaultLogger()
	defer log.Close()

	// Basic logging
	log.Log(context.Background(), core.INFO, []byte("This is the unified API"), nil)
	
	// With context
	ctx := context.Background()
	log.Log(ctx, core.INFO, []byte("Context-aware logging"), nil)
	
	// With fields
	fields := map[string][]byte{
		"user_id": logger.I2B(12345),
		"action":  []byte("login"),
		"success": logger.B2B(true),
		"score":   logger.F2B(98.5),
	}
	log.Log(context.Background(), core.INFO, []byte("User login successful"), fields)
	
	// Complete API: context + fields
	log.Log(ctx, core.INFO, []byte("Complete logging"), fields)
}
