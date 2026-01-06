package middleware

import "net/http"

type Middleware func(next http.Handler) http.Handler

func Wrap(next http.Handler, mw Middleware) http.Handler {
	return mw(next)
}
