# APP 调起签约（直连商户）

> **来源**：[APP调起签约 - WXLaunchMiniProgram](https://pay.weixin.qq.com/doc/v2/merchant/4015996790.md)
> **接口名称**：`WXLaunchMiniProgram`
> **接口类型**：客户端 OpenSDK 调用（无服务端 HTTP 接口）

商户在完成"委托代扣 APP 预签约"后，可在移动端 APP 集成 OpenSDK 调起微信，请求用户签约。

## 接入前注意

1. 该能力依赖微信 [Open SDK](https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Pay/Vendor_Service_Center.html)，需在微信开放平台申请开通移动应用的微信支付能力
2. 商户在使用 WXLaunchMiniProgram 调起 APP 签约前，需要邮件申请，参考[委托代扣-APP纯签约申请流程](https://doc.weixin.qq.com/doc/w3_AdgAIgbzAB8CNJOvTlNjRTpGQ88aj)
3. **预签约成功后首次调起微信有严格时间限制**：需在 2 分钟内调起微信（该时间将来会调整，最短可能缩短至 20 秒）

## 平台要求

| 平台 | 最低 SDK 版本 | 资源下载 |
|---|---|---|
| Android | >=5.3.1 | [Android资源下载](https://developers.weixin.qq.com/doc/oplatform/Downloads/Android_Resource.html) |
| iOS | >=1.8.4 | [iOS资源下载](https://developers.weixin.qq.com/doc/oplatform/Downloads/iOS_Resource.html) |
| 鸿蒙 | - | [鸿蒙资源下载](https://developers.weixin.qq.com/doc/oplatform/Downloads/HarmonyOS_Resource.html) |

## 请求参数

| 名称 | 变量 | 必填 | 类型 | 示例 | 描述 |
|---|---|---|---|---|---|
| 跳转签约小程序的 username | `userName` | 是 | String(32) | `gh_xxxxxxxxxxxxx` | 从预签约接口的返回参数 `miniprogram_username` 获取 |
| 跳转签约小程序的 path | `path` | 是 | String(256) | `pages/index/index?xxxxxx` | 从预签约接口的返回参数 `miniprogram_path` 获取 |
| 跳转的小程序版本 | `miniprogramType` | 是 | Integer | `0` | 固定传值为 0（正式版） |

## 客户端调用示例

官方提供以下平台的开发示例（外链）：

- [Android 开发示例](https://developers.weixin.qq.com/doc/oplatform/Mobile_App/Launching_a_Mini_Program/Android_Development_example.html)
- [iOS 开发示例](https://developers.weixin.qq.com/doc/oplatform/Mobile_App/Launching_a_Mini_Program/iOS_Development_example.html)
- [鸿蒙开发示例](https://developers.weixin.qq.com/doc/oplatform/Mobile_App/Launching_a_Mini_Program/OHOS_Development_example.html)

## 返回参数

返回参数内容无需关注。如果签约成功，商户系统会收到 [签约/解约结果通知](../7-异步结果回调/签约-解约结果通知.md)。

---

## OpenBusinessWebview vs WXLaunchMiniProgram

| 对比项 | OpenBusinessWebview | WXLaunchMiniProgram |
|---|---|---|
| 跳转微信耗时 | 中等 | 更短 |
| 用户体验 | 中等 | 更好 |
| 支持 Android | 是 | 是 |
| 支持 iOS | 是 | 是 |
| 支持鸿蒙 | 否 | 是 |
| 新增模板 | 禁用 | 邮件申请后可用 |
| 存量模板（已接入 OpenBusinessWebview） | 邮件申请关闭后禁用 | 邮件申请后可用 |

> 2025-09-23 之后申请的模板**只能用 WXLaunchMiniProgram**。
> 切换流程详见官方文档"存量委托代扣模板升级指引"章节。

OpenBusinessWebview 调用方式参见 [APP纯签约 - 步骤2](./APP纯签约.md#步骤-2签约接口客户端-opensdk-调用)。

## 完整字段说明

请直接参考官方文档：https://pay.weixin.qq.com/doc/v2/merchant/4015996790.md
