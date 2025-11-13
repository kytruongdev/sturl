package id

import (
	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// Init initializes the global Snowflake node used for generating unique IDs.
//
// Each node represents a unique machine or process in a distributed system.
// nodeID must be between 0 and 1023 (fits in 10 bits).
//
// This function should be called once during service startup (e.g., in main.go).
// Example:
//
//	err := id.Init(1)
//	if err != nil { panic(err) }
//
// After initialization, use New() to generate IDs.
func Init(nodeID int64) error {
	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		return err
	}
	node = n
	return nil
}

// New generates a new globally unique int64 ID using the initialized Snowflake node.
//
// This returns a 64-bit integer composed of:
//   - timestamp (ms)
//   - node ID (10 bits)
//   - sequence number (12 bits)
//
// REQUIREMENT:
//
//	Init() must be called before using New(), otherwise node will be nil
//	and calling this will panic.
//
// Example:
//
//	id := id.New()
//	fmt.Println(id)  // 1823412341234123456
func New() int64 {
	return node.Generate().Int64()
}
