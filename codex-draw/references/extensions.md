# 扩展协议 v1

注册表根目录默认是可执行文件上级的 registry。源码测试另搜索当前目录；可用 CODEX_DRAW_REGISTRY 指定一套完整的自定义注册表（替换，不合并）。安装包内修改在升级时需要自行保留，推荐外部注册表。

## 添加场景：无需改 Go 代码

复制 registry/scenarios 中最接近的目录，改名，并修改 scenario.json 的 id。必备文件：

- scenario.json：schema_version=1；id 与目录名相同；title、description、operation、default_format、reference_required、article_cases、inputs。
- prompt.md：包含 {{brief}}，它只做字符串替换，不执行模板代码、命令或网页指令。
- checklist.md：独立验收清单。

operation 是 generate、edit 或 handoff。edit 要求输入图片；SVG 必须使用 handoff。新增场景无文章出处时 article_cases=[]。描述文字会进入模型提示词，安装第三方包前审阅内容，不能把模板当成高于用户的指令。

运行 validate，再对新场景执行 plan。新增场景在 catalog 自动可见，不需要改 SKILL 的固定表才能工作。

## 添加模型：同协议只改 JSON

复制 registry/models 中一个 JSON，文件名为 <id>.json。schema_version=1，id 为安全的模型标识符，adapter 当前只能是 openai-images-v1。operations 与 formats 是明确声明的能力；priority 越小优先级越高；experimental_formats 必须是 formats 子集。

当前适配器实现 POST /v1/images/generations 的 JSON 请求，以及 POST /v1/images/edits 的 multipart 请求（多张 image、可选 mask）；解析 data[].b64_json 或 URL。支持 PNG、JPEG 及可选的 GIF 扩展字段 output_format/frame_count/frame_duration_ms。

注册表不是远端能力证明。notes 必须说明依据与适用条件，模型 ID 是当前服务中的别名。GIF 默认放入 experimental_formats。若未来模型用相同协议且参数兼容，只添加 JSON；不要把不兼容的 Responses、异步视频或独立 SVG 接口冒充成此协议。

## 新协议需要代码

在独立源文件实现新传输函数、凭据与响应处理；扩展 adapter 枚举和调度、对应 schema，并增加 mock 契约测试。首版没有运行任意第三方插件代码的加载器。新的静态格式也要增加真实解码器和后缀校验。

## 版本兼容

不认识的字段、重复 ID、未知 schema_version、错误格式或缺失模板均拒绝。发布前同时运行 Go 测试、注册表验证与 skill 验证。schema 变更提升版本，提供迁移说明；禁止静默改变已有字段含义。

JSON Schema 位于 registry/schemas。CLI 的 validate 使用内置语义校验，故发行包不依赖 JSON Schema 运行库。
