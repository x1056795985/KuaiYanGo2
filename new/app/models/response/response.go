package response

type GetList struct {
	List  interface{} `json:"List"`  // 列表
	Count int64       `json:"Count"` // 总数
}
