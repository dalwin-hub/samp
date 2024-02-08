package utils

func UnauthorizedError() HttpResp {
	return HttpResp{
		Status: false,
		Error:  "You are not authorized to access this resource",
	}
}

func NotFoundError() *HttpResp {
	return &HttpResp{
		Status: false,
		Error:  "The requested resource was not found",
	}
}

func DataAccessLayerError(message string) *HttpResp {
	return &HttpResp{
		Status: false,
		Error:  message,
	}
}

func BadRequestError(message string) *HttpResp {
	return &HttpResp{
		Status: false,
		Error:  message,
	}
}
