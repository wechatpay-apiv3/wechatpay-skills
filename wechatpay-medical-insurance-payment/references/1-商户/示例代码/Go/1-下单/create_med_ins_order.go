package main

import (
	"bytes"
	"demo/wxpay_utility" // 引用微信支付工具库，参考 https://pay.weixin.qq.com/doc/v3/merchant/4015119334
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

	// 加密敏感字段：payer.name / payer.id_digest / relative.name / relative.id_digest
	encryptedName, _ := wxpay_utility.EncryptOAEPWithPublicKey("张三", config.WechatPayPublicKey())
	encryptedIdDigest, _ := wxpay_utility.EncryptOAEPWithPublicKey("09eb26e839ff3a2e3980352ae45ef09e", config.WechatPayPublicKey())

	request := &CreateOrderRequest{
		MixPayType: MIXPAYTYPE_CASH_AND_INSURANCE.Ptr(),
		OrderType:  ORDERTYPE_REG_PAY.Ptr(),
		Appid:      wxpay_utility.String("wxdace645e0bc2cXXX"),
		Openid:     wxpay_utility.String("o4GgauInH_RCEdvrrNGrntXDuXXX"),
		Payer: &PersonIdentification{
			Name:     wxpay_utility.String(encryptedName),
			IdDigest: wxpay_utility.String(encryptedIdDigest),
			CardType: USERCARDTYPE_ID_CARD.Ptr(),
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
			CashAddType: CASHADDTYPE_FREIGHT.Ptr(),
		}},
		CashReduceDetail: []CashReduceEntity{{
			CashReduceFee:  wxpay_utility.Int64(10000),
			CashReduceType: CASHREDUCETYPE_DEFAULT_REDUCE_TYPE.Ptr(),
		}},
		CallbackUrl:   wxpay_utility.String("https://www.weixin.qq.com/wxpay/pay.php"),
		PrepayId:      wxpay_utility.String("wx201410272009395522657a690389285100"),
		Attach:        wxpay_utility.String("{}"),
		MedInsTestEnv: wxpay_utility.Bool(false),
	}

	response, err := CreateOrder(config, request)
	if err != nil {
		fmt.Printf("请求失败: %+v\n", err)
		// TODO: 请求失败，根据状态码执行不同的处理
		return
	}

	// TODO: 请求成功，继续业务逻辑（response.MixTradeNo 用于后续调起支付/查询/退款通知）
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
		err = wxpay_utility.ValidateResponse(
			config.WechatPayPublicKeyId(),
			config.WechatPayPublicKey(),
			&httpResponse.Header,
			respBody,
		)
		if err != nil {
			return nil, err
		}
		response := &OrderEntity{}
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

