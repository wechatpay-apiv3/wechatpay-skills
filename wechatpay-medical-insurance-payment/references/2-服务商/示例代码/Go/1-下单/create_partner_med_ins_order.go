package main

import (
	"bytes"
	"demo/wxpay_utility"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// 服务商医保自费混合收款下单（同时适用于服务商模式与间连模式）
//
// 与商户版的差异：
//   - 必传 sub_mchid + sub_appid（医疗机构商户号 + AppID）
//   - openid 与 sub_openid 二选一：
//       openid    → 调起时用 appid（服务商 AppID）
//       sub_openid → 调起时用 sub_appid（医疗机构 AppID）
//   - 签名仍使用服务商 API 证书私钥，Wechatpay-Serial 仍是服务商微信支付公钥 ID
func main() {
	config, err := wxpay_utility.CreateMchConfig(
		"1900000100",                 // 服务商商户号
		"1DDE55AD98Exxxxxxxxxx",      // 服务商 API 证书序列号
		"/path/to/apiclient_key.pem", // 服务商 API 证书私钥路径
		"PUB_KEY_ID_xxxxxxxxxxxxx",   // 服务商微信支付公钥 ID
		"/path/to/wxp_pub.pem",       // 服务商微信支付公钥路径
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	encryptedName, _ := wxpay_utility.EncryptOAEPWithPublicKey("张三", config.WechatPayPublicKey())
	encryptedIdDigest, _ := wxpay_utility.EncryptOAEPWithPublicKey("09eb26e839ff3a2e3980352ae45ef09e", config.WechatPayPublicKey())

	request := &CreateOrderRequest{
		MixPayType: wxpay_utility.String("CASH_AND_INSURANCE"),
		OrderType:  wxpay_utility.String("REG_PAY"),
		Appid:      wxpay_utility.String("wxdace645e0bc2cXXX"),
		SubAppid:   wxpay_utility.String("wxdace645e0bc2cYYY"),
		SubMchid:   wxpay_utility.String("1900000109"),
		SubOpenid:  wxpay_utility.String("o4GgauInH_RCEdvrrNGrntXDuXXX"),
		Payer: &PersonIdentification{
			Name:     wxpay_utility.String(encryptedName),
			IdDigest: wxpay_utility.String(encryptedIdDigest),
			CardType: wxpay_utility.String("ID_CARD"),
		},
		PayForRelatives:       wxpay_utility.Bool(false),
		OutTradeNo:            wxpay_utility.String("202204022005169952975171534816"),
		SerialNo:              wxpay_utility.String("1217752501201"),
		PayOrderId:            wxpay_utility.String("ORD530100202204022006350000021"),
		PayAuthNo:             wxpay_utility.String("AUTH530100202204022006310000034"),
		GeoLocation:           wxpay_utility.String("102.682296,25.054260"),
		CityId:                wxpay_utility.String("530100"),
		MedInstName:           wxpay_utility.String("北大医院"),
		MedInstNo:             wxpay_utility.String("1217752501201407033233368318"),
		MedInsOrderCreateTime: wxpay_utility.String("2015-05-20T13:29:35+08:00"),
		TotalFee:              wxpay_utility.Int64(202000),
		MedInsGovFee:          wxpay_utility.Int64(100000),
		MedInsSelfFee:         wxpay_utility.Int64(45000),
		MedInsOtherFee:        wxpay_utility.Int64(5000),
		MedInsCashFee:         wxpay_utility.Int64(50000),
		WechatPayCashFee:      wxpay_utility.Int64(42000),
		CashAddDetail: []CashAddEntity{{
			CashAddFee:  wxpay_utility.Int64(2000),
			CashAddType: wxpay_utility.String("FREIGHT"),
		}},
		CashReduceDetail: []CashReduceEntity{{
			CashReduceFee:  wxpay_utility.Int64(10000),
			CashReduceType: wxpay_utility.String("DEFAULT_REDUCE_TYPE"),
		}},
		CallbackUrl:   wxpay_utility.String("https://www.weixin.qq.com/wxpay/pay.php"),
		PrepayId:      wxpay_utility.String("wx201410272009395522657a690389285100"),
		MedInsTestEnv: wxpay_utility.Bool(false),
	}

	response, err := CreateOrder(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		return
	}
	fmt.Printf("请求成功: %+v\n", response)
}

func CreateOrder(config *wxpay_utility.MchConfig, request *CreateOrderRequest) (response *OrderEntity, err error) {
	const (
		host   = "https://api.mch.weixin.qq.com"
		method = "POST"
		path   = "/v3/med-ins/orders"
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

type CreateOrderRequest struct {
	MixPayType                *string               `json:"mix_pay_type,omitempty"`
	OrderType                 *string               `json:"order_type,omitempty"`
	Appid                     *string               `json:"appid,omitempty"`
	SubAppid                  *string               `json:"sub_appid,omitempty"`
	SubMchid                  *string               `json:"sub_mchid,omitempty"`
	Openid                    *string               `json:"openid,omitempty"`
	SubOpenid                 *string               `json:"sub_openid,omitempty"`
	Payer                     *PersonIdentification `json:"payer,omitempty"`
	PayForRelatives           *bool                 `json:"pay_for_relatives,omitempty"`
	Relative                  *PersonIdentification `json:"relative,omitempty"`
	OutTradeNo                *string               `json:"out_trade_no,omitempty"`
	SerialNo                  *string               `json:"serial_no,omitempty"`
	PayOrderId                *string               `json:"pay_order_id,omitempty"`
	PayAuthNo                 *string               `json:"pay_auth_no,omitempty"`
	GeoLocation               *string               `json:"geo_location,omitempty"`
	CityId                    *string               `json:"city_id,omitempty"`
	MedInstName               *string               `json:"med_inst_name,omitempty"`
	MedInstNo                 *string               `json:"med_inst_no,omitempty"`
	MedInsOrderCreateTime     *string               `json:"med_ins_order_create_time,omitempty"`
	TotalFee                  *int64                `json:"total_fee,omitempty"`
	MedInsGovFee              *int64                `json:"med_ins_gov_fee,omitempty"`
	MedInsSelfFee             *int64                `json:"med_ins_self_fee,omitempty"`
	MedInsOtherFee            *int64                `json:"med_ins_other_fee,omitempty"`
	MedInsCashFee             *int64                `json:"med_ins_cash_fee,omitempty"`
	WechatPayCashFee          *int64                `json:"wechat_pay_cash_fee,omitempty"`
	CashAddDetail             []CashAddEntity       `json:"cash_add_detail,omitempty"`
	CashReduceDetail          []CashReduceEntity    `json:"cash_reduce_detail,omitempty"`
	CallbackUrl               *string               `json:"callback_url,omitempty"`
	PrepayId                  *string               `json:"prepay_id,omitempty"`
	PassthroughRequestContent *string               `json:"passthrough_request_content,omitempty"`
	Extends                   *string               `json:"extends,omitempty"`
	Attach                    *string               `json:"attach,omitempty"`
	ChannelNo                 *string               `json:"channel_no,omitempty"`
	MedInsTestEnv             *bool                 `json:"med_ins_test_env,omitempty"`
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
	CallbackUrl      *string `json:"callback_url,omitempty"`
	PrepayId         *string `json:"prepay_id,omitempty"`
}

type PersonIdentification struct {
	Name     *string `json:"name,omitempty"`
	IdDigest *string `json:"id_digest,omitempty"`
	CardType *string `json:"card_type,omitempty"`
}

type CashAddEntity struct {
	CashAddFee  *int64  `json:"cash_add_fee,omitempty"`
	CashAddType *string `json:"cash_add_type,omitempty"`
}

type CashReduceEntity struct {
	CashReduceFee  *int64  `json:"cash_reduce_fee,omitempty"`
	CashReduceType *string `json:"cash_reduce_type,omitempty"`
}
