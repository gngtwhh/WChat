package service

import "strings"

func buildConversationID(sessionType int8, userID, targetID string) string {
	if sessionType == 1 {
		return "g:" + targetID
	}

	left, right := userID, targetID
	if strings.Compare(left, right) > 0 {
		left, right = right, left
	}
	return "p:" + left + ":" + right
}
