package com.java.demo;

import com.java.utils.WXPayUtility; // 引用微信支付工具库，参考：https://pay.weixin.qq.com/doc/v3/merchant/4014931831

import com.google.gson.annotations.SerializedName;
import com.google.gson.annotations.Expose;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 医保自费混合收款下单
 */
public class CreateMedInsOrder {
  private static String HOST = "https://api.mch.weixin.qq.com";
  private static String METHOD = "POST";
  private static String PATH = "/v3/med-ins/orders";

  public static void main(String[] args) {
    // TODO: 请准备商户开发必要参数，参考：https://pay.weixin.qq.com/doc/v3/merchant/4013070756
    CreateMedInsOrder client = new CreateMedInsOrder(
      "19xxxxxxxx",                    // 商户号，是由微信支付系统生成并分配给每个商户的唯一标识符，商户号获取方式参考 https://pay.weixin.qq.com/doc/v3/merchant/4013070756
      "1DDE55AD98Exxxxxxxxxx",         // 商户API证书序列号，如何获取请参考 https://pay.weixin.qq.com/doc/v3/merchant/4013053053
      "/path/to/apiclient_key.pem",     // 商户API证书私钥文件路径，本地文件路径
      "PUB_KEY_ID_xxxxxxxxxxxxx",      // 微信支付公钥ID，如何获取请参考 https://pay.weixin.qq.com/doc/v3/merchant/4013038816
      "/path/to/wxp_pub.pem"           // 微信支付公钥文件路径，本地文件路径
    );

    CreateOrderRequest request = new CreateOrderRequest();
    request.mixPayType = MixPayType.CASH_AND_INSURANCE;
    request.orderType = OrderType.REG_PAY;
    request.appid = "wxdace645e0bc2cXXX";
    request.openid = "o4GgauInH_RCEdvrrNGrntXDuXXX";
    request.payer = new PersonIdentification();
    request.payer.name = client.encrypt("张三");
    request.payer.idDigest = client.encrypt("09eb26e839ff3a2e3980352ae45ef09e");
    request.payer.cardType = UserCardType.ID_CARD;
    request.payForRelatives = false;
    request.outTradeNo = "202204022005169952975171534816";
    request.serialNo = "1217752501201";
    request.payOrderId = "ORD530100202204022006350000021";
    request.payAuthNo = "AUTH530100202204022006310000034";
    request.geoLocation = "102.682296,25.054260";
    request.cityId = "530100";
    request.medInstName = "北大医院";
    request.medInstNo = "1217752501201407033233368318";
    request.medInsOrderCreateTime = "2015-05-20T13:29:35+08:00";
    request.totalFee = 202000L;
    request.medInsGovFee = 100000L;
    request.medInsSelfFee = 45000L;
    request.medInsOtherFee = 5000L;
    request.medInsCashFee = 50000L;
    request.wechatPayCashFee = 42000L;
    request.cashAddDetail = new ArrayList<>();
    {
      CashAddEntity cashAddDetailItem = new CashAddEntity();
      cashAddDetailItem.cashAddFee = 2000L;
      cashAddDetailItem.cashAddType = CashAddType.FREIGHT;
      request.cashAddDetail.add(cashAddDetailItem);
    };
    request.cashReduceDetail = new ArrayList<>();
    {
      CashReduceEntity cashReduceDetailItem = new CashReduceEntity();
      cashReduceDetailItem.cashReduceFee = 10000L;
      cashReduceDetailItem.cashReduceType = CashReduceType.DEFAULT_REDUCE_TYPE;
      request.cashReduceDetail.add(cashReduceDetailItem);
    };
    request.callbackUrl = "https://www.weixin.qq.com/wxpay/pay.php";
    request.prepayId = "wx201410272009395522657a690389285100";
    request.attach = "{}";
    request.medInsTestEnv = false;
    try {
      OrderEntity response = client.run(request);
        // TODO: 请求成功，继续业务逻辑（response.mixTradeNo 用于后续调起支付/查询/退款通知）
        System.out.println(response);
    } catch (WXPayUtility.ApiException e) {
        // TODO: 请求失败，根据状态码执行不同的逻辑
        e.printStackTrace();
    }
  }

