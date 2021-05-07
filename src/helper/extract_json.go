package Helper

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func GetJSONRawBody(c echo.Context) map[string]interface{} {

	jsonBody := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonBody)
	if err != nil {

		log.Error("empty json body")
		return nil
	}

	return jsonBody
}
