package main

import (
	"bytes"
	"demo/wxpay_utility" // 引用微信支付工具库，参考 https://pay.weixin.qq.com/doc/v3/merchant/4015119334
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	// TODO: 请准备商户开发必要参数，参考：https://pay.weixin.qq.com/doc/v3/merchant/4013070756
	config, err := wxpay_utility.CreateMchConfig(
		"19xxxxxxxx",                 // 商户号，是由微信支付系统生成并分配给每个商户的唯一标识符，商户号获取方式参考 https://pay.weixin.qq.com/doc/v3/merchant/4013070756
		"1DDE55AD98Exxxxxxxxxx",      // 商户API证书序列号，如何获取请参考 https://pay.weixin.qq.com/doc/v3/merchant/4013053053
		"/path/to/apiclient_key.pem", // 商户API证书私钥文件路径，本地文件路径
		"PUB_KEY_ID_xxxxxxxxxxxxx",   // 微信支付公钥ID，如何获取请参考 https://pay.weixin.qq.com/doc/v3/merchant/4013038816
		"/path/to/wxp_pub.pem",       // 微信支付公钥文件路径，本地文件路径
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	request := &CancelServiceOrderRequest{
		OutOrderNo: wxpay_utility.String("2304203423948239423"),
		Appid:      wxpay_utility.String("wxd678efh567hg6787"),
		ServiceId:  wxpay_utility.String("2002000000000558128851361561536"),
		Reason:     wxpay_utility.String("用户投诉"),
	}

	response, err := CancelServiceOrder(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		// TODO: 请求失败，根据状态码执行不同的处理
		return
	}

	// TODO: 请求成功，继续业务逻辑
	fmt.Printf("请求成功: %+v\n", response)
}

func CancelServiceOrder(config *wxpay_utility.MchConfig, request *CancelServiceOrderRequest) (response *CancelServiceOrderResponse, err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/payscore/serviceorder/{out_order_no}/cancel"
	)

	reqUrl, err := url.Parse(fmt.Sprintf("%s%s", host, path))
	if err != nil {
		return nil, err
	}
	reqUrl.Path = strings.Replace(reqUrl.Path, "{out_order_no}", url.PathEscape(*request.OutOrderNo), -1)
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequest(method, reqUrl.String(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Wechatpay-Serial", config.WechatPayPublicKeyId())
	httpRequest.Header.Set("Content-Type", "application/json")
	authorization, err := wxpay_utility.BuildAuthorization(config.MchId(), config.CertificateSerialNo(), config.PrivateKey(), method, reqUrl.RequestURI(), reqBody)
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
		err = wxpay_utility.ValidateResponse(
			config.WechatPayPublicKeyId(),
			config.WechatPayPublicKey(),
			&httpResponse.Header,
			respBody,
		)
		if err != nil {
			return nil, err
		}
		response := &CancelServiceOrderResponse{}
		if err := json.Unmarshal(respBody, response); err != nil {
			return nil, err
		}

		return response, nil
	} else {
		return nil, wxpay_utility.NewApiException(
			httpResponse.StatusCode,
			httpResponse.Header,
			respBody,
		)
	}
}

type CancelServiceOrderRequest struct {
	OutOrderNo *string `json:"out_order_no,omitempty"`
	Appid      *string `json:"appid,omitempty"`
	ServiceId  *string `json:"service_id,omitempty"`
	Reason     *string `json:"reason,omitempty"`
}

func (o *CancelServiceOrderRequest) MarshalJSON() ([]byte, error) {
	type Alias CancelServiceOrderRequest
	a := &struct {
		OutOrderNo *string `json:"out_order_no,omitempty"`
		*Alias
	}{
		OutOrderNo: nil,
		Alias:      (*Alias)(o),
	}
	return json.Marshal(a)
}

type CancelServiceOrderResponse struct {
	Appid      *string `json:"appid,omitempty"`
	Mchid      *string `json:"mchid,omitempty"`
	OutOrderNo *string `json:"out_order_no,omitempty"`
	ServiceId  *string `json:"service_id,omitempty"`
	OrderId    *string `json:"order_id,omitempty"`
}