  public OrderEntity run(CreateOrderRequest request) {
    String uri = PATH;
    String reqBody = WXPayUtility.toJson(request);

    Request.Builder reqBuilder = new Request.Builder().url(HOST + uri);
    reqBuilder.addHeader("Accept", "application/json");
    reqBuilder.addHeader("Wechatpay-Serial", wechatPayPublicKeyId);
    reqBuilder.addHeader("Authorization", WXPayUtility.buildAuthorization(mchid, certificateSerialNo, privateKey, METHOD, uri, reqBody));
    reqBuilder.addHeader("Content-Type", "application/json");
    RequestBody requestBody = RequestBody.create(MediaType.parse("application/json; charset=utf-8"), reqBody);
    reqBuilder.method(METHOD, requestBody);
    Request httpRequest = reqBuilder.build();

    OkHttpClient client = new OkHttpClient.Builder().build();
    try (Response httpResponse = client.newCall(httpRequest).execute()) {
      String respBody = WXPayUtility.extractBody(httpResponse);
      if (httpResponse.code() >= 200 && httpResponse.code() < 300) {
        WXPayUtility.validateResponse(this.wechatPayPublicKeyId, this.wechatPayPublicKey,
            httpResponse.headers(), respBody);
        return WXPayUtility.fromJson(respBody, OrderEntity.class);
      } else {
        throw new WXPayUtility.ApiException(httpResponse.code(), respBody, httpResponse.headers());
      }
    } catch (IOException e) {
      throw new UncheckedIOException("Sending request to " + uri + " failed.", e);
    }
  }

  private final String mchid;
  private final String certificateSerialNo;
  private final PrivateKey privateKey;
  private final String wechatPayPublicKeyId;
  private final PublicKey wechatPayPublicKey;

  public CreateMedInsOrder(String mchid, String certificateSerialNo, String privateKeyFilePath, String wechatPayPublicKeyId, String wechatPayPublicKeyFilePath) {
    this.mchid = mchid;
    this.certificateSerialNo = certificateSerialNo;
    this.privateKey = WXPayUtility.loadPrivateKeyFromPath(privateKeyFilePath);
    this.wechatPayPublicKeyId = wechatPayPublicKeyId;
    this.wechatPayPublicKey = WXPayUtility.loadPublicKeyFromPath(wechatPayPublicKeyFilePath);
  }

  /** 加密敏感字段（payer.name / payer.id_digest / relative.name / relative.id_digest 必须加密后再传） */
  public String encrypt(String plainText) {
    return WXPayUtility.encrypt(this.wechatPayPublicKey, plainText);
  }

  public static class CreateOrderRequest {
    @SerializedName("mix_pay_type") public MixPayType mixPayType;
    @SerializedName("order_type") public OrderType orderType;
    @SerializedName("appid") public String appid;
    @SerializedName("openid") public String openid;
    @SerializedName("payer") public PersonIdentification payer;
    @SerializedName("pay_for_relatives") public Boolean payForRelatives;
    @SerializedName("relative") public PersonIdentification relative;
    @SerializedName("out_trade_no") public String outTradeNo;
    @SerializedName("serial_no") public String serialNo;
    @SerializedName("pay_order_id") public String payOrderId;
    @SerializedName("pay_auth_no") public String payAuthNo;
    @SerializedName("geo_location") public String geoLocation;
    @SerializedName("city_id") public String cityId;
    @SerializedName("med_inst_name") public String medInstName;
    @SerializedName("med_inst_no") public String medInstNo;
    @SerializedName("med_ins_order_create_time") public String medInsOrderCreateTime;
    @SerializedName("total_fee") public Long totalFee;
    @SerializedName("med_ins_gov_fee") public Long medInsGovFee;
    @SerializedName("med_ins_self_fee") public Long medInsSelfFee;
    @SerializedName("med_ins_other_fee") public Long medInsOtherFee;
    @SerializedName("med_ins_cash_fee") public Long medInsCashFee;
    @SerializedName("wechat_pay_cash_fee") public Long wechatPayCashFee;
    @SerializedName("cash_add_detail") public List<CashAddEntity> cashAddDetail;
    @SerializedName("cash_reduce_detail") public List<CashReduceEntity> cashReduceDetail;
    @SerializedName("callback_url") public String callbackUrl;
    @SerializedName("prepay_id") public String prepayId;
    @SerializedName("passthrough_request_content") public String passthroughRequestContent;
    @SerializedName("extends") public String _extends;
    @SerializedName("attach") public String attach;
    @SerializedName("channel_no") public String channelNo;
    @SerializedName("med_ins_test_env") public Boolean medInsTestEnv;
  }

