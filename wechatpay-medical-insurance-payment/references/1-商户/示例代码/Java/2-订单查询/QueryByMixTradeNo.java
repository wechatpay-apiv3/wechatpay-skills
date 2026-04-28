package com.java.demo;

import com.java.utils.WXPayUtility; // 引用微信支付工具库，参考：https://pay.weixin.qq.com/doc/v3/merchant/4014931831

import com.google.gson.annotations.SerializedName;
import com.google.gson.annotations.Expose;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.util.List;

/**
 * 使用医保自费混合订单号查看下单结果
 */
public class QueryByMixTradeNo {
  private static String HOST = "https://api.mch.weixin.qq.com";
  private static String METHOD = "GET";
  private static String PATH = "/v3/med-ins/orders/mix-trade-no/{mix_trade_no}";

  public static void main(String[] args) {
    QueryByMixTradeNo client = new QueryByMixTradeNo(
      "19xxxxxxxx",
      "1DDE55AD98Exxxxxxxxxx",
      "/path/to/apiclient_key.pem",
      "PUB_KEY_ID_xxxxxxxxxxxxx",
      "/path/to/wxp_pub.pem"
    );

    QueryRequest request = new QueryRequest();
    request.mixTradeNo = "202204022005169952975171534816";
    try {
      OrderEntity response = client.run(request);
      System.out.println(response);
    } catch (WXPayUtility.ApiException e) {
      e.printStackTrace();
    }
  }

  public OrderEntity run(QueryRequest request) {
    String uri = PATH.replace("{mix_trade_no}", WXPayUtility.urlEncode(request.mixTradeNo));

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

  public QueryByMixTradeNo(String mchid, String certificateSerialNo, String privateKeyFilePath, String wechatPayPublicKeyId, String wechatPayPublicKeyFilePath) {
    this.mchid = mchid;
    this.certificateSerialNo = certificateSerialNo;
    this.privateKey = WXPayUtility.loadPrivateKeyFromPath(privateKeyFilePath);
    this.wechatPayPublicKeyId = wechatPayPublicKeyId;
    this.wechatPayPublicKey = WXPayUtility.loadPublicKeyFromPath(wechatPayPublicKeyFilePath);
  }

  public static class QueryRequest {
    @SerializedName("mix_trade_no")
    @Expose(serialize = false)
    public String mixTradeNo;
  }

  /** 应答结构与下单接口一致，新增 med_ins_fail_reason（失败原因，仅查询时返回） */
  public static class OrderEntity {
    @SerializedName("mix_trade_no") public String mixTradeNo;
    @SerializedName("mix_pay_status") public String mixPayStatus;
    @SerializedName("self_pay_status") public String selfPayStatus;
    @SerializedName("med_ins_pay_status") public String medInsPayStatus;
    @SerializedName("paid_time") public String paidTime;
    @SerializedName("passthrough_response_content") public String passthroughResponseContent;
    @SerializedName("mix_pay_type") public String mixPayType;
    @SerializedName("order_type") public String orderType;
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
    @SerializedName("callback_url") public String callbackUrl;
    @SerializedName("prepay_id") public String prepayId;
    @SerializedName("attach") public String attach;
    @SerializedName("channel_no") public String channelNo;
    @SerializedName("med_ins_test_env") public Boolean medInsTestEnv;
    @SerializedName("med_ins_fail_reason") public String medInsFailReason;
  }
}
