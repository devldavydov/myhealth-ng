package cases

import (
	"strings"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

func normalizePageRequest(request entity.PageRequest) (entity.PageRequest, error) {
	request.Query = strings.TrimSpace(request.Query)
	if request.Page == 0 {
		request.Page = entity.DefaultPage
	}
	if request.PageSize == 0 {
		request.PageSize = entity.DefaultPageSize
	}
	details := make(map[string][]string)
	if request.Page < 1 {
		details["page"] = []string{"Номер страницы должен быть не меньше 1"}
	}
	if request.PageSize != 10 && request.PageSize != 20 && request.PageSize != 50 {
		details["pageSize"] = []string{"Допустимые значения: 10, 20 или 50"}
	}
	if len(details) > 0 {
		return entity.PageRequest{}, &entity.ValidationError{Fields: details}
	}
	return request, nil
}
