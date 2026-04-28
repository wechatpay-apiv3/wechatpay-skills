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
import java.util.HashMap;
import java.util.Map;

/**
 * 服务商医保退款通知 —— 服务商代子商户告知微信医保侧已发生退款
 *
 * 与商户版差异：请求体必须额外传入 sub_mchid（医疗机构商户号）
 */
public class NotifyPartnerMedInsRefund {
  private static String HOST = "https://api.mch.weixin.qq.com";
  private static String METHOD = "POST";
  private static String PATH = "/v3/med-ins/refunds/notify";

  public static void main(String[] args) {
    NotifyPartnerMedInsRefund client = new NotifyPartnerMedInsRefund(
      "1900000100",
      "1DDE55AD98Exxxxxxxxxx",
      "/path/to/apiclient_key.pem",
      "PUB_KEY_ID_xxxxxxxxxxxxx",
      "/path/to/wxp_pub.pem"
    );

    NotifyRefundRequest request = new NotifyRefundRequest();
    request.mixTradeNo = "202204022005169952975171534816";
    request.subMchid = "1900000109";
    request.medRefundTotalFee = 45000L;
    request.medRefundGovFee = 45000L;
    request.medRefundSelfFee = 0L;
    request.medRefundOtherFee = 0L;
    request.refundTime = "2015-05-20T13:29:35+08:00";
    request.outRefundNo = "R202204022005169952975171534816";
    try {
      client.run(request);
      System.out.println("医保退款通知成功");
    } catch (WXPayUtility.ApiException e) {
      e.printStackTrace();
    }
  }

  public void run(NotifyRefundRequest request) {
    String uri = PATH;
    Map<String, Object> args = new HashMap<>();
    args.put("mix_trade_no", request.mixTradeNo);
    String queryString = WXPayUtility.urlEncode(args);
    if (!queryString.isEmpty()) {
      uri = uri + "?" + queryString;
    }
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
        return;
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

  public NotifyPartnerMedInsRefund(String mchid, String certificateSerialNo, String privateKeyFilePath, String wechatPayPublicKeyId, String wechatPayPublicKeyFilePath) {
    this.mchid = mchid;
    this.certificateSerialNo = certificateSerialNo;
    this.privateKey = WXPayUtility.loadPrivateKeyFromPath(privateKeyFilePath);
    this.wechatPayPublicKeyId = wechatPayPublicKeyId;
    this.wechatPayPublicKey = WXPayUtility.loadPublicKeyFromPath(wechatPayPublicKeyFilePath);
  }

  public static class NotifyRefundRequest {
    @SerializedName("mix_trade_no")
    @Expose(serialize = false)
    public String mixTradeNo;

    @SerializedName("sub_mchid") public String subMchid;
    @SerializedName("med_refund_total_fee") public Long medRefundTotalFee;
    @SerializedName("med_refund_gov_fee") public Long medRefundGovFee;
    @SerializedName("med_refund_self_fee") public Long medRefundSelfFee;
    @SerializedName("med_refund_other_fee") public Long medRefundOtherFee;
    @SerializedName("refund_time") public String refundTime;
    @SerializedName("out_refund_no") public String outRefundNo;
  }
}
