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

	request := &ModifyServiceOrderRequest{
		OutOrderNo: wxpay_utility.String("1234323JKHDFE1243252"),
		Appid:      wxpay_utility.String("wxd678efh567hg6787"),
		ServiceId:  wxpay_utility.String("2002000000000558128851361561536"),
		PostPayments: []Payment{Payment{
			Name:        wxpay_utility.String("就餐费用"),
			Amount:      wxpay_utility.Int64(40000),
			Description: wxpay_utility.String("就餐人均100元"),
			Count:       wxpay_utility.Int64(4),
		}},
		PostDiscounts: []ServiceOrderCoupon{ServiceOrderCoupon{
			Name:        wxpay_utility.String("满20减1元"),
			Description: wxpay_utility.String("不与其他优惠叠加"),
			Amount:      wxpay_utility.Int64(100),
			Count:       wxpay_utility.Int64(2),
		}},
		TotalAmount: wxpay_utility.Int64(50000),
		Reason:      wxpay_utility.String("用户投诉"),
		Device: &Device{
			StartDeviceId: wxpay_utility.String("HG123456"),
			EndDeviceId:   wxpay_utility.String("HG123456"),
			MaterielNo:    wxpay_utility.String("example_materiel_no"),
		},
	}

	response, err := ModifyServiceOrder(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		// TODO: 请求失败，根据状态码执行不同的处理
		return
	}

	// TODO: 请求成功，继续业务逻辑
	fmt.Printf("请求成功: %+v\n", response)
}

func ModifyServiceOrder(config *wxpay_utility.MchConfig, request *ModifyServiceOrderRequest) (response *ServiceOrderEntity, err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/payscore/serviceorder/{out_order_no}/modify"
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

type ModifyServiceOrderRequest struct {
	OutOrderNo    *string              `json:"out_order_no,omitempty"`
	Appid         *string              `json:"appid,omitempty"`
	ServiceId     *string              `json:"service_id,omitempty"`
	PostPayments  []Payment            `json:"post_payments,omitempty"`
	PostDiscounts []ServiceOrderCoupon `json:"post_discounts,omitempty"`
	TotalAmount   *int64               `json:"total_amount,omitempty"`
	Reason        *string              `json:"reason,omitempty"`
	Device        *Device              `json:"device,omitempty"`
}

func (o *ModifyServiceOrderRequest) MarshalJSON() ([]byte, error) {
	type Alias ModifyServiceOrderRequest
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

type Device struct {
	StartDeviceId *string `json:"start_device_id,omitempty"`
	EndDeviceId   *string `json:"end_device_id,omitempty"`
	MaterielNo    *string `json:"materiel_no,omitempty"`
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

