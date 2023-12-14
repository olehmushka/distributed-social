package schemas

type GetPingRespData struct {
	Ok bool `json:"ok"`
}

type GetPingResp SuccessResp[GetPingRespData]
