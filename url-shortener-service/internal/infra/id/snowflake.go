package id

import (
	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func Init(nodeID int64) error {
	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		return err
	}
	node = n
	return nil
}

func New() int64 {
	return node.Generate().Int64()
}
