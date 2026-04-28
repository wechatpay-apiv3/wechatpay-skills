package com.java.demo;

import com.java.utils.WXPayUtility;

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
import java.util.List;

/**
 * 服务商医保自费混合收款下单（同时适用于服务商模式与间连模式）
 *
 * 与商户版的差异：
 *   - 必传 sub_mchid + sub_appid（医疗机构商户号 + AppID）
 *   - openid 与 sub_openid 二选一：
 *       openid    → 调起时用 appid（服务商 AppID）
 *       sub_openid → 调起时用 sub_appid（医疗机构 AppID）
 *   - 签名仍使用服务商 API 证书私钥，Wechatpay-Serial 仍是服务商微信支付公钥 ID
 */
public class CreatePartnerMedInsOrder {
  private static String HOST = "https://api.mch.weixin.qq.com";
  private static String METHOD = "POST";
  private static String PATH = "/v3/med-ins/orders";

  public static void main(String[] args) {
    CreatePartnerMedInsOrder client = new CreatePartnerMedInsOrder(
      "1900000100",                     // 服务商商户号
      "1DDE55AD98Exxxxxxxxxx",          // 服务商 API 证书序列号
      "/path/to/apiclient_key.pem",     // 服务商 API 证书私钥路径
      "PUB_KEY_ID_xxxxxxxxxxxxx",       // 服务商微信支付公钥 ID
      "/path/to/wxp_pub.pem"            // 服务商微信支付公钥路径
    );

    CreateOrderRequest request = new CreateOrderRequest();
    request.mixPayType = "CASH_AND_INSURANCE";
    request.orderType = "REG_PAY";
    request.appid = "wxdace645e0bc2cXXX";          // 服务商 AppID
    request.subAppid = "wxdace645e0bc2cYYY";       // 医疗机构 AppID
    request.subMchid = "1900000109";               // 医疗机构商户号
    request.subOpenid = "o4GgauInH_RCEdvrrNGrntXDuXXX";  // 与 openid 二选一
    request.payer = new PersonIdentification();
    request.payer.name = client.encrypt("张三");
    request.payer.idDigest = client.encrypt("09eb26e839ff3a2e3980352ae45ef09e");
    request.payer.cardType = "ID_CARD";
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
      CashAddEntity item = new CashAddEntity();
      item.cashAddFee = 2000L;
      item.cashAddType = "FREIGHT";
      request.cashAddDetail.add(item);
    }
    request.cashReduceDetail = new ArrayList<>();
    {
      CashReduceEntity item = new CashReduceEntity();
      item.cashReduceFee = 10000L;
      item.cashReduceType = "DEFAULT_REDUCE_TYPE";
      request.cashReduceDetail.add(item);
    }
    request.callbackUrl = "https://www.weixin.qq.com/wxpay/pay.php";
    request.prepayId = "wx201410272009395522657a690389285100";
    request.medInsTestEnv = false;
    try {
      OrderEntity response = client.run(request);
      System.out.println(response);
    } catch (WXPayUtility.ApiException e) {
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
        WXPayUtility.validateResponse(this.wechatPayPublicKeyId, this.wechatPayPublicKey, httpResponse.headers(), respBody);
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

  public CreatePartnerMedInsOrder(String mchid, String certificateSerialNo, String privateKeyFilePath, String wechatPayPublicKeyId, String wechatPayPublicKeyFilePath) {
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
    @SerializedName("mix_pay_type") public String mixPayType;
    @SerializedName("order_type") public String orderType;
    @SerializedName("appid") public String appid;
    @SerializedName("sub_appid") public String subAppid;
    @SerializedName("sub_mchid") public String subMchid;
    @SerializedName("openid") public String openid;
    @SerializedName("sub_openid") public String subOpenid;
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
    @SerializedName("mix_pay_status") public String mixPayStatus;
    @SerializedName("self_pay_status") public String selfPayStatus;
    @SerializedName("med_ins_pay_status") public String medInsPayStatus;
    @SerializedName("paid_time") public String paidTime;
    @SerializedName("mix_pay_type") public String mixPayType;
    @SerializedName("order_type") public String orderType;
    @SerializedName("appid") public String appid;
    @SerializedName("sub_appid") public String subAppid;
    @SerializedName("sub_mchid") public String subMchid;
    @SerializedName("openid") public String openid;
    @SerializedName("sub_openid") public String subOpenid;
    @SerializedName("out_trade_no") public String outTradeNo;
    @SerializedName("total_fee") public Long totalFee;
    @SerializedName("med_ins_gov_fee") public Long medInsGovFee;
    @SerializedName("med_ins_self_fee") public Long medInsSelfFee;
    @SerializedName("med_ins_other_fee") public Long medInsOtherFee;
    @SerializedName("med_ins_cash_fee") public Long medInsCashFee;
    @SerializedName("wechat_pay_cash_fee") public Long wechatPayCashFee;
    @SerializedName("callback_url") public String callbackUrl;
    @SerializedName("prepay_id") public String prepayId;
  }

  public static class PersonIdentification {
    @SerializedName("name") public String name;
    @SerializedName("id_digest") public String idDigest;
    @SerializedName("card_type") public String cardType;
  }

  public static class CashAddEntity {
    @SerializedName("cash_add_fee") public Long cashAddFee;
    @SerializedName("cash_add_type") public String cashAddType;
  }

  public static class CashReduceEntity {
    @SerializedName("cash_reduce_fee") public Long cashReduceFee;
    @SerializedName("cash_reduce_type") public String cashReduceType;
  }
}
