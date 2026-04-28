package com.java.demo;

import com.java.utils.WXPayUtility;

import com.google.gson.annotations.SerializedName;
import com.google.gson.annotations.Expose;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.security.PrivateKey;
import java.security.PublicKey;

/**
 * 使用商户订单号（out_trade_no）查看医保订单结果
 */
public class QueryByOutTradeNo {
  private static String HOST = "https://api.mch.weixin.qq.com";
  private static String METHOD = "GET";
  private static String PATH = "/v3/med-ins/orders/out-trade-no/{out_trade_no}";

  public static void main(String[] args) {
    QueryByOutTradeNo client = new QueryByOutTradeNo(
      "19xxxxxxxx",
      "1DDE55AD98Exxxxxxxxxx",
      "/path/to/apiclient_key.pem",
      "PUB_KEY_ID_xxxxxxxxxxxxx",
      "/path/to/wxp_pub.pem"
    );

    QueryRequest request = new QueryRequest();
    request.outTradeNo = "202204022005169952975171534816";
    try {
      OrderEntity response = client.run(request);
      System.out.println(response);
    } catch (WXPayUtility.ApiException e) {
      e.printStackTrace();
    }
  }

  public OrderEntity run(QueryRequest request) {
    String uri = PATH.replace("{out_trade_no}", WXPayUtility.urlEncode(request.outTradeNo));

    Request.Builder reqBuilder = new Request.Builder().url(HOST + uri);
    reqBuilder.addHeader("Accept", "application/json");
    reqBuilder.addHeader("Wechatpay-Serial", wechatPayPublicKeyId);
    reqBuilder.addHeader("Authorization", WXPayUtility.buildAuthorization(mchid, certificateSerialNo, privateKey, METHOD, uri, null));
    reqBuilder.method(METHOD, null);
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

  public QueryByOutTradeNo(String mchid, String certificateSerialNo, String privateKeyFilePath, String wechatPayPublicKeyId, String wechatPayPublicKeyFilePath) {
    this.mchid = mchid;
    this.certificateSerialNo = certificateSerialNo;
    this.privateKey = WXPayUtility.loadPrivateKeyFromPath(privateKeyFilePath);
    this.wechatPayPublicKeyId = wechatPayPublicKeyId;
    this.wechatPayPublicKey = WXPayUtility.loadPublicKeyFromPath(wechatPayPublicKeyFilePath);
  }

  public static class QueryRequest {
    @SerializedName("out_trade_no")
    @Expose(serialize = false)
    public String outTradeNo;
  }

  /** 应答结构与 QueryByMixTradeNo 一致 */
  public static class OrderEntity {
    @SerializedName("mix_trade_no") public String mixTradeNo;
    @SerializedName("mix_pay_status") public String mixPayStatus;
    @SerializedName("self_pay_status") public String selfPayStatus;
    @SerializedName("med_ins_pay_status") public String medInsPayStatus;
    @SerializedName("paid_time") public String paidTime;
    @SerializedName("mix_pay_type") public String mixPayType;
    @SerializedName("order_type") public String orderType;
    @SerializedName("appid") public String appid;
    @SerializedName("openid") public String openid;
    @SerializedName("out_trade_no") public String outTradeNo;
    @SerializedName("total_fee") public Long totalFee;
    @SerializedName("med_ins_gov_fee") public Long medInsGovFee;
    @SerializedName("med_ins_self_fee") public Long medInsSelfFee;
    @SerializedName("med_ins_other_fee") public Long medInsOtherFee;
    @SerializedName("med_ins_cash_fee") public Long medInsCashFee;
    @SerializedName("wechat_pay_cash_fee") public Long wechatPayCashFee;
    @SerializedName("callback_url") public String callbackUrl;
    @SerializedName("prepay_id") public String prepayId;
    @SerializedName("med_ins_fail_reason") public String medInsFailReason;
  }
}
