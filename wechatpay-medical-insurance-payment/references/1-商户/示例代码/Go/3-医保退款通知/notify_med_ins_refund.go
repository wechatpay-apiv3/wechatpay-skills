package main

import (
	"bytes"
	"demo/wxpay_utility"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// NotifyRefund 医保退款通知 —— 商户主动告知微信医保侧已发生退款
//
// 流程:
//
//	1) 医院 HIS 在医保局发起医保退款 → 医保局完成退款
//	2) 商户调用本接口告知微信
//	3) 若同时存在自费退款，请先调用 POST /v3/refund/domestic/refunds，再用相同 out_refund_no 调用本接口
func main() {
	config, err := wxpay_utility.CreateMchConfig(
		"19xxxxxxxx",
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