  public static class OrderEntity {
    @SerializedName("mix_trade_no") public String mixTradeNo;
    @SerializedName("mix_pay_status") public MixPayStatus mixPayStatus;
    @SerializedName("self_pay_status") public SelfPayStatus selfPayStatus;
    @SerializedName("med_ins_pay_status") public MedInsPayStatus medInsPayStatus;
    @SerializedName("paid_time") public String paidTime;
    @SerializedName("passthrough_response_content") public String passthroughResponseContent;
    @SerializedName("mix_pay_type") public MixPayType mixPayType;
    @SerializedName("order_type") public OrderType orderType;
    @SerializedName("appid") public String appid;
    @SerializedName("openid") public String openid;
    @SerializedName("pay_for_relatives") public Boolean payForRelatives;
    @SerializedName("out_trade_no") public String outTradeNo;
    @SerializedName("serial_no") public String serialNo;
    @SerializedName("pay_order_id") public String payOrderId;
    @SerializedName("pay_auth_no") public String payAuthNo;
    @SerializedName("geo_location") public String geoLocation;
    @SerializedName("city_id") public String cityId;
    @SerializedName("med_inst_name") public String medInstName;
    @SerializedName("med_inst_no") public String medInstNo;
    @SerializedName("med_ins_order_create_time") public String medInsOrderCreateTime;
    @SerializedName("total_fee") public Long totalFee;
    @SerializedName("med_ins_gov_fee") public Long medInsGovFee;
    @SerializedName("med_ins_self_fee") public Long medInsSelfFee;
    @SerializedName("med_ins_other_fee") public Long medInsOtherFee;
    @SerializedName("med_ins_cash_fee") public Long medInsCashFee;
    @SerializedName("wechat_pay_cash_fee") public Long wechatPayCashFee;
    @SerializedName("cash_add_detail") public List<CashAddEntity> cashAddDetail;
    @SerializedName("cash_reduce_detail") public List<CashReduceEntity> cashReduceDetail;
    @SerializedName("callback_url") public String callbackUrl;
    @SerializedName("prepay_id") public String prepayId;
    @SerializedName("passthrough_request_content") public String passthroughRequestContent;
    @SerializedName("extends") public String _extends;
    @SerializedName("attach") public String attach;
    @SerializedName("channel_no") public String channelNo;
    @SerializedName("med_ins_test_env") public Boolean medInsTestEnv;
  }

  public enum MixPayType {
    @SerializedName("CASH_ONLY") CASH_ONLY,
    @SerializedName("INSURANCE_ONLY") INSURANCE_ONLY,
    @SerializedName("CASH_AND_INSURANCE") CASH_AND_INSURANCE
  }

  public enum OrderType {
    @SerializedName("REG_PAY") REG_PAY,
    @SerializedName("DIAG_PAY") DIAG_PAY,
    @SerializedName("COVID_EXAM_PAY") COVID_EXAM_PAY,
    @SerializedName("IN_HOSP_PAY") IN_HOSP_PAY,
    @SerializedName("PHARMACY_PAY") PHARMACY_PAY,
    @SerializedName("INSURANCE_PAY") INSURANCE_PAY,
    @SerializedName("INT_REG_PAY") INT_REG_PAY,
    @SerializedName("INT_RE_DIAG_PAY") INT_RE_DIAG_PAY,
    @SerializedName("INT_RX_PAY") INT_RX_PAY,
    @SerializedName("COVID_ANTIGEN_PAY") COVID_ANTIGEN_PAY,
    @SerializedName("MED_PAY") MED_PAY
  }

