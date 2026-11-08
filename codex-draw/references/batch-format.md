# 通用批量任务

每行一个 JSON 对象，必填 prompt，可选 model、size、quality、n、out。空行忽略。示例：

```jsonl
{"prompt":"A blue ceramic mug on white","out":"mug.png"}
{"prompt":"A red paper kite in a clear sky","size":"1536x1024","quality":"low","n":2,"out":"kite.png"}
```

使用 generate-batch --input jobs.jsonl --out-dir output/draw --concurrency 2。不同主题分别写行；同主题变体用 n。仅文件名形式的 out 会放到 out-dir 下；含目录的相对路径以当前工作目录为准。各任务须使用唯一名称。n>1 时在文件名追加 -1、-2 等序号。

全局 --output-format、--frame-count、--frame-duration-ms、--allow-experimental 可设置 GIF 批任务条件；不能在 JSONL 行内添加未支持的字段。此命令用于通用生成，不是场景编辑批处理。每行设置不同模型时都必须已注册且支持该请求。

输出 JSON 包含每项成功/失败、实际元数据与偏差警告；任一任务失败则退出非零。--fail-fast 停止继续排队，不取消已经提交的请求。重试默认 1 次，可能计费的失败应由用户决定是否重跑。
