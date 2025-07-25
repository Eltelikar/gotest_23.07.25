package response

import "gotest_23.07.25/internal/postgre"

type Response struct {
	Status       string                      `json:"status"`
	Error        string                      `json:"error,omitempty"`
	Message      string                      `json:"message,omitempty"`
	Fields       postgre.RequestFields       `json:"fields,omitempty"`
	FieldsUpdate postgre.RequestUpdateFields `json:"fieldsUpd,omitempty"`
}

func OK(msg string, rb *postgre.RequestFields) Response {
	return Response{
		Status:  "success",
		Message: msg,
		Fields:  *rb,
	}
}

func OKUpdate(msg string, rb *postgre.RequestUpdateFields) Response {
	return Response{
		Status:       "success",
		Message:      msg,
		FieldsUpdate: *rb,
	}
}

func Error(msg string) Response {
	return Response{
		Status: "error",
		Error:  msg,
	}
}
