package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func decodeJSON(ctx *gin.Context, target any) error {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return bodyValidationError("Ожидается корректный JSON не более 100 КБ")
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return bodyValidationError("Ожидается один JSON-объект")
	}
	return nil
}
