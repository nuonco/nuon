package handlers

type paginationInfo struct {
	HasNext bool `json:"hasNext"`
}

type timelinePayload struct {
	Data       any            `json:"data"`
	Pagination paginationInfo `json:"pagination"`
}

type actionRunTimelinePayload struct {
	Data       any            `json:"data"`
	Pagination paginationInfo `json:"pagination"`
}

func isNotFoundErr(err error) bool {
	if c, ok := err.(interface{ Code() int }); ok {
		return c.Code() == 404
	}
	return false
}
