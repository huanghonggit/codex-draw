---
name: codex-draw
description: 可扩展的绘图与图片编辑工作流，使用已配置的 OpenAI-compatible 图片服务。用于 codex-draw 配置、文生图、局部编辑、角色设定、游戏动作素材、商品海报、条件支持的 GIF 循环动图、草图渲染、UI 视觉稿、知识图解、图片翻译和故事分镜；也可输出 SVG 图标的设计交接说明。支持新增场景包和模型配置，不把模型名称视为能力证明。
---

# Codex Draw

使用本 skill 的原生 CLI。发行包无需安装 Python、Node 或 Go。选择当前系统对应的可执行文件：

- Windows x64：`bin/codex-draw-windows-amd64.exe`
- Windows ARM64：`bin/codex-draw-windows-arm64.exe`
- macOS Intel：`bin/codex-draw-darwin-amd64`
- macOS Apple Silicon：`bin/codex-draw-darwin-arm64`

下文 `<cli>` 表示所选文件。macOS 如执行权限未保留，执行 `chmod +x <cli>`。日常使用不现场编译。

## 路由与流程

1. 首次使用或扩展后运行 `<cli> validate`；用 `<cli> catalog` 获取当前场景与模型，不假定下面的索引完整。
2. 按需求选择场景，仅阅读对应 `registry/scenarios/<id>/scenario.json`、`prompt.md`、`checklist.md`。理解输入图片，指出每张参考图的角色。用户指定的模型与细节优先；不要擅自替换模型。
3. 整理任务简报：用途、主体、精确文字、构图、不变项、允许改动项。缺少关键事实时询问；普通风格选择可合理补全。专业图解先取得用户提供或核实的事实。
4. 使用 `plan` 离线检查提示词、能力匹配和输出路径。它不读取凭据、不查模型列表、不发起图片请求。模型在注册表中并不意味着远端可用。
5. 生成前运行一次 `status`。未配置时按下文配置；已配置且任务已授权即可 `draw`，无需为了例行生成重复要求确认。
6. 检查结果中的 `actual` 和 `warnings`，再逐项进行视觉验收。修改时只改一个明确问题，保留原图。GIF 必须查看动态播放及循环接缝。
7. 报告实际模型、绝对文件路径、真实尺寸；动图另报实际帧数、时长、循环标记，以及尚未满足的要求。明确区分原生模型产物与后处理结果。

## 场景索引

| ID | 用途 |
| --- | --- |
| character-loop | 角色循环表情 |
| game-sprites | 游戏动作素材 |
| character-sheet | 多角度角色设定 |
| product-seasonal | 商品场景海报 |
| product-loop | 商品循环广告 |
| vector-icons | SVG 图标设计与交接 |
| localized-edit | 圈选与局部修改 |
| sketch-render | 草图渲染 |
| ui-mockup | 应用与网页视觉稿 |
| knowledge-diagram | 知识过程与结构图解 |
| image-localization | 图片文字翻译 |
| storyboard | 四格漫画与分镜 |

通用绘图可直接使用 `generate`、`edit`；不同主题的批任务用 `generate-batch`，详见 [批量格式](references/batch-format.md)。场景包可独立扩展，新增之后以 `catalog` 为准。

## 常见调用

PowerShell 示例，将 `<skill-dir>` 替换为安装路径：

```powershell
$cli = "<skill-dir>\bin\codex-draw-windows-amd64.exe"
& $cli plan --scenario ui-mockup --brief "日历应用首页，白色背景、蓝色强调色，包含今日安排和月份切换"
& $cli draw --scenario ui-mockup --brief-file ".\brief.txt" --out ".\output\draw\calendar.png"
& $cli draw --scenario product-seasonal --image ".\product.png" --brief "冬季窗边布景；商品包装和标签文字完全保留" --out ".\output\draw\winter.png"
& $cli inspect ".\output\draw\winter.png"
```

`--model auto` 根据已注册能力筛选并按优先级尝试匹配服务商模型列表；列表不可用时回退到兼容候选模型，实际调用仍可能失败。精确模型需已注册。一次调用默认只有 1 次提交；重试可能重复计费，不自动扩大重试。

## GIF 与其他特殊交付

当用户明确要求 GIF，先检查所选模型的配置及服务商实际支持情况。内置 2.6 别名的 GIF 字段是服务商扩展，不能据此声称是官方通用接口。

```powershell
& $cli plan --scenario character-loop --image ".\character.png" --brief "轻轻眨眼并微笑，然后回到初始姿态；固定背景" --model gpt-image-2.6-sunburst --allow-experimental
& $cli draw --scenario character-loop --image ".\character.png" --brief-file ".\loop-brief.txt" --model gpt-image-2.6-sunburst --allow-experimental --frame-count 24 --frame-duration-ms 120 --timeout 720 --out ".\output\draw\loop.gif"
```

`--allow-experimental` 仅在服务商支持已经确认、或用户明确授权尝试该扩展时使用；它不验证远端能力。不能悄悄回退成静态图。帧数、透明背景、精确尺寸、身份保持、圈外像素不变和无缝循环都是需要验收的要求。

SVG 场景只提供设计交接：`plan --scenario vector-icons ...` 返回 `handoff_required`；真实 `draw` 返回非零状态，不发送生图请求。不得将 PNG 改后缀当作 SVG。需要原生矢量、视频、精灵切片、前端代码或确定性的蒙版合成时，明确采用额外工具的范围。

## 配置与迁移

- `status` 只证明本地配置可读，不证明远端认证、权限或模型能力。
- `configure` 启动临时本地配置页并立即返回。让用户亲自在浏览器填写 API 地址与密钥；如果浏览器未打开，提供 CLI 返回的本地链接。
- 不读取凭据文件或钥匙串，不让用户把密钥发到聊天，不通过命令或自动化填密钥，不点击“显示”。Windows 用 DPAPI，macOS 用 Keychain。
- 保存前探测模型列表。未发现已注册模型的服务需要用户在页面明确选择仍然保存，不能代点。认证、网络、限流或服务器错误不能绕过。
- 可选的计费生图测试需要用户明确同意；不要用真实账户跑自动测试。
- `import-image2` 仅在用户要求迁移配置时执行，由程序读取旧配置并重新加密；目标已存在则拒绝，原配置保留。不自动迁移。
- 新配置位于操作系统用户配置目录的 `codex-draw` 子目录；优先于兼容的 `CODEX_API_URL/CODEX_API_KEY` 环境变量。已有损坏配置不能悄悄降级到环境变量。未设地址时默认官方 API 地址，具体模型仍由服务提供。
- 输出默认在当前工作目录 `output/draw/`。已有文件默认拒绝覆盖；仅用户授权时用 `--force`。

## 扩展与验收

新增场景、模型、协议适配的边界见 [扩展协议](references/extensions.md)。仅修改注册表时不用重新编译，但必须运行 `validate` 和对应的离线计划。

二进制校验只保证可解码、格式与文件后缀一致，并报告尺寸、透明度或 GIF 元数据。它不验证文字、科学事实、人物身份或主观视觉质量。
