package sqlmodel

import (
	"database/sql"

	"github.com/bwmarrin/snowflake"
)

func parseSnowflakeStringToNullInt64(str string) (sql.NullInt64, error) {
	var ret sql.NullInt64
	if str != "" {
		sfID, err := snowflake.ParseString(str)
		if err != nil {
			return ret, err
		}
		_ = ret.Scan(sfID.String())
	}

	return ret, nil
}

func parseSnowflakeStringToInt64(str string) (int64, error) {
	sfID, err := snowflake.ParseString(str)
	if err != nil {
		return 0, err
	}

	return sfID.Int64(), nil
}

func parseNullInt64ToSnowflakeString(i sql.NullInt64) string {
	if !i.Valid {
		return ""
	}

	return snowflake.ParseInt64(i.Int64).String()
}

func parseInt64ToSnowflakeString(i int64) string {
	return snowflake.ParseInt64(i).String()
}
