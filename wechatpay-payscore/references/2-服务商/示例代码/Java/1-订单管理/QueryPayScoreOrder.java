package com.java.demo;

import com.java.utils.WXPayUtility; // 引用微信支付工具库，参考：https://pay.weixin.qq.com/doc/v3/partner/4014985777

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
 * 查询订单
 */
public class GetPartnerServiceOrder {
  private static String HOST = "https://api.mch.weixin.qq.com";
  private static String METHOD = "GET";
  private static String PATH = "/v3/payscore/partner/serviceorder";

  public static void main(String[] args) {
    // TODO: 请准备商户开发必要参数，参考：https://pay.weixin.qq.com/doc/v3/partner/4013080340
    GetPartnerServiceOrder client = new GetPartnerServiceOrder(
      "19xxxxxxxx",                    // 商户号，是由微信支付系统生成并分配给每个商户的唯一标识符，商户号获取方式参考 https://pay.weixin.qq.com/doc/v3/partner/4013080340
      "1DDE55AD98Exxxxxxxxxx",         // 商户API证书序列号，如何获取请参考 https://pay.weixin.qq.com/doc/v3/partner/4013058924
      "/path/to/apiclient_key.pem",     // 商户API证书私钥文件路径，本地文件路径
      "PUB_KEY_ID_xxxxxxxxxxxxx",      // 微信支付公钥ID，如何获取请参考 https://pay.weixin.qq.com/doc/v3/partner/4013038589
      "/path/to/wxp_pub.pem"           // 微信支付公钥文件路径，本地文件路径
    );

    GetPartnerServiceOrderRequest request = new GetPartnerServiceOrderRequest();
    request.serviceId = "2002000000000558128851361561536";
    request.subMchid = "1900000109";
    request.outOrderNo = "1234323JKHDFE1243252";
    request.queryId = "15646546545165651651";
    try {
      ServiceOrderEntity response = client.run(request);
        // TODO: 请求成功，继续业务逻辑
        System.out.println(response);
    } catch (WXPayUtility.ApiException e) {
        // TODO: 请求失败，根据状态码执行不同的逻辑
        e.printStackTrace();
    }
  }

