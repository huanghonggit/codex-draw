# Codex Draw v1.0.0

- 从 codex-image2 演进为独立 codex-draw skill，保留 MIT 署名。
- 12 个可独立扩展场景，覆盖来源文章 10 个玩法。
- JSON 模型能力配置、优先级与实验格式控制；同协议扩展不用改代码。
- catalog / validate / plan / draw / inspect / version 命令。
- 本地加密配置与显式 import-image2；拒绝覆盖旧配置目标。
- PNG/JPEG/GIF 真实解码校验、实际尺寸和 GIF 帧/时长/循环信息、偏差警告。
- 默认一次请求；mock 测试、Windows/Linux CI、四平台二进制与 SHA256SUMS。
- SVG 仅设计交接；未实现原生 SVG、视频、切片、前端代码或自动循环修复。
