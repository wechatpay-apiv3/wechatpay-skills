package main

import (
	"bytes"
	"demo/wxpay_utility" // 引用微信支付工具库，参考 https://pay.weixin.qq.com/doc/v3/partner/4015119446
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

	request := &ModifyPartnerServiceOrderRequest{
		OutOrderNo: wxpay_utility.String("1234323JKHDFE1243252"),
		ServiceId:  wxpay_utility.String("2002000000000558128851361561536"),
		SubMchid:   wxpay_utility.String("1900000109"),
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

	err = ModifyPartnerServiceOrder(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		// TODO: 请求失败，根据状态码执行不同的处理
		return
	}

	// TODO: 请求成功，继续业务逻辑
	fmt.Println("请求成功")
}

func ModifyPartnerServiceOrder(config *wxpay_utility.MchConfig, request *ModifyPartnerServiceOrderRequest) (err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/payscore/partner/serviceorder/{out_order_no}/modify"
	)

	reqUrl, err := url.Parse(fmt.Sprintf("%s%s", host, path))
	if err != nil {
		return err
	}
	reqUrl.Path = strings.Replace(reqUrl.Path, "{out_order_no}", url.PathEscape(*request.OutOrderNo), -1)
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
		// 2XX 成功，验证应答签名
		err = wxpay_utility.ValidateResponse(
			config.WechatPayPublicKeyId(),
			config.WechatPayPublicKey(),
			&httpResponse.Header,
			respBody,
		)
		if err != nil {
			return err
		}
		return nil
	} else {
		return wxpay_utility.NewApiException(
			httpResponse.StatusCode,
			httpResponse.Header,
			respBody,
		)
	}
}

type ModifyPartnerServiceOrderRequest struct {
	ServiceId     *string              `json:"service_id,omitempty"`
	SubMchid      *string              `json:"sub_mchid,omitempty"`
	OutOrderNo    *string              `json:"out_order_no,omitempty"`
	PostPayments  []Payment            `json:"post_payments,omitempty"`
	PostDiscounts []ServiceOrderCoupon `json:"post_discounts,omitempty"`
	TotalAmount   *int64               `json:"total_amount,omitempty"`
	Reason        *string              `json:"reason,omitempty"`
	Device        *Device              `json:"device,omitempty"`
}

func (o *ModifyPartnerServiceOrderRequest) MarshalJSON() ([]byte, error) {
	type Alias ModifyPartnerServiceOrderRequest
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