  public ServiceOrderEntity run(GetPartnerServiceOrderRequest request) {
    String uri = PATH;
    Map<String, Object> args = new HashMap<>();
    args.put("service_id", request.serviceId);
    args.put("sub_mchid", request.subMchid);
    args.put("out_order_no", request.outOrderNo);
    args.put("query_id", request.queryId);
    String queryString = WXPayUtility.urlEncode(args);
    if (!queryString.isEmpty()) {
        uri = uri + "?" + queryString;
    }

    Request.Builder reqBuilder = new Request.Builder().url(HOST + uri);
    reqBuilder.addHeader("Accept", "application/json");
    reqBuilder.addHeader("Wechatpay-Serial", wechatPayPublicKeyId);
    reqBuilder.addHeader("Authorization", WXPayUtility.buildAuthorization(mchid, certificateSerialNo, privateKey, METHOD, uri, null));
    reqBuilder.method(METHOD, null);
    Request httpRequest = reqBuilder.build();

    // 发送HTTP请求
    OkHttpClient client = new OkHttpClient.Builder().build();
    try (Response httpResponse = client.newCall(httpRequest).execute()) {
      String respBody = WXPayUtility.extractBody(httpResponse);
      if (httpResponse.code() >= 200 && httpResponse.code() < 300) {
        // 2XX 成功，验证应答签名
        WXPayUtility.validateResponse(this.wechatPayPublicKeyId, this.wechatPayPublicKey,
            httpResponse.headers(), respBody);

        // 从HTTP应答报文构建返回数据
        return WXPayUtility.fromJson(respBody, ServiceOrderEntity.class);
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

  public GetPartnerServiceOrder(String mchid, String certificateSerialNo, String privateKeyFilePath, String wechatPayPublicKeyId, String wechatPayPublicKeyFilePath) {
    this.mchid = mchid;
    this.certificateSerialNo = certificateSerialNo;
    this.privateKey = WXPayUtility.loadPrivateKeyFromPath(privateKeyFilePath);
    this.wechatPayPublicKeyId = wechatPayPublicKeyId;
    this.wechatPayPublicKey = WXPayUtility.loadPublicKeyFromPath(wechatPayPublicKeyFilePath);
  }

  public static class GetPartnerServiceOrderRequest {
    @SerializedName("service_id")
    @Expose(serialize = false)
    public String serviceId;

    @SerializedName("sub_mchid")
    @Expose(serialize = false)
    public String subMchid;

    @SerializedName("out_order_no")
    @Expose(serialize = false)
    public String outOrderNo;

    @SerializedName("query_id")
    @Expose(serialize = false)
    public String queryId;
  }

  public static class ServiceOrderEntity {
    @SerializedName("out_order_no")
    public String outOrderNo;

    @SerializedName("service_id")
    public String serviceId;

    @SerializedName("appid")
    public String appid;

    @SerializedName("mchid")
    public String mchid;

    @SerializedName("sub_appid")
    public String subAppid;

    @SerializedName("sub_mchid")
    public String subMchid;

    @SerializedName("service_introduction")
    public String serviceIntroduction;

    @SerializedName("state")
    public String state;

    @SerializedName("state_description")
    public String stateDescription;

    @SerializedName("post_payments")
    public Payment postPayments;

    @SerializedName("post_discounts")
    public List<ServiceOrderCoupon> postDiscounts;

    @SerializedName("risk_fund")
    public RiskFund riskFund;

    @SerializedName("total_amount")
    public Long totalAmount;

    @SerializedName("need_collection")
    public Boolean needCollection;

    @SerializedName("collection")
    public Collection collection;

    @SerializedName("time_range")
    public TimeRange timeRange;

    @SerializedName("location")
    public Location location;

    @SerializedName("attach")
    public String attach;

    @SerializedName("notify_url")
    public String notifyUrl;

    @SerializedName("openid")
    public String openid;

    @SerializedName("sub_openid")
    public String subOpenid;

    @SerializedName("order_id")
    public String orderId;
  }

  public static class Payment {
    @SerializedName("name")
    public String name;

    @SerializedName("amount")
    public Long amount;

    @SerializedName("description")
    public String description;

    @SerializedName("count")
    public Long count;
  }

  public static class ServiceOrderCoupon {
    @SerializedName("name")
    public String name;

    @SerializedName("description")
    public String description;

    @SerializedName("amount")
    public Long amount;

    @SerializedName("count")
    public Long count;
  }

  public static class RiskFund {
    @SerializedName("name")
    public String name;

    @SerializedName("amount")
    public Long amount;

    @SerializedName("description")
    public String description;
  }

  public static class Collection {
    @SerializedName("state")
    public String state;

    @SerializedName("total_amount")
    public Long totalAmount;

    @SerializedName("paying_amount")
    public Long payingAmount;

    @SerializedName("paid_amount")
    public Long paidAmount;

    @SerializedName("details")
    public List<Detail> details;
  }

  public static class TimeRange {
    @SerializedName("start_time")
    public String startTime;

    @SerializedName("end_time")
    public String endTime;

    @SerializedName("start_time_remark")
    public String startTimeRemark;

    @SerializedName("end_time_remark")
    public String endTimeRemark;
  }

  public static class Location {
    @SerializedName("start_location")
    public String startLocation;

    @SerializedName("end_location")
    public String endLocation;
  }

  public static class Detail {
    @SerializedName("seq")
    public Long seq;

    @SerializedName("amount")
    public Long amount;

    @SerializedName("paid_type")
    public String paidType;

    @SerializedName("paid_time")
    public String paidTime;

    @SerializedName("transaction_id")
    public String transactionId;
  }

}

