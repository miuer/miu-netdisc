package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/miuer/miu-netdisc/filestore/model/mysql"
)

// CheckTokenValidity -
func (ctl *Controller) CheckTokenValidity(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("tk")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Failed to get token, err:"+err.Error())
			return
		}

		if cookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Invalid token")
			return
		}

		token := cookie.Value

		ts, _ := strconv.ParseInt(token[32:], 16, 64)
		tm := time.Now().Unix()

		if (tm - ts - 1800) > 0 {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "expired token")
			return
		}

		id, err := mysql.GetIDByToken(ctl.Reader, token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to get token id, err:"+err.Error())
			return
		}

		if id > 0 {
			//	w.WriteHeader(http.StatusFound)
			handle.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Invalid token")
		return
	})
}
