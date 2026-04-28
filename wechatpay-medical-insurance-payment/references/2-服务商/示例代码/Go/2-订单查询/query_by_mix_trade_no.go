package main

import (
	"demo/wxpay_utility"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	config, err := wxpay_utility.CreateMchConfig(
		"1900000100",
		"1DDE55AD98Exxxxxxxxxx",
		"/path/to/apiclient_key.pem",
		"PUB_KEY_ID_xxxxxxxxxxxxx",
		"/path/to/wxp_pub.pem",
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	request := &QueryRequest{
		MixTradeNo: wxpay_utility.String("202204022005169952975171534816"),
		SubMchid:   wxpay_utility.String("1900000109"),
	}

	response, err := QueryByMixTradeNo(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		return
	}
	fmt.Printf("请求成功: %+v\n", response)
}

func QueryByMixTradeNo(config *wxpay_utility.MchConfig, request *QueryRequest) (response *OrderEntity, err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "GET"
		path   = "/v3/med-ins/orders/mix-trade-no/{mix_trade_no}"
	)

	reqUrl, err := url.Parse(fmt.Sprintf("%s%s", host, path))
	if err != nil {
		return nil, err
	}
	reqUrl.Path = strings.Replace(reqUrl.Path, "{mix_trade_no}", url.PathEscape(*request.MixTradeNo), -1)
	query := reqUrl.Query()
	if request.SubMchid != nil {
		query.Add("sub_mchid", *request.SubMchid)
	}
	reqUrl.RawQuery = query.Encode()

	httpRequest, err := http.NewRequest(method, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Wechatpay-Serial", config.WechatPayPublicKeyId())
	authorization, err := wxpay_utility.BuildAuthorization(config.MchId(), config.CertificateSerialNo(), config.PrivateKey(), method, reqUrl.RequestURI(), nil)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", authorization)

	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	respBody, err := wxpay_utility.ExtractResponseBody(httpResponse)
	if err != nil {
		return nil, err
	}
	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 {
		err = wxpay_utility.ValidateResponse(config.WechatPayPublicKeyId(), config.WechatPayPublicKey(), &httpResponse.Header, respBody)
		if err != nil {
			return nil, err
		}
		response := &OrderEntity{}
		if err := json.Unmarshal(respBody, response); err != nil {
			return nil, err
		}
		return response, nil
	}
	return nil, wxpay_utility.NewApiException(httpResponse.StatusCode, httpResponse.Header, respBody)
}

type QueryRequest struct {
	MixTradeNo *string `json:"mix_trade_no,omitempty"`
	SubMchid   *string `json:"sub_mchid,omitempty"`
}

func (o *QueryRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{}{})
}

type OrderEntity struct {
	MixTradeNo       *string `json:"mix_trade_no,omitempty"`
	MixPayStatus     *string `json:"mix_pay_status,omitempty"`
	SelfPayStatus    *string `json:"self_pay_status,omitempty"`
	MedInsPayStatus  *string `json:"med_ins_pay_status,omitempty"`
	PaidTime         *string `json:"paid_time,omitempty"`
	MixPayType       *string `json:"mix_pay_type,omitempty"`
	OrderType        *string `json:"order_type,omitempty"`
	Appid            *string `json:"appid,omitempty"`
	SubAppid         *string `json:"sub_appid,omitempty"`
	SubMchid         *string `json:"sub_mchid,omitempty"`
	Openid           *string `json:"openid,omitempty"`
	SubOpenid        *string `json:"sub_openid,omitempty"`
	OutTradeNo       *string `json:"out_trade_no,omitempty"`
	TotalFee         *int64  `json:"total_fee,omitempty"`
	MedInsGovFee     *int64  `json:"med_ins_gov_fee,omitempty"`
	MedInsSelfFee    *int64  `json:"med_ins_self_fee,omitempty"`
	MedInsOtherFee   *int64  `json:"med_ins_other_fee,omitempty"`
	MedInsCashFee    *int64  `json:"med_ins_cash_fee,omitempty"`
	WechatPayCashFee *int64  `json:"wechat_pay_cash_fee,omitempty"`
	MedInsFailReason *string `json:"med_ins_fail_reason,omitempty"`
}
