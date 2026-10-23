package headers

import "net/http"

const UserID = "X-User-ID"

func GetUserID(r *http.Request) (string, bool) {
	userID := r.Header.Get(UserID)
	return userID, len(userID) != 0
}