type CreateOrderRequest struct {
	MixPayType                *MixPayType           `json:"mix_pay_type,omitempty"`
	OrderType                 *OrderType            `json:"order_type,omitempty"`
	Appid                     *string               `json:"appid,omitempty"`
	Openid                    *string               `json:"openid,omitempty"`
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
	MixTradeNo                 *string            `json:"mix_trade_no,omitempty"`
	MixPayStatus               *MixPayStatus      `json:"mix_pay_status,omitempty"`
	SelfPayStatus              *SelfPayStatus     `json:"self_pay_status,omitempty"`
	MedInsPayStatus            *MedInsPayStatus   `json:"med_ins_pay_status,omitempty"`
	PaidTime                   *string            `json:"paid_time,omitempty"`
	PassthroughResponseContent *string            `json:"passthrough_response_content,omitempty"`
	MixPayType                 *MixPayType        `json:"mix_pay_type,omitempty"`
	OrderType                  *OrderType         `json:"order_type,omitempty"`
	Appid                      *string            `json:"appid,omitempty"`
	Openid                     *string            `json:"openid,omitempty"`
	PayForRelatives            *bool              `json:"pay_for_relatives,omitempty"`
	OutTradeNo                 *string            `json:"out_trade_no,omitempty"`
	SerialNo                   *string            `json:"serial_no,omitempty"`
	PayOrderId                 *string            `json:"pay_order_id,omitempty"`
	PayAuthNo                  *string            `json:"pay_auth_no,omitempty"`
	GeoLocation                *string            `json:"geo_location,omitempty"`
	CityId                     *string            `json:"city_id,omitempty"`
	MedInstName                *string            `json:"med_inst_name,omitempty"`
	MedInstNo                  *string            `json:"med_inst_no,omitempty"`
	MedInsOrderCreateTime      *string            `json:"med_ins_order_create_time,omitempty"`
	TotalFee                   *int64             `json:"total_fee,omitempty"`
	MedInsGovFee               *int64             `json:"med_ins_gov_fee,omitempty"`
	MedInsSelfFee              *int64             `json:"med_ins_self_fee,omitempty"`
	MedInsOtherFee             *int64             `json:"med_ins_other_fee,omitempty"`
	MedInsCashFee              *int64             `json:"med_ins_cash_fee,omitempty"`
	WechatPayCashFee           *int64             `json:"wechat_pay_cash_fee,omitempty"`
	CashAddDetail              []CashAddEntity    `json:"cash_add_detail,omitempty"`
	CashReduceDetail           []CashReduceEntity `json:"cash_reduce_detail,omitempty"`
	CallbackUrl                *string            `json:"callback_url,omitempty"`
	PrepayId                   *string            `json:"prepay_id,omitempty"`
	PassthroughRequestContent  *string            `json:"passthrough_request_content,omitempty"`
	Extends                    *string            `json:"extends,omitempty"`
	Attach                     *string            `json:"attach,omitempty"`
	ChannelNo                  *string            `json:"channel_no,omitempty"`
	MedInsTestEnv              *bool              `json:"med_ins_test_env,omitempty"`
}

type MixPayType string

func (e MixPayType) Ptr() *MixPayType { return &e }

const (
	MIXPAYTYPE_CASH_ONLY          MixPayType = "CASH_ONLY"
	MIXPAYTYPE_INSURANCE_ONLY     MixPayType = "INSURANCE_ONLY"
	MIXPAYTYPE_CASH_AND_INSURANCE MixPayType = "CASH_AND_INSURANCE"
)

type OrderType string

func (e OrderType) Ptr() *OrderType { return &e }

const (
	ORDERTYPE_REG_PAY           OrderType = "REG_PAY"
	ORDERTYPE_DIAG_PAY          OrderType = "DIAG_PAY"
	ORDERTYPE_COVID_EXAM_PAY    OrderType = "COVID_EXAM_PAY"
	ORDERTYPE_IN_HOSP_PAY       OrderType = "IN_HOSP_PAY"
	ORDERTYPE_PHARMACY_PAY      OrderType = "PHARMACY_PAY"
	ORDERTYPE_INSURANCE_PAY     OrderType = "INSURANCE_PAY"
	ORDERTYPE_INT_REG_PAY       OrderType = "INT_REG_PAY"
	ORDERTYPE_INT_RE_DIAG_PAY   OrderType = "INT_RE_DIAG_PAY"
	ORDERTYPE_INT_RX_PAY        OrderType = "INT_RX_PAY"
	ORDERTYPE_COVID_ANTIGEN_PAY OrderType = "COVID_ANTIGEN_PAY"
	ORDERTYPE_MED_PAY           OrderType = "MED_PAY"
)

type PersonIdentification struct {
	Name     *string       `json:"name,omitempty"`
	IdDigest *string       `json:"id_digest,omitempty"`
	CardType *UserCardType `json:"card_type,omitempty"`
}

type CashAddEntity struct {
	CashAddFee  *int64       `json:"cash_add_fee,omitempty"`
	CashAddType *CashAddType `json:"cash_add_type,omitempty"`
}

type CashReduceEntity struct {
	CashReduceFee  *int64          `json:"cash_reduce_fee,omitempty"`
	CashReduceType *CashReduceType `json:"cash_reduce_type,omitempty"`
}

type MixPayStatus string

func (e MixPayStatus) Ptr() *MixPayStatus { return &e }

const (
	MIXPAYSTATUS_MIX_PAY_CREATED MixPayStatus = "MIX_PAY_CREATED"
	MIXPAYSTATUS_MIX_PAY_SUCCESS MixPayStatus = "MIX_PAY_SUCCESS"
	MIXPAYSTATUS_MIX_PAY_REFUND  MixPayStatus = "MIX_PAY_REFUND"
	MIXPAYSTATUS_MIX_PAY_FAIL    MixPayStatus = "MIX_PAY_FAIL"
)

