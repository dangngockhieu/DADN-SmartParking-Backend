package internal

type Meta struct {
	TotalElements int64 `json:"totalElements"`
	TotalPages    int   `json:"totalPages"`
	CurrentPage   int   `json:"currentPage"`
	PageSize      int   `json:"pageSize"`
}

type Response[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

func NewResponse[T any](
	data []T,
	totalElements int64,
	currentPage int,
	pageSize int,
) Response[T] {
	if currentPage < 1 {
		currentPage = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := 0
	if totalElements > 0 {
		totalPages = int((totalElements + int64(pageSize) - 1) / int64(pageSize))
	}

	return Response[T]{
		Data: data,
		Meta: Meta{
			TotalElements: totalElements,
			TotalPages:    totalPages,
			CurrentPage:   currentPage,
			PageSize:      pageSize,
		},
	}
}