  public static class PersonIdentification {
    @SerializedName("name") public String name;
    @SerializedName("id_digest") public String idDigest;
    @SerializedName("card_type") public UserCardType cardType;
  }

  public static class CashAddEntity {
    @SerializedName("cash_add_fee") public Long cashAddFee;
    @SerializedName("cash_add_type") public CashAddType cashAddType;
  }

  public static class CashReduceEntity {
    @SerializedName("cash_reduce_fee") public Long cashReduceFee;
    @SerializedName("cash_reduce_type") public CashReduceType cashReduceType;
  }

  public enum MixPayStatus {
    @SerializedName("MIX_PAY_CREATED") MIX_PAY_CREATED,
    @SerializedName("MIX_PAY_SUCCESS") MIX_PAY_SUCCESS,
    @SerializedName("MIX_PAY_REFUND") MIX_PAY_REFUND,
    @SerializedName("MIX_PAY_FAIL") MIX_PAY_FAIL
  }

  public enum SelfPayStatus {
    @SerializedName("SELF_PAY_CREATED") SELF_PAY_CREATED,
    @SerializedName("SELF_PAY_SUCCESS") SELF_PAY_SUCCESS,
    @SerializedName("SELF_PAY_REFUND") SELF_PAY_REFUND,
    @SerializedName("SELF_PAY_FAIL") SELF_PAY_FAIL,
    @SerializedName("NO_SELF_PAY") NO_SELF_PAY
  }

  public enum MedInsPayStatus {
    @SerializedName("MED_INS_PAY_CREATED") MED_INS_PAY_CREATED,
    @SerializedName("MED_INS_PAY_SUCCESS") MED_INS_PAY_SUCCESS,
    @SerializedName("MED_INS_PAY_REFUND") MED_INS_PAY_REFUND,
    @SerializedName("MED_INS_PAY_FAIL") MED_INS_PAY_FAIL,
    @SerializedName("NO_MED_INS_PAY") NO_MED_INS_PAY
  }

  public enum UserCardType {
    @SerializedName("ID_CARD") ID_CARD,
    @SerializedName("HOUSEHOLD_REGISTRATION") HOUSEHOLD_REGISTRATION,
    @SerializedName("FOREIGNER_PASSPORT") FOREIGNER_PASSPORT,
    @SerializedName("MAINLAND_TRAVEL_PERMIT_FOR_TW") MAINLAND_TRAVEL_PERMIT_FOR_TW,
    @SerializedName("MAINLAND_TRAVEL_PERMIT_FOR_MO") MAINLAND_TRAVEL_PERMIT_FOR_MO,
    @SerializedName("MAINLAND_TRAVEL_PERMIT_FOR_HK") MAINLAND_TRAVEL_PERMIT_FOR_HK,
    @SerializedName("FOREIGN_PERMANENT_RESIDENT") FOREIGN_PERMANENT_RESIDENT
  }

  public enum CashAddType {
    @SerializedName("DEFAULT_ADD_TYPE") DEFAULT_ADD_TYPE,
    @SerializedName("FREIGHT") FREIGHT,
    @SerializedName("OTHER_MEDICAL_EXPENSES") OTHER_MEDICAL_EXPENSES
  }

  public enum CashReduceType {
    @SerializedName("DEFAULT_REDUCE_TYPE") DEFAULT_REDUCE_TYPE,
    @SerializedName("HOSPITAL_REDUCE") HOSPITAL_REDUCE,
    @SerializedName("PHARMACY_DISCOUNT") PHARMACY_DISCOUNT,
    @SerializedName("DISCOUNT") DISCOUNT,
    @SerializedName("PRE_PAYMENT") PRE_PAYMENT,
    @SerializedName("DEPOSIT_DEDUCTION") DEPOSIT_DEDUCTION
  }
}
