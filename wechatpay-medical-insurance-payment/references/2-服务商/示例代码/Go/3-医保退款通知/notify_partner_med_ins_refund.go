package main

import (
	"bytes"
	"demo/wxpay_utility"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// 服务商医保退款通知 —— 服务商代子商户告知微信医保侧已发生退款
// 与商户版差异：请求体必须额外传入 sub_mchid（医疗机构商户号）
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

	request := &NotifyRefundRequest{
		MixTradeNo:        wxpay_utility.String("202204022005169952975171534816"),
		SubMchid:          wxpay_utility.String("1900000109"),
		MedRefundTotalFee: wxpay_utility.Int64(45000),
		MedRefundGovFee:   wxpay_utility.Int64(45000),
		MedRefundSelfFee:  wxpay_utility.Int64(0),
		MedRefundOtherFee: wxpay_utility.Int64(0),
		RefundTime:        wxpay_utility.String("2015-05-20T13:29:35+08:00"),
		OutRefundNo:       wxpay_utility.String("R202204022005169952975171534816"),
	}

	if err := NotifyRefund(config, request); err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		return
	}
	fmt.Println("医保退款通知成功")
}

func NotifyRefund(config *wxpay_utility.MchConfig, request *NotifyRefundRequest) error {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/med-ins/refunds/notify"
	)

	reqUrl, err := url.Parse(fmt.Sprintf("%s%s", host, path))
	if err != nil {
		return err
	}
	query := reqUrl.Query()
	if request.MixTradeNo != nil {
		query.Add("mix_trade_no", *request.MixTradeNo)
	}
	reqUrl.RawQuery = query.Encode()
	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}
	httpRequest, err := http.NewRequest(method, reqUrl.String(), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Wechatpay-Serial", config.WechatPayPublicKeyId())
	httpRequest.Header.Set("Content-Type", "application/json")
	authorization, err := wxpay_utility.BuildAuthorization(config.MchId(), config.CertificateSerialNo(), config.PrivateKey(), method, reqUrl.RequestURI(), reqBody)
	if err != nil {
		return err
	}
	httpRequest.Header.Set("Authorization", authorization)

	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return err
	}
	respBody, err := wxpay_utility.ExtractResponseBody(httpResponse)
	if err != nil {
		return err
	}
	if httpResponse.StatusCode >= 200 && httpResponse.StatusCode < 300 {
		return wxpay_utility.ValidateResponse(config.WechatPayPublicKeyId(), config.WechatPayPublicKey(), &httpResponse.Header, respBody)
	}
	return wxpay_utility.NewApiException(httpResponse.StatusCode, httpResponse.Header, respBody)
}

type NotifyRefundRequest struct {
	MixTradeNo        *string `json:"mix_trade_no,omitempty"`
	SubMchid          *string `json:"sub_mchid,omitempty"`
	MedRefundTotalFee *int64  `json:"med_refund_total_fee,omitempty"`
	MedRefundGovFee   *int64  `json:"med_refund_gov_fee,omitempty"`
	MedRefundSelfFee  *int64  `json:"med_refund_self_fee,omitempty"`
	MedRefundOtherFee *int64  `json:"med_refund_other_fee,omitempty"`
	RefundTime        *string `json:"refund_time,omitempty"`
	OutRefundNo       *string `json:"out_refund_no,omitempty"`
}

func (o *NotifyRefundRequest) MarshalJSON() ([]byte, error) {
	type Alias NotifyRefundRequest
	a := &struct {
		MixTradeNo *string `json:"mix_trade_no,omitempty"`
		*Alias
	}{
		MixTradeNo: nil,
		Alias:      (*Alias)(o),
	}
	return json.Marshal(a)
}
