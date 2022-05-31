package idutils

import (
	"github.com/bwmarrin/snowflake"
	"github.com/pkg/errors"
)

func GenerateSnowflakeId() (string, error) {
	// Generate an ID
	sfNode, err := snowflake.NewNode(1)
	if err != nil {
		return "", errors.Wrap(err, "无法生成 ID")
	}

	id := sfNode.Generate().String()
	return id, nil
}
