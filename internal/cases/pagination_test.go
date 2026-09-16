package cases

import (
	"errors"
	"testing"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

func TestNormalizePageRequest(t *testing.T) {
	request, err := normalizePageRequest(entity.PageRequest{Query: "  творог  "})
	if err != nil {
		t.Fatal(err)
	}
	if request.Query != "творог" || request.Page != entity.DefaultPage || request.PageSize != entity.DefaultPageSize {
		t.Fatalf("request=%+v", request)
	}
	for _, pageSize := range []int{10, 20, 50} {
		request, err = normalizePageRequest(entity.PageRequest{Page: 2, PageSize: pageSize})
		if err != nil || request.Page != 2 || request.PageSize != pageSize {
			t.Fatalf("pageSize=%d request=%+v err=%v", pageSize, request, err)
		}
	}
}

func TestNormalizePageRequestRejectsInvalidValues(t *testing.T) {
	for _, request := range []entity.PageRequest{
		{Page: -1, PageSize: 20},
		{Page: 1, PageSize: 15},
	} {
		_, err := normalizePageRequest(request)
		var validationError *entity.ValidationError
		if !errors.As(err, &validationError) {
			t.Fatalf("request=%+v error=%v", request, err)
		}
	}
}
