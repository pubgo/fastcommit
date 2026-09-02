# [Unreleased]

> 推荐维护方式：`fastgit changelog draft|release`

## 新增

- 新增顶层 git 动词子命令：`fastgit status|status short`、`fastgit add`、`fastgit log|log graph`、`fastgit diff [--staged|--unstaged]`、`fastgit branch current|list|checkout|checkout-remote|create|delete`、`fastgit fetch [--prune]`、`fastgit rebase <upstream>|--continue|--abort|--skip`、`fastgit remote|remote list`
- `fastgit pull` 新增 `--rebase` flag（rebase 方式拉取，与 `--all`/`--hard` 互斥）
- `fastgit tag show <tag>`：查看 tag 详情（`git show <tag>`）
- `fastgit ui`：原 `ggc` 交互命令面（fuzzy 搜索 + workflow + alias），配置文件仍为 `ggc.yaml`，路径不变

## 修复

暂无

## 变更

- 移除 `fastgit ggc` 统一入口，原有命令摊平为顶层子命令；`ggc pull current|pull rebase`、`ggc push current|push force`、`ggc tag list`、`ggc commit <message>` 分别由 `pull`、`push`、`tag`、`git commit -m` / `fastgit commit` 承接

## 文档

- README、docs/features、docs/architecture、docs/roadmap 同步新命令面

## 影响范围

- CLI 命令面：`fastgit ggc` 用户需改用顶层子命令或 `fastgit ui`
- `ggc.yaml`（workflow/alias）路径与格式不变，无需迁移

## 验证建议

- `go build ./... && go test ./...`
- `fastgit status`、`fastgit branch current`、`fastgit remote`、`fastgit pull --help`（确认 `--rebase`）、`fastgit ui list`、`fastgit ui path`

## 回滚建议

- 回退本次提交即可恢复 `fastgit ggc` 入口
