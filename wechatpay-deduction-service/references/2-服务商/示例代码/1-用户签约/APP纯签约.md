# APP 纯签约（服务商）

> **来源**：[服务商-APP纯签约](https://pay.weixin.qq.com/doc/v2/partner/4011988366.md)
> **协议版本**：API V2
> **签名方式**：HMAC-SHA256 / MD5（默认 MD5）
> **是否需要证书**：否

外部 APP 拉起微信客户端发起签约前，需先后台调用预签约接口完成预签约，获取 `pre_entrustweb_id`，再拉起微信客户端，完成签约，返回 APP。

> 该业务接口需要单独申请，详情查看：[委托代扣-App&H5纯签约申请流程](https://doc.weixin.qq.com/doc/w3_AJ8AyQbhAD4CN41Af5nZSQvqQrvcM)

---

## 步骤 1：预签约接口

| 项 | 值 |
|---|---|
| 适用对象 | 服务商 |
| 请求 URL | `https://api.mch.weixin.qq.com/papay/partner/preentrustweb` |
| 请求方式 | POST |
| 数据格式 | XML |
| 签名方式 | HMAC-SHA256 / MD5（默认 MD5） |

### 请求示例（XML，来源：官方文档原文）

```xml
<xml>
 <appid>wxcbda96de0b165486</appid>
 <mch_id>1200009811</mch_id>
 <sub_appid>wxcbda96de0b165489</sub_appid>
 <sub_mch_id>1900000109</sub_mch_id>
 <plan_id>12535</plan_id>
 <contract_code>100000</contract_code>
 <request_serial>1000</request_serial>
 <contract_display_account>微信代扣</contract_display_account>
 <notify_url>https://weixin.qq.com</notify_url>
 <version>1.0</version>
 <sign_type>MD5</sign_type>
 <sign>C380BEC2BFD727A4B6845133519F3AD6</sign>
 <timestamp>1414488825</timestamp>
 <return_app>Y</return_app>
</xml>
```

### 响应示例（XML，来源：官方文档原文）

```xml
<xml>
 <return_code><![CDATA[SUCCESS]]></return_code>
 <return_msg><![CDATA[OK]]></return_msg>
 <result_code><![CDATA[SUCCESS]]></result_code>
 <appid><![CDATA[wxcbda96de0b165486]]></appid>
 <mch_id><![CDATA[10000098]]></mch_id>
 <sub_appid><![CDATA[wxcbda96de0b165489]]></sub_appid>
 <sub_mch_id><![CDATA[1900000109]]></sub_mch_id>
 <miniprogram_username><![CDATA[gh_xxxxxxxxxxxxx]]></miniprogram_username>
 <miniprogram_path><![CDATA[pages/index/index?sign_scene=app&domain_type=cn&pre_entrustweb_id=xxxxxxxxxx]]></miniprogram_path>
 <nonce_str><![CDATA[IITRi8Iabbblz1Jc]]></nonce_str>
 <sign><![CDATA[E1EE61A91C8E90F299DE6AE075D60A2D]]></sign>
 <pre_entrustweb_id><![CDATA[5778aadY9nltAsZzXixCkFIGYnV2V]]></pre_entrustweb_id>
</xml>
```

> `miniprogram_username` / `miniprogram_path` 仅以下情况返回：
> 1. 模板 ID 为 2025-09-23 及之后申请的模板
> 2. 模板 ID 为 2025-09-23 之前申请且申请了 WXLaunchMiniProgram 权限
>
> 若返回，可使用 LaunchMiniProgram SDK 签约（推荐，详见 [APP调起签约](./APP调起签约.md)）；若未返回则只能用 OpenBusinessWebview。

---

## 步骤 2：签约接口（客户端 OpenSDK 调用）

> ‼️ OpenBusinessWebview 仅支持 2025-09-22 之前申请过权限的商户在存量模板使用。**新增模板请用 WXLaunchMiniProgram**。

预签约 ID（`pre_entrustweb_id`）从步骤 1 获取。

### iOS（来源：官方文档原文）

```objectivec
WXOpenBusinessWebViewReq *req = [[WXOpenBusinessWebViewReq alloc] init];
req.businessType =12; //固定值
NSMutableDictionary *queryInfoDic = [NSMutableDictionary dictionary];
[queryInfoDic setObject:"5778aadY9nltAsZzXixCkFIGYnV2V" forKey:"pre_entrustweb_id"];
req.queryInfoDic = queryInfoDic;
[WxApi sendReq:req];
```

### Android（来源：官方文档原文）

```java
WXOpenBusinessWebview.Req req = new WXOpenBusinessWebview.Req();
req.businessType = 12;//固定值
HashMap  queryInfo = new HashMap<>();
queryInfo.put("pre_entrustweb_id","5778aadY9nltAsZzXixCkFIGYnV2V");
req.queryInfo = queryInfo;
api.sendReq(req);
```

### 返回参数

`WXOpenBusinessWebview.Resp`，**返回参数内容无需关注**。如果签约成功，商户系统会收到 [签约/解约结果通知](../7-异步结果回调/签约-解约结果通知.md)。

## 完整字段说明

请直接参考官方文档：https://pay.weixin.qq.com/doc/v2/partner/4011988366.md
