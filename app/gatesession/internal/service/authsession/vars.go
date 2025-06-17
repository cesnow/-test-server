package authsession

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	onlineKeyPrefix  = "online"           //
	userKeyIdsPrefix = "user_online_keys" //
)

func GetOnlineUserKey(id int64) string {
	return fmt.Sprintf("%s#%d", userKeyIdsPrefix, id)
}

func getIdByUserKey(k string) int64 {
	a := strings.Split(k, "#")
	if len(a) < 2 {
		return 0
	}
	i, _ := strconv.ParseInt(a[1], 10, 64)

	return i
}
