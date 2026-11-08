# Codex Draw

> 用自然语言完成图片生成、图片编辑与条件支持的 GIF 动图制作，让场景和模型都能独立扩展。

[![Tests](https://github.com/huanghonggit/codex-draw/actions/workflows/ci.yml/badge.svg)](https://github.com/huanghonggit/codex-draw/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/huanghonggit/codex-draw)](https://github.com/huanghonggit/codex-draw/releases/latest) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Codex Draw 是面向 Codex 的绘图 skill，适用于商品营销、角色创作、界面设计和内容配图。内置 **12 个场景**，把需求整理、提示词组装、模型能力匹配、图片请求和输出检查串成工作流。使用发行包无需安装 Python、Node 或 Go；图片生成需要自行配置兼容的图片服务。

[下载安装](#下载安装) · [应用场景](#应用场景) · [生成效果](#生成效果) · [快速开始](#快速开始) · [扩展与开发](#扩展与开发)

## 功能特性

- **自然语言绘图**：描述主体、风格、文字和用途，由 Codex 选择场景并组织生成任务。
- **生成与编辑**：支持文生图、参考图编辑、多图输入、可选蒙版和批量生成；具体效果取决于服务与模型。
- **独立场景包**：每个场景包含输入定义、提示词模板和验收清单，可以单独新增、修改和复用。
- **可扩展模型配置**：同协议模型通过 JSON 注册能力和优先级，支持自动选择或明确指定模型。
- **输出检查**：验证文件能否解码，报告真实格式、尺寸、透明度，以及 GIF 帧数、时长、循环标记和首尾差异。
- **本地配置**：浏览器配置页填写服务地址与密钥，Windows 使用 DPAPI、macOS 使用 Keychain 保存凭据。
- **跨平台发行包**：提供 Windows、macOS 的 x64 与 ARM64 原生程序；默认拒绝覆盖已有文件，每次请求默认只提交一次。

## 下载安装

从 [GitHub Releases](https://github.com/huanghonggit/codex-draw/releases/latest) 下载 ZIP 和 `SHA256SUMS`，核对校验和后解压。把包内的 `codex-draw` 文件夹放入 `~/.codex/skills/`，然后新开 Codex 会话。

| 平台 | 安装目录 | 包内程序 |
| --- | --- | --- |
| Windows x64 / ARM64 | `%USERPROFILE%\.codex\skills\codex-draw` | `codex-draw-windows-amd64.exe` / `codex-draw-windows-arm64.exe` |
| macOS Intel / Apple Silicon | `~/.codex/skills/codex-draw` | `codex-draw-darwin-amd64` / `codex-draw-darwin-arm64` |

程序位于 skill 的 `bin/` 目录。macOS 若未保留执行权限，需为对应程序添加执行权限。macOS 和 ARM64 包目前经过交叉编译验证，实际系统交互仍需真机确认；安装命令见 [使用指南](docs/USAGE.md#安装)。

## 应用场景

下表是内置工作流及示例需求，不代表每个场景都已完成真实服务测试。已发布的实测文件见下方“生成效果”。除场景工作流外，也可直接使用通用文生图、图片编辑和批量生成。

### 角色与故事创作

| 场景 | 需要提供 | 用途与交付 | 示例需求 |
| --- | --- | --- | --- |
| [角色循环表情](codex-draw/registry/scenarios/character-loop) `character-loop` | 角色参考图、动作顺序、固定特征、时长 | 请求角色表情 GIF；需服务商支持，并检查身份与循环接缝 | 保持发型、服装和背景，眨眼后微笑，再回到初始姿态 |
| [游戏动作素材](codex-draw/registry/scenarios/game-sprites) `game-sprites` | 角色参考图、动作朝向、帧格尺寸、调色板 | 生成动作画板或精灵图视觉素材；切片、锚点和引擎集成需另行处理 | 为像素角色设计向右行走动作，统一格子尺寸和配色 |
| [多角度角色设定](codex-draw/registry/scenarios/character-sheet) `character-sheet` | 角色参考图、视角、不可变特征 | 生成正面、侧面、背面等设定参考，核对服装与身体比例 | 为这位绿围巾小狐狸制作三视图，保留毛色和配饰 |
| [故事分镜](codex-draw/registry/scenarios/storyboard) `storyboard` | 角色设定、情节顺序、画风、对白 | 生成四格漫画或分镜图，检查情节和角色连贯性 | 画四格小狐狸泡茶：取茶、投茶、注水、捧杯微笑 |

### 商品与图片编辑

| 场景 | 需要提供 | 用途与交付 | 示例需求 |
| --- | --- | --- | --- |
| [商品场景海报](codex-draw/registry/scenarios/product-seasonal) `product-seasonal` | 商品原图、主题、标签原文、比例 | 编辑季节或营销背景，复核商品形状、颜色和标签 | 把这只绿色陶瓷杯放到冬季窗边，加入松枝和柔和光线 |
| [商品循环广告](codex-draw/registry/scenarios/product-loop) `product-loop` | 商品图、单一动作、锁定细节、时长 | 请求商品 GIF；需服务商支持，不保证无缝循环 | 固定杯子与镜头，只让热气和窗外雪花轻微移动 |
| [局部编辑](codex-draw/registry/scenarios/localized-edit) `localized-edit` | 原图、标记图或 PNG 蒙版、允许改动区域 | 定向修改颜色、材质或局部内容；圈外像素保持需验收 | 仅将靠垫改成深蓝色，保留沙发和房间布置 |
| [草图渲染](codex-draw/registry/scenarios/sketch-render) `sketch-render` | 草图、布局、材质、光照与风格 | 把草图转为效果图，检查空间和结构是否保留 | 按这张客厅草图生成原木风效果图，保留门窗位置 |
| [图片文字翻译](codex-draw/registry/scenarios/image-localization) `image-localization` | 原图、目标语言、术语表、保留内容 | 翻译图中文字并重绘排版，逐项核对文字与布局 | 将这张英文信息图改为简体中文，保留图标和数值 |

### 界面与信息设计

| 场景 | 需要提供 | 用途与交付 | 示例需求 |
| --- | --- | --- | --- |
| [UI 视觉稿](codex-draw/registry/scenarios/ui-mockup) `ui-mockup` | 页面用途、模块、屏幕比例、精确文案 | 生成应用或网页视觉稿；不包含可运行前端代码 | 设计中文阅读应用首页，包含阅读进度、继续阅读和底部导航 |
| [知识图解](codex-draw/registry/scenarios/knowledge-diagram) `knowledge-diagram` | 已核实事实、步骤、结构关系、术语 | 生成流程图解或结构配图，需核对事实、箭头和文字 | 根据提供的资料绘制水循环图，清楚标注各阶段 |
| [SVG 图标设计交接](codex-draw/registry/scenarios/vector-icons) `vector-icons` | 图标清单、viewBox、描边、圆角、颜色 | 仅输出设计交接说明，供矢量或代码工具继续完成；尚未直接生成 SVG | 设计六种天气图标，统一 24×24 网格和 2px 描边 |

## 生成效果

以下为真实服务返回的原始文件。全部记录的实际模型为服务商别名 `gpt-image-2.6-sunburst`；本次由 `codex-draw` 规划和检查，已配置的 `codex-image2` 执行请求。完整提示词、参数、校验和、验收意见及复现命令见 [案例测试](docs/EXAMPLES.md)。

| 商品参考图 | 冬季场景编辑 |
| --- | --- |
| ![生成的陶瓷杯参考图](examples/results/product-reference.png) | ![冬季商品海报](examples/results/winter-product.png) |

商品案例展示“先生成参考图，再编辑背景”的流程。颜色与杯柄朝向基本保留，杯身比例、杯缘和杯柄厚度略有变化。

| 中文 UI 视觉稿 | 四格故事分镜 |
| --- | --- |
| ![阅读应用首页](examples/results/reading-ui.png) | ![小狐狸泡茶四格漫画](examples/results/fox-storyboard.png) |

UI 案例的 8 处指定中文通过图像理解检查；分镜案例的情节顺序和角色基本连贯，勺子角度略不自然。四张静态图均为 1024×1024。

**商品 GIF**

![商品动图](examples/results/winter-product-loop.gif)

实测 **9 帧、1.8 秒、439×388、无限循环**。请求尺寸为 1024×1024，且首尾蒸汽存在跳变，因此尺寸和无缝要求未达标；此处保留真实结果，不作为无缝动画通过案例。检查采用解码与抽帧复核，动态播放仍需人工验收。

## 快速开始

### 1. 配置图片服务

安装后在 Codex 中输入：

```text
使用 $codex-draw，帮我配置图片服务。
```

在打开的本地配置页自行填写 API 地址和密钥，不必把密钥发到聊天中。服务需兼容当前 Images 协议；具体模型、编辑能力、价格和 GIF 支持以该服务为准。新旧 skill 配置独立；如需迁移已有 `codex-image2` 配置，请明确提出迁移要求。

### 2. 描述需求并附上参考图

```text
使用 $codex-draw，设计一张中文阅读应用首页。米白与森林绿配色，
包含“今日阅读”、阅读进度、“继续阅读”按钮，以及“首页、书架、我的”导航。

使用 $codex-draw，把上传的商品图改成冬季窗边海报，
保留杯子颜色和杯柄朝向，不新增文字，输出到 output/draw/winter.png。

使用 $codex-draw，把这张海报生成 GIF。固定商品与镜头，仅让蒸汽缓慢飘动。
使用 gpt-image-2.6-sunburst，允许尝试服务商 GIF 扩展，完成后报告真实帧数和循环接缝。
```

编辑类场景需要原图。明确用途、精确文字、必须保留的内容和允许变化的范围，有助于后续验收。

### 3. 检查结果

默认输出目录为当前工作目录的 `output/draw/`，已有文件默认拒绝覆盖。CLI 检查格式与元数据；文字、商品或角色一致性、专业事实和动画观感仍需视觉复核。

<details>
<summary>命令行示例：先离线规划，再生成与检查（PowerShell）</summary>

```powershell
$cli = Join-Path $env:USERPROFILE ".codex/skills/codex-draw/bin/codex-draw-windows-amd64.exe"
& $cli validate
& $cli catalog
& $cli plan --scenario ui-mockup --brief "中文阅读应用首页，米白与森林绿配色"
# draw 会发送可能计费的请求；plan 不发送请求
& $cli draw --scenario ui-mockup --brief "中文阅读应用首页，米白与森林绿配色" --out "./output/draw/reading.png"
& $cli inspect "./output/draw/reading.png"
```

更多编辑、GIF、配置迁移和批量用法见 [使用指南](docs/USAGE.md) 与 [批量格式](codex-draw/references/batch-format.md)。

</details>

## 模型与能力边界

以下是仓库内置的**服务商模型别名及注册能力**，不代表官方模型发布、远端可用性或质量保证。

| 模型配置 | 生成 / 编辑 | 注册格式 |
| --- | --- | --- |
| `gpt-image-2` | 支持 | PNG、JPEG |
| `gpt-image-2.5` | 支持 | PNG、JPEG |
| `gpt-image-2.6-sunburst` | 支持 | PNG、JPEG；GIF 为实验扩展 |
| `gpt-image-2.6-flare` | 支持 | PNG、JPEG；GIF 为实验扩展 |

`auto` 按任务能力与注册优先级选择候选，并尝试匹配服务商的模型列表。GIF 必须明确允许实验能力并由服务商实际支持；请求帧数、尺寸和无缝要求不等于实际结果。当前没有原生视频生成、自动补帧、精灵切片或可运行 UI 代码输出；SVG 场景仅用于设计交接。

## 扩展与开发

场景定义“做什么”，模型配置定义“可请求哪些能力”，两者分别维护。

| 扩展目标 | 操作 | 是否需要改代码 |
| --- | --- | --- |
| 新增应用场景 | 在 `registry/scenarios/<id>/` 添加 `scenario.json`、`prompt.md`、`checklist.md` | 不需要；新增后运行 `validate` 和 `plan` |
| 新增同协议模型 | 在 `registry/models/` 添加 JSON，声明操作、格式、优先级和实验能力 | 不需要；仍需验证服务端兼容性 |
| 接入新协议或格式 | 增加传输适配、响应处理或解码器，并补契约测试 | 需要 |
| 保留自定义注册表 | 用 `CODEX_DRAW_REGISTRY` 指向外部完整注册表 | 不需要；当前为替换语义，不与内置目录合并 |

### 项目结构

```text
codex-draw/
├── codex-draw/
│   ├── SKILL.md              # Codex 工作流入口
│   ├── registry/scenarios/   # 独立场景包
│   ├── registry/models/      # 模型能力配置
│   ├── registry/schemas/     # 注册表结构定义
│   ├── references/           # 扩展协议与调用说明
│   └── src/                  # Go CLI 与测试
├── examples/                 # 案例定义、简报、原始产物与验收记录
├── scripts/                  # 构建、案例执行与离线校验
└── docs/                     # 使用指南、方案与来源说明
```

维护者需要 Go 1.22+；构建与案例脚本使用 Python 3。Windows/Linux CI 运行单元与 mock 集成测试、注册表和案例完整性检查，Linux 另运行竞争检测；CI 不调用付费图片服务。

- [使用与开发指南](docs/USAGE.md)：配置、命令、测试和发行流程。
- [案例测试](docs/EXAMPLES.md)：真实产物、验收范围与复现方式。
- [扩展协议](codex-draw/references/extensions.md)：场景、模型与新协议的接入规范。
- [完整方案](docs/PLAN.md)：分层设计、工程要求与未实现的路线图。

## 贡献与许可

欢迎提交场景包、模型配置、问题复现和案例验收改进。场景贡献应同时包含模板与检查项，模型贡献应说明服务与协议依据；提交前运行注册表验证和相关测试，发布案例时保留真实参数与未达标项。

基于 [codex-image2-skill](https://github.com/fengfengzhidao/codex-image2-skill) 扩展，采用 [MIT 许可](LICENSE)。场景分类与文章映射见 [来源说明](docs/SOURCES.md)，上游署名见 [NOTICE](NOTICE.md)。
