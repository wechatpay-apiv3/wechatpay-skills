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
 * 创建
 */
public class CreateServiceOrder {
  private static String HOST = "https://api.mch.weixin.qq.com";
  private static String METHOD = "POST";
  private static String PATH = "/v3/payscore/serviceorder";

  public static void main(String[] args) {
    // TODO: 请准备商户开发必要参数，参考：https://pay.weixin.qq.com/doc/v3/merchant/4013070756
    CreateServiceOrder client = new CreateServiceOrder(
      "19xxxxxxxx",                    // 商户号，是由微信支付系统生成并分配给每个商户的唯一标识符，商户号获取方式参考 https://pay.weixin.qq.com/doc/v3/merchant/4013070756
      "1DDE55AD98Exxxxxxxxxx",         // 商户API证书序列号，如何获取请参考 https://pay.weixin.qq.com/doc/v3/merchant/4013053053
      "/path/to/apiclient_key.pem",     // 商户API证书私钥文件路径，本地文件路径
      "PUB_KEY_ID_xxxxxxxxxxxxx",      // 微信支付公钥ID，如何获取请参考 https://pay.weixin.qq.com/doc/v3/merchant/4013038816
      "/path/to/wxp_pub.pem"           // 微信支付公钥文件路径，本地文件路径
    );

    CreateServiceOrderRequest request = new CreateServiceOrderRequest();
    request.outOrderNo = "1234323JKHDFE1243252";
    request.appid = "wxd678efh567hg6787";
    request.serviceId = "2002000000000558128851361561536";
    request.serviceIntroduction = "某某酒店";
    request.postPayments = new ArrayList<>();
    {
      Payment postPaymentsItem = new Payment();
      postPaymentsItem.name = "就餐费用";
      postPaymentsItem.amount = 40000L;
      postPaymentsItem.description = "就餐人均100元";
      postPaymentsItem.count = 4L;
      request.postPayments.add(postPaymentsItem);
    };
    request.postDiscounts = new ArrayList<>();
    {
      ServiceOrderCoupon postDiscountsItem = new ServiceOrderCoupon();
      postDiscountsItem.name = "满20减1元";
      postDiscountsItem.description = "不与其他优惠叠加";
      postDiscountsItem.amount = 100L;
      postDiscountsItem.count = 2L;
      request.postDiscounts.add(postDiscountsItem);
    };
    request.timeRange = new TimeRange();
    request.timeRange.startTime = "20091225091010";
    request.timeRange.endTime = "20091225121010";
    request.timeRange.startTimeRemark = "备注1";
    request.timeRange.endTimeRemark = "备注2";
    request.location = new Location();
    request.location.startLocation = "嗨客时尚主题展餐厅";
    request.location.endLocation = "嗨客时尚主题展餐厅";
    request.riskFund = new RiskFund();
    request.riskFund.name = "DEPOSIT";
    request.riskFund.amount = 10000L;
    request.riskFund.description = "就餐的预估费用";
    request.attach = "Easdfowealsdkjfnlaksjdlfkwqoi&wl3l2sald";
    request.notifyUrl = "https://api.test.com";
    request.needUserConfirm = false;
    request.device = new Device();
    request.device.startDeviceId = "HG123456";
    request.device.endDeviceId = "HG123456";
    request.device.materielNo = "example_materiel_no";
    try {
      CreateServiceOrderResponse response = client.run(request);
        // TODO: 请求成功，继续业务逻辑
        System.out.println(response);
    } catch (WXPayUtility.ApiException e) {
        // TODO: 请求失败，根据状态码执行不同的逻辑
        e.printStackTrace();
    }
  }

  public CreateServiceOrderResponse run(CreateServiceOrderRequest request) {
    String uri = PATH;
    String reqBody = WXPayUtility.toJson(request);

    Request.Builder reqBuilder = new Request.Builder().url(HOST + uri);
    reqBuilder.addHeader("Accept", "application/json");
    reqBuilder.addHeader("Wechatpay-Serial", wechatPayPublicKeyId);
    reqBuilder.addHeader("Authorization", WXPayUtility.buildAuthorization(mchid, certificateSerialNo,privateKey, METHOD, uri, reqBody));
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

        return WXPayUtility.fromJson(respBody, CreateServiceOrderResponse.class);
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

  public CreateServiceOrder(String mchid, String certificateSerialNo, String privateKeyFilePath, String wechatPayPublicKeyId, String wechatPayPublicKeyFilePath) {
    this.mchid = mchid;
    this.certificateSerialNo = certificateSerialNo;
    this.privateKey = WXPayUtility.loadPrivateKeyFromPath(privateKeyFilePath);
    this.wechatPayPublicKeyId = wechatPayPublicKeyId;
    this.wechatPayPublicKey = WXPayUtility.loadPublicKeyFromPath(wechatPayPublicKeyFilePath);
  }

  public static class CreateServiceOrderRequest {
    @SerializedName("out_order_no")
    public String outOrderNo;

    @SerializedName("appid")
    public String appid;

    @SerializedName("service_id")
    public String serviceId;

    @SerializedName("service_introduction")
    public String serviceIntroduction;

    @SerializedName("post_payments")
    public List<Payment> postPayments;

    @SerializedName("post_discounts")
    public List<ServiceOrderCoupon> postDiscounts;

    @SerializedName("time_range")
    public TimeRange timeRange;

    @SerializedName("location")
    public Location location;

    @SerializedName("risk_fund")
    public RiskFund riskFund;

    @SerializedName("attach")
    public String attach;

    @SerializedName("notify_url")
    public String notifyUrl;

    @SerializedName("need_user_confirm")
    public Boolean needUserConfirm;

    @SerializedName("device")
    public Device device;
  }

  public static class CreateServiceOrderResponse {
    @SerializedName("appid")
    public String appid;

    @SerializedName("mchid")
    public String mchid;

    @SerializedName("out_order_no")
    public String outOrderNo;

    @SerializedName("service_id")
    public String serviceId;

    @SerializedName("service_introduction")
    public String serviceIntroduction;

    @SerializedName("state")
    public String state;

    @SerializedName("state_description")
    public String stateDescription;

    @SerializedName("post_payments")
    public List<Payment> postPayments;

    @SerializedName("post_discounts")
    public List<ServiceOrderCoupon> postDiscounts;

    @SerializedName("risk_fund")
    public RiskFund riskFund;

    @SerializedName("time_range")
    public TimeRange timeRange;

    @SerializedName("location")
    public Location location;

    @SerializedName("attach")
    public String attach;

    @SerializedName("notify_url")
    public String notifyUrl;

    @SerializedName("order_id")
    public String orderId;

    @SerializedName("package")
    public String _package;
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

  public static class RiskFund {
    @SerializedName("name")
    public String name;

    @SerializedName("amount")
    public Long amount;

    @SerializedName("description")
    public String description;
  }

  public static class Device {
    @SerializedName("start_device_id")
    public String startDeviceId;

    @SerializedName("end_device_id")
    public String endDeviceId;

    @SerializedName("materiel_no")
    public String materielNo;
  }

}
