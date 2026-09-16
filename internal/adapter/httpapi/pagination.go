package httpapi

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type paginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

func decodePageRequest(ctx *gin.Context) (entity.PageRequest, error) {
	details := make(map[string][]string)
	page := parseOptionalPageNumber(ctx.Query("page"), "page", details)
	pageSize := parseOptionalPageNumber(ctx.Query("pageSize"), "pageSize", details)
	if len(details) > 0 {
		return entity.PageRequest{}, &entity.ValidationError{Fields: details}
	}
	return entity.PageRequest{Query: ctx.Query("q"), Page: page, PageSize: pageSize}, nil
}

func parseOptionalPageNumber(raw, field string, details map[string][]string) int {
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		details[field] = []string{"Ожидается целое число"}
		return 0
	}
	return value
}

func toPaginationResponse[T any](page entity.Page[T]) paginationResponse {
	return paginationResponse{Page: page.Page, PageSize: page.PageSize, Total: page.Total, TotalPages: page.TotalPages}
}
