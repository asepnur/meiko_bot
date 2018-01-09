package webserver

import (
	"github.com/asepnur/meiko_bot/src/util/auth"
	"github.com/asepnur/meiko_bot/src/webserver/handler/bot"
	"github.com/julienschmidt/httprouter"
)

// Load returns all routing of this server
func loadRouter(r *httprouter.Router) {
	// ========================== User Handler ==========================
	// User section
	r.GET("/api/v1/bot", auth.MustAuthorize(bot.LoadHistoryHandler))
	r.POST("/api/v1/bot", auth.MustAuthorize(bot.BotHandler))
	// ========================= End Bot Handler ========================
}
