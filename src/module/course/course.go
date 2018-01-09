package course

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/asepnur/meiko_bot/src/util/conn"
)

func SelectAllName() ([]string, error) {

	var names []string
	query := fmt.Sprintf(`SELECT name FROM courses;`)
	err := conn.DB.Select(&names, query)
	if err != nil && err != sql.ErrNoRows {
		return names, err
	}

	return names, nil
}

func SelectAllAssistantID() ([]int64, error) {

	userIDs := []int64{}
	query := fmt.Sprintf(`SELECT
		users_id
	FROM
		p_users_schedules
	WHERE 
		status = (%d);`, PStatusAssistant)
	err := conn.DB.Select(&userIDs, query)
	if err != nil && err != sql.ErrNoRows {
		return userIDs, err
	}

	return userIDs, nil
}

func SelectEnrolledSchedule(userID int64) ([]CourseSchedule, error) {

	var res []CourseSchedule

	data := url.Values{}
	data.Set("user_id", strconv.FormatInt(userID, 10))
	data.Set("role", "student")

	params := data.Encode()
	req, err := http.NewRequest("POST", "http://localhost:9001/api/internal/v1/course/getall", strings.NewReader(params))
	if err != nil {
		return res, err
	}
	req.Header.Add("Authorization", "abc")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(params)))

	client := http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	jsn := GetAllHTTPResponse{}
	err = json.Unmarshal(body, &jsn)
	if err != nil {
		return res, err
	}

	if jsn.Code != 200 {
		return res, fmt.Errorf("error request")
	}

	return jsn.Data, nil
}
