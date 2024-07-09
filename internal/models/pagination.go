package models

type OffsetPaginationDTO struct {
	Offset  int   `json:"offset"`
	Limit   int   `json:"limit"`
	Records int64 `json:"records"`
}

type PostCursorPaginationDTO struct {
	LastPostId string `json:"lastPostId"`
	Limit      int    `json:"limit"`
	Records    int64  `json:"records"`
}
