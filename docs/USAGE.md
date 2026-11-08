# 使用与开发指南

[![test-and-release](https://github.com/huanghonggit/codex-draw/actions/workflows/ci.yml/badge.svg)](https://github.com/huanghonggit/codex-draw/actions/workflows/ci.yml)

一个可扩展的 Codex 绘图 skill：**场景包 + 模型能力注册表 + 原生 Go CLI + 输出验收**。

基于 [fengfengzhidao/codex-image2-skill](https://github.com/fengfengzhidao/codex-image2-skill) 扩展，保留本地配置页、系统加密凭据、生成/编辑/批量调用。新增 12 个可独立维护的场景包，覆盖所引文章的 10 个玩法。

**边界**：首版执行 OpenAI-compatible Images 协议。模型别名不代表官方发布或统一能力；GIF 是可选的服务商扩展。SVG 图标输出设计交接说明，尚无原生 SVG 生成接口。没有自动视频生成、前端代码生成或精灵切片功能。

## 安装

从 [Releases](https://github.com/huanghonggit/codex-draw/releases) 下载 `codex-draw-v1.0.0.zip` 及 `SHA256SUMS`。压缩包含 Windows x64/ARM64、macOS Intel/Apple Silicon 四个原生程序，普通使用无需 Python、Node 或 Go。

Windows 示例（路径按下载位置调整；目标目录应尚不存在）：

```powershell
Get-FileHash .\codex-draw-v1.0.0.zip -Algorithm SHA256
# 与 SHA256SUMS 核对后解压到一个空目录
Expand-Archive .\codex-draw-v1.0.0.zip -DestinationPath .\codex-draw-install
Copy-Item -LiteralPath .\codex-draw-install\codex-draw -Destination "$env:USERPROFILE\.codex\skills\codex-draw" -Recurse
$cli = "$env:USERPROFILE\.codex\skills\codex-draw\bin\codex-draw-windows-amd64.exe"
& $cli version
& $cli validate
```

macOS：解压后将其中 `codex-draw` 目录放入 `~/.codex/skills/`；必要时给对应二进制加执行权限。二进制没有 Apple 公证或 Windows 代码签名；只使用自己核对来源及校验和的发行包。源码 checkout 不含预编译程序，维护者构建方法见后文。

在新会话中使用 `$codex-draw`。如果技能列表未刷新，重开会话。旧 `codex-image2` 可以继续保留，显式调用新的名称避免两者混淆。

## 配置与迁移

```powershell
& $cli configure
& $cli status
# 仅在希望沿用原 codex-image2 服务时执行：
& $cli import-image2
```

配置页由用户输入 API 地址和密钥；不要将密钥发到聊天。Windows 使用当前用户 DPAPI，macOS 使用 Keychain。配置保存在操作系统用户配置目录下的 `codex-draw/config.json`。新配置优先于兼容的 `CODEX_API_URL/CODEX_API_KEY` 环境变量。默认 API 地址为 `https://api.openai.com/v1`，第三方服务需自行配置。

迁移由程序完成解密与重新加密，不输出密钥、不删除旧配置、拒绝覆盖已有新配置；迁移本身不验证远端权限。首次配置与迁移二选一即可。模型列表检测不计费生图；页面内的生图测试需用户明确同意，可能计费。

## 从需求到交付

```powershell
& $cli catalog
& $cli plan --scenario ui-mockup --brief "阅读应用首页，展示正在阅读、书架和今日阅读时长"
& $cli draw --scenario ui-mockup --brief-file .\brief.txt --out .\output\draw\reading.png
& $cli draw --scenario localized-edit --image .\original.png --mask .\mask.png --brief "仅将指定区域换成深蓝色，保留其余内容" --out .\output\draw\edited.png
& $cli inspect .\output\draw\edited.png
```

`plan` 始终离线，不读取凭据。它输出最终提示词、模型候选、路径和验收清单；可重定向保存为本地 JSON。缺少参考图、未知模型或不支持的能力在请求前报错。

`draw` 会发送可能计费的请求。自动模型选择遵循注册表优先级并尝试匹配服务商列表，列表失败时可以回退到兼容候选；并不保证远端支持。用户也可明确指定已注册的 `--model`。默认仅 1 次提交，避免超时后重复计费；显式 `--max-attempts` 最多 5 次。

生成结果 JSON 区分 `requested_size/requested_frame_count` 与 `actual`。实际尺寸不同、帧数不同、未无限循环或首尾不同会给出 `warnings`。图像无法完整解码或格式与后缀不一致则拒绝保存。已有文件默认拒绝覆盖。多个输出在写入前全部解码验证，但磁盘写入不是跨文件事务，磁盘故障仍可能产生部分输出。

### 场景清单

| 场景 ID | 应用 | 对应文章玩法 |
| --- | --- | --- |
| character-loop | 角色循环表情 | 1 |
| game-sprites | 游戏动作素材 | 2 |
| character-sheet | 多角度角色设定 | 3 |
| product-seasonal | 商品场景海报 | 4 |
| product-loop | 商品循环广告 | 4 |
| vector-icons | SVG 图标设计与交接 | 5 |
| localized-edit | 圈选与局部修改 | 6 |
| sketch-render | 草图渲染 | 7 |
| ui-mockup | 应用与网页视觉稿 | 8 |
| knowledge-diagram | 知识过程与结构图解 | 9 |
| image-localization | 图片文字翻译 | 9 |
| storyboard | 四格漫画与分镜 | 10 |

商品静态海报与动态图、知识图解与图片翻译分别拆包，便于单独扩展。图解必须核对专业事实；UI 输出是视觉稿；游戏动作素材还需切片、锚点和引擎验证。

### GIF

```powershell
& $cli draw --scenario character-loop --image .\character.png --brief "固定镜头与背景，眨眼并微笑，然后回到初始姿态" --model gpt-image-2.6-sunburst --allow-experimental --frame-count 24 --frame-duration-ms 120 --timeout 720 --out .\output\draw\loop.gif
```

仅在已确认服务支持或明确接受探索性尝试时使用 `--allow-experimental`。请求帧数不等于实际帧数，首尾像素相同也不等于动作自然。尺寸、身份、背景稳定性和无缝循环都需查看实际结果。首版不内置补帧、光流或自动表情合成。

### SVG

`plan --scenario vector-icons --brief "六种天气图标，统一描边与 viewBox"` 输出交接说明。真实 `draw` 会明确报未实现，不调用生图接口。可由具备矢量/代码输出能力的工具继续完成；不得把位图改名为 SVG。

## 扩展

- **新场景**：新增 `registry/scenarios/<id>/{scenario.json,prompt.md,checklist.md}`；无需改主流程。
- **新模型**：兼容现有请求/响应协议时，新增 `registry/models/<id>.json`；能力、优先级及实验格式显式声明。
- **新服务商**：使用相同 Images 协议时通过配置页切换地址；首版一次配置一个服务，不管理多账户池。
- **新协议或格式**：需增加代码适配与解码器，并补契约测试；仅添加模型名不能接入任意接口。
- **外部扩展目录**：设置 `CODEX_DRAW_REGISTRY` 指向完整注册表，升级安装包时无需改动该目录；首版是替换语义，不是多目录合并。

详见 [扩展协议](../codex-draw/references/extensions.md)、[完整方案](PLAN.md)、[来源和场景映射](SOURCES.md)。新增后运行 `validate` 和 `plan`。JSON Schema 随包提供，运行时不依赖外部库。

## 开发、测试和发布

维护者需要 Go 1.22+；打包脚本需要 Python 3。没有第三方 Go 依赖。

```powershell
go test ./codex-draw/src
go vet ./codex-draw/src
go run ./codex-draw/src validate
python scripts/build.py
```

Linux CI 另跑 `go test -race`；Windows CI 运行临时目录下的合成凭据加密集成测试。全部网络测试使用本地 mock 服务，不使用真实账户。打包只包含白名单文件，不包含配置、密钥、个人照片、生成素材或源码测试。

更新 VERSION 与 CHANGELOG，提交后推送对应 `vX.Y.Z` tag。CI 先测试再交叉构建四平台原生程序，生成 ZIP + SHA256SUMS 并发布 GitHub Release。macOS/ARM64 在首版是交叉编译验证，真实系统凭据交互仍需当地用户验收。

## 许可证

MIT，保留上游版权声明。参见 [LICENSE](../LICENSE) 和 [NOTICE.md](../NOTICE.md)。文章只用于场景分类与工作流启发，本仓库未打包文章图片或整篇正文。
