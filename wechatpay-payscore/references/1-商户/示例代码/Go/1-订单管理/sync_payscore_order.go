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

	request := &SyncServiceOrderRequest{
		OutOrderNo: wxpay_utility.String("1234323JKHDFE1243252"),
		Appid:      wxpay_utility.String("wxd678efh567hg6787"),
		ServiceId:  wxpay_utility.String("2002000000000558128851361561536"),
		Type:       wxpay_utility.String("Order_Paid"),
		Detail: &SyncDetail{
			PaidTime: wxpay_utility.String("20091225091210"),
		},
	}

	response, err := SyncServiceOrder(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		// TODO: 请求失败，根据状态码执行不同的处理
		return
	}

	// TODO: 请求成功，继续业务逻辑
	fmt.Printf("请求成功: %+v\n", response)
}

func SyncServiceOrder(config *wxpay_utility.MchConfig, request *SyncServiceOrderRequest) (response *ServiceOrderEntity, err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/payscore/serviceorder/{out_order_no}/sync"
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
		// 2XX 成功，验证应答签名
		err = wxpay_utility.ValidateResponse(
			config.WechatPayPublicKeyId(),
			config.WechatPayPublicKey(),
			&httpResponse.Header,
			respBody,
		)
		if err != nil {
			return nil, err
		}
		response := &ServiceOrderEntity{}
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

type SyncServiceOrderRequest struct {
	OutOrderNo *string     `json:"out_order_no,omitempty"`
	Appid      *string     `json:"appid,omitempty"`
	ServiceId  *string     `json:"service_id,omitempty"`
	Type       *string     `json:"type,omitempty"`
	Detail     *SyncDetail `json:"detail,omitempty"`
}

func (o *SyncServiceOrderRequest) MarshalJSON() ([]byte, error) {
	type Alias SyncServiceOrderRequest
	a := &struct {
		OutOrderNo *string `json:"out_order_no,omitempty"`
		*Alias
	}{
		// 序列化时移除非 Body 字段
		OutOrderNo: nil,
		Alias:      (*Alias)(o),
	}
	return json.Marshal(a)
}

type ServiceOrderEntity struct {
	OutOrderNo          *string              `json:"out_order_no,omitempty"`
	ServiceId           *string              `json:"service_id,omitempty"`
	Appid               *string              `json:"appid,omitempty"`
	Mchid               *string              `json:"mchid,omitempty"`
	ServiceIntroduction *string              `json:"service_introduction,omitempty"`
	State               *string              `json:"state,omitempty"`
	StateDescription    *string              `json:"state_description,omitempty"`
	PostPayments        *Payment             `json:"post_payments,omitempty"`
	PostDiscounts       []ServiceOrderCoupon `json:"post_discounts,omitempty"`
	RiskFund            *RiskFund            `json:"risk_fund,omitempty"`
	TotalAmount         *int64               `json:"total_amount,omitempty"`
	NeedCollection      *bool                `json:"need_collection,omitempty"`
	Collection          *Collection          `json:"collection,omitempty"`
	TimeRange           *TimeRange           `json:"time_range,omitempty"`
	Location            *Location            `json:"location,omitempty"`
	Attach              *string              `json:"attach,omitempty"`
	NotifyUrl           *string              `json:"notify_url,omitempty"`
	Openid              *string              `json:"openid,omitempty"`
	OrderId             *string              `json:"order_id,omitempty"`
}

type SyncDetail struct {
	PaidTime *string `json:"paid_time,omitempty"`
}

type Payment struct {
	Name        *string `json:"name,omitempty"`
	Amount      *int64  `json:"amount,omitempty"`
	Description *string `json:"description,omitempty"`
	Count       *int64  `json:"count,omitempty"`
}

type ServiceOrderCoupon struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Amount      *int64  `json:"amount,omitempty"`
	Count       *int64  `json:"count,omitempty"`
}

type RiskFund struct {
	Name        *string `json:"name,omitempty"`
	Amount      *int64  `json:"amount,omitempty"`
	Description *string `json:"description,omitempty"`
}

type Collection struct {
	State        *string  `json:"state,omitempty"`
	TotalAmount  *int64   `json:"total_amount,omitempty"`
	PayingAmount *int64   `json:"paying_amount,omitempty"`
	PaidAmount   *int64   `json:"paid_amount,omitempty"`
	Details      []Detail `json:"details,omitempty"`
}

type TimeRange struct {
	StartTime       *string `json:"start_time,omitempty"`
	EndTime         *string `json:"end_time,omitempty"`
	StartTimeRemark *string `json:"start_time_remark,omitempty"`
	EndTimeRemark   *string `json:"end_time_remark,omitempty"`
}

type Location struct {
	StartLocation *string `json:"start_location,omitempty"`
	EndLocation   *string `json:"end_location,omitempty"`
}

type Detail struct {
	Seq           *int64  `json:"seq,omitempty"`
	Amount        *int64  `json:"amount,omitempty"`
	PaidType      *string `json:"paid_type,omitempty"`
	PaidTime      *string `json:"paid_time,omitempty"`
	TransactionId *string `json:"transaction_id,omitempty"`
}