type SelfPayStatus string

func (e SelfPayStatus) Ptr() *SelfPayStatus { return &e }

const (
	SELFPAYSTATUS_SELF_PAY_CREATED SelfPayStatus = "SELF_PAY_CREATED"
	SELFPAYSTATUS_SELF_PAY_SUCCESS SelfPayStatus = "SELF_PAY_SUCCESS"
	SELFPAYSTATUS_SELF_PAY_REFUND  SelfPayStatus = "SELF_PAY_REFUND"
	SELFPAYSTATUS_SELF_PAY_FAIL    SelfPayStatus = "SELF_PAY_FAIL"
	SELFPAYSTATUS_NO_SELF_PAY      SelfPayStatus = "NO_SELF_PAY"
)

type MedInsPayStatus string

func (e MedInsPayStatus) Ptr() *MedInsPayStatus { return &e }

const (
	MEDINSPAYSTATUS_MED_INS_PAY_CREATED MedInsPayStatus = "MED_INS_PAY_CREATED"
	MEDINSPAYSTATUS_MED_INS_PAY_SUCCESS MedInsPayStatus = "MED_INS_PAY_SUCCESS"
	MEDINSPAYSTATUS_MED_INS_PAY_REFUND  MedInsPayStatus = "MED_INS_PAY_REFUND"
	MEDINSPAYSTATUS_MED_INS_PAY_FAIL    MedInsPayStatus = "MED_INS_PAY_FAIL"
	MEDINSPAYSTATUS_NO_MED_INS_PAY      MedInsPayStatus = "NO_MED_INS_PAY"
)

type UserCardType string

func (e UserCardType) Ptr() *UserCardType { return &e }

const (
	USERCARDTYPE_ID_CARD                       UserCardType = "ID_CARD"
	USERCARDTYPE_HOUSEHOLD_REGISTRATION        UserCardType = "HOUSEHOLD_REGISTRATION"
	USERCARDTYPE_FOREIGNER_PASSPORT            UserCardType = "FOREIGNER_PASSPORT"
	USERCARDTYPE_MAINLAND_TRAVEL_PERMIT_FOR_TW UserCardType = "MAINLAND_TRAVEL_PERMIT_FOR_TW"
	USERCARDTYPE_MAINLAND_TRAVEL_PERMIT_FOR_MO UserCardType = "MAINLAND_TRAVEL_PERMIT_FOR_MO"
	USERCARDTYPE_MAINLAND_TRAVEL_PERMIT_FOR_HK UserCardType = "MAINLAND_TRAVEL_PERMIT_FOR_HK"
	USERCARDTYPE_FOREIGN_PERMANENT_RESIDENT    UserCardType = "FOREIGN_PERMANENT_RESIDENT"
)

type CashAddType string

func (e CashAddType) Ptr() *CashAddType { return &e }

const (
	CASHADDTYPE_DEFAULT_ADD_TYPE       CashAddType = "DEFAULT_ADD_TYPE"
	CASHADDTYPE_FREIGHT                CashAddType = "FREIGHT"
	CASHADDTYPE_OTHER_MEDICAL_EXPENSES CashAddType = "OTHER_MEDICAL_EXPENSES"
)

type CashReduceType string

func (e CashReduceType) Ptr() *CashReduceType { return &e }

const (
	CASHREDUCETYPE_DEFAULT_REDUCE_TYPE CashReduceType = "DEFAULT_REDUCE_TYPE"
	CASHREDUCETYPE_HOSPITAL_REDUCE     CashReduceType = "HOSPITAL_REDUCE"
	CASHREDUCETYPE_PHARMACY_DISCOUNT   CashReduceType = "PHARMACY_DISCOUNT"
	CASHREDUCETYPE_DISCOUNT            CashReduceType = "DISCOUNT"
	CASHREDUCETYPE_PRE_PAYMENT         CashReduceType = "PRE_PAYMENT"
	CASHREDUCETYPE_DEPOSIT_DEDUCTION   CashReduceType = "DEPOSIT_DEDUCTION"
)
