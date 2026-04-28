package main

import (
	"bytes"
	"demo/wxpay_utility" // 引用微信支付工具库，参考 https://pay.weixin.qq.com/doc/v3/partner/4015119446
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	// TODO: 请准备商户开发必要参数，参考：https://pay.weixin.qq.com/doc/v3/partner/4013080340
	config, err := wxpay_utility.CreateMchConfig(
		"19xxxxxxxx",                 // 商户号，是由微信支付系统生成并分配给每个商户的唯一标识符，商户号获取方式参考 https://pay.weixin.qq.com/doc/v3/partner/4013080340
		"1DDE55AD98Exxxxxxxxxx",      // 商户API证书序列号，如何获取请参考 https://pay.weixin.qq.com/doc/v3/partner/4013058924
		"/path/to/apiclient_key.pem", // 商户API证书私钥文件路径，本地文件路径
		"PUB_KEY_ID_xxxxxxxxxxxxx",   // 微信支付公钥ID，如何获取请参考 https://pay.weixin.qq.com/doc/v3/partner/4013038589
		"/path/to/wxp_pub.pem",       // 微信支付公钥文件路径，本地文件路径
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	request := &CreatePartnerServiceOrderRequest{
		ServiceId:           wxpay_utility.String("2002000000000558128851361561536"),
		Appid:               wxpay_utility.String("wxd678efh567hg6787"),
		SubMchid:            wxpay_utility.String("1900000109"),
		SubAppid:            wxpay_utility.String("wxd678efh567hg6999"),
		OutOrderNo:          wxpay_utility.String("1234323JKHDFE1243252"),
		ServiceIntroduction: wxpay_utility.String("XX充电宝"),
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
		RiskFund: &RiskFund{
			Name:        wxpay_utility.String("DEPOSIT"),
			Amount:      wxpay_utility.Int64(10000),
			Description: wxpay_utility.String("就餐的预估费用"),
		},
		TimeRange: &TimeRange{
			StartTime:       wxpay_utility.String("20091225091010"),
			EndTime:         wxpay_utility.String("20091225121010"),
			StartTimeRemark: wxpay_utility.String("备注1"),
			EndTimeRemark:   wxpay_utility.String("备注2"),
		},
		Location: &Location{
			StartLocation: wxpay_utility.String("嗨客时尚主题展餐厅"),
			EndLocation:   wxpay_utility.String("嗨客时尚主题展餐厅"),
		},
		NeedUserConfirm: wxpay_utility.Bool(false),
		NotifyUrl:       wxpay_utility.String("https://api.test.com"),
		Attach:          wxpay_utility.String("Easdfowealsdkjfnlaksjdlfkwqoi&wl3l2sald"),
		Device: &Device{
			StartDeviceId: wxpay_utility.String("HG123456"),
			EndDeviceId:   wxpay_utility.String("HG123456"),
			MaterielNo:    wxpay_utility.String("example_materiel_no"),
		},
	}

	response, err := CreatePartnerServiceOrder(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		// TODO: 请求失败，根据状态码执行不同的处理
		return
	}

	// TODO: 请求成功，继续业务逻辑
	fmt.Printf("请求成功: %+v\n", response)
}

func CreatePartnerServiceOrder(config *wxpay_utility.MchConfig, request *CreatePartnerServiceOrderRequest) (response *CreatePartnerServiceOrderResponse, err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/payscore/partner/serviceorder"
	)

	reqUrl, err := url.Parse(fmt.Sprintf("%s%s", host, path))
	if err != nil {
		return nil, err
	}
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
		response := &CreatePartnerServiceOrderResponse{}
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

type CreatePartnerServiceOrderRequest struct {
	ServiceId           *string              `json:"service_id,omitempty"`
	Appid               *string              `json:"appid,omitempty"`
	SubMchid            *string              `json:"sub_mchid,omitempty"`
	SubAppid            *string              `json:"sub_appid,omitempty"`
	OutOrderNo          *string              `json:"out_order_no,omitempty"`
	ServiceIntroduction *string              `json:"service_introduction,omitempty"`
	PostPayments        []Payment            `json:"post_payments,omitempty"`
	PostDiscounts       []ServiceOrderCoupon `json:"post_discounts,omitempty"`
	RiskFund            *RiskFund            `json:"risk_fund,omitempty"`
	TimeRange           *TimeRange           `json:"time_range,omitempty"`
	Location            *Location            `json:"location,omitempty"`
	NeedUserConfirm     *bool                `json:"need_user_confirm,omitempty"`
	NotifyUrl           *string              `json:"notify_url,omitempty"`
	Attach              *string              `json:"attach,omitempty"`
	Device              *Device              `json:"device,omitempty"`
}

type CreatePartnerServiceOrderResponse struct {
	OutOrderNo          *string              `json:"out_order_no,omitempty"`
	ServiceId           *string              `json:"service_id,omitempty"`
	Appid               *string              `json:"appid,omitempty"`
	Mchid               *string              `json:"mchid,omitempty"`
	SubAppid            *string              `json:"sub_appid,omitempty"`
	SubMchid            *string              `json:"sub_mchid,omitempty"`
	ServiceIntroduction *string              `json:"service_introduction,omitempty"`
	State               *string              `json:"state,omitempty"`
	StateDescription    *string              `json:"state_description,omitempty"`
	PostPayments        []Payment            `json:"post_payments,omitempty"`
	PostDiscounts       []ServiceOrderCoupon `json:"post_discounts,omitempty"`
	RiskFund            *RiskFund            `json:"risk_fund,omitempty"`
	TimeRange           *TimeRange           `json:"time_range,omitempty"`
	Location            *Location            `json:"location,omitempty"`
	Attach              *string              `json:"attach,omitempty"`
	NotifyUrl           *string              `json:"notify_url,omitempty"`
	OrderId             *string              `json:"order_id,omitempty"`
	Package             *string              `json:"package,omitempty"`
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

type Device struct {
	StartDeviceId *string `json:"start_device_id,omitempty"`
	EndDeviceId   *string `json:"end_device_id,omitempty"`
	MaterielNo    *string `json:"materiel_no,omitempty"`
}

