package bot

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/asepnur/meiko_bot/src/module/bot"
	"github.com/asepnur/meiko_bot/src/util/auth"
	"github.com/asepnur/meiko_bot/src/util/conn"
	"github.com/asepnur/meiko_bot/src/webserver/template"
	"github.com/julienschmidt/httprouter"
)

func BotHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	sess := r.Context().Value("User").(*auth.User)
	params := messageParams{
		Text: r.FormValue("text"),
	}

	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError(err.Error()))
		return
	}

	// get text intent
	intent, _ := bot.GetIntent(args.NormalizedText)

	// convert intent into assistant
	data := []map[string]interface{}{}
	switch intent {
	case intentAssistant:
		data, err = handleAssistant(args.NormalizedText, sess.ID)
	case intentGrade:
		data, err = handleGrade(args.NormalizedText, sess.ID)
	case intentAssignment:
		data, err = handleAssignment(args.NormalizedText, sess.ID)
	case intentInformation:
		data, err = handleInformation(args.NormalizedText, sess.ID)
	case intentSchedule:
		data, err = handleSchedule(args.NormalizedText, sess.ID)
	case intentUnknown:
		break
	default:
		break
	}

	// intent and entity
	respData := map[string]interface{}{
		"intent": intent,
		"entity": data,
	}

	// log message into database
	jsn, _ := json.Marshal(respData)
	jsnStr := string(jsn)

	tx, err := conn.DB.Beginx()
	if err != nil {
		return
	}

	_, err = bot.Insert(args.Text, sess.ID, bot.StatusUser, tx)
	if err != nil {
		tx.Rollback()
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}

	lastInsertID, err := bot.Insert(jsnStr, sess.ID, bot.StatusBot, tx)
	if err != nil {
		tx.Rollback()
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}

	tx.Commit()

	// prepare for response
	respData["id"] = lastInsertID
	resp := messageResponse{
		Status:    bot.StatusBot,
		Text:      args.Text,
		TimeStamp: time.Now().Unix(),
		Response:  respData,
	}

	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetData(resp))
	return
}

func LoadHistoryHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	sess := r.Context().Value("User").(*auth.User)
	params := loadHistoryParams{
		id: r.FormValue("id"),
	}

	args, err := params.validate()
	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusBadRequest).
			AddError("Invalid Request"))
		return
	}

	var log []bot.Log
	if args.id.Valid {
		log, err = bot.LoadByID(args.id.Int64, sess.ID)
	} else {
		log, err = bot.LoadByUserID(sess.ID)
	}

	if err != nil {
		template.RenderJSONResponse(w, new(template.Response).
			SetCode(http.StatusInternalServerError))
		return
	}

	resp := []map[string]interface{}{}
	for _, val := range log {
		if val.Status == bot.StatusUser {
			resp = append(resp, map[string]interface{}{
				"status": bot.StatusUser,
				"time":   val.CreatedAt.Unix(),
				"message": map[string]interface{}{
					"id":   val.ID,
					"text": val.Message,
				},
			})
			continue
		}

		jsnMap := map[string]interface{}{}
		json.Unmarshal([]byte(val.Message), &jsnMap)
		jsnMap["id"] = val.ID
		resp = append(resp, map[string]interface{}{
			"status":  bot.StatusBot,
			"time":    val.CreatedAt.Unix(),
			"message": jsnMap,
		})
	}

	template.RenderJSONResponse(w, new(template.Response).
		SetCode(http.StatusOK).
		SetData(resp))
	return
}
