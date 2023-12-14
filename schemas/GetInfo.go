package schemas

type GetInfoRespData struct {
	Name string `json:"name"`
}

type GetInfoResp SuccessResp[GetInfoRespData]
