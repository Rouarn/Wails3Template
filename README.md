# Wails3Template

基于 **Wails v3 + Vite Plus（`vp`）** 的桌面应用模板：Go 单进程承载系统 WebView，前端使用 Vue 3.5 + TypeScript + Naive UI + UnoCSS。

## 环境要求

| 工具 | 版本 | 说明 |
| --- | --- | --- |
| Go | 1.26.0 | Wails v3 官方模板要求（见 `go.mod`） |
| Wails CLI | >= v3.1.0 | `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`（当前解析到 v3.1.0），须与 `go.mod` 一致 |
| Node.js | >= 22.18.0 | 与 pnpm 12 的引擎要求一致 |
| pnpm | 12.16.1 | `corepack enable && corepack prepare pnpm@12.16.1 --activate` |
| Task | 3.x | Wails 构建脚本依赖 `task`（`brew install go-task`） |
| golangci-lint | v2.15.0 | `brew install golangci-lint` 或 `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0` |

> macOS 打包 DMG 需要 Xcode 命令行工具；Windows 安装包需要 NSIS；Linux 打包需要 `nfpm`（由 Wails 工具链按需拉取）。

## CLI 版本要求（Wails3 >= v3.1.0）

`pnpm dev` / `pnpm build` / `pnpm package` / `pnpm check:bindings` / Taskfile 任务都会先校验 CLI 版本（实现见 [scripts/lib/wails.mjs](scripts/lib/wails.mjs)），低于 **v3.1.0** 直接中止并给出中文提示，不会静默降级参数——旧版生成的 bindings 与项目契约不兼容。

```bash
wails3 version        # 查看当前版本，例如 v3.1.0
pnpm cli:check        # 走项目内的校验逻辑，打印实际二进制与版本
```

安装/升级（推荐跟随稳定线，当前 `@latest` = v3.1.0）：

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
which wails3          # 通常是 $(go env GOPATH)/bin/wails3
```

需要固定到某一版本时把 `@latest` 换成 `@v3.1.0`。若机器上还残留旧版（`go/bin/wails3` 之外还有 alpha 版在 PATH 前面），可以显式指定二进制：

```bash
WAILS3_BIN=/path/to/wails3 pnpm dev
```

版本过低时的典型报错（搜这条即可定位）：

```
ERROR   Error: flag provided but not defined: -config
        See 'wails dev --help' for usage
```

`wails3 generate bindings -clean=true -ts -i` 在旧版上则会报 `flag provided but not defined: -clean`（或 `-ts`）；都说明 CLI 需要升级。

## 常用命令

```bash
pnpm install          # 安装 workspace 依赖（含 frontend）
pnpm dev              # 自动挑选空闲端口并启动 wails3 dev
pnpm build            # wails3 build（Taskfile 流程：bindings → 前端 → Go 二进制）
pnpm package          # 当前平台安装包（任务已定义；仓库内仅验证过 macOS .app，.dmg 与 Windows/Linux 产物尚未生成）
pnpm typecheck        # vue-tsc 类型检查
pnpm lint / pnpm format / pnpm format:check
pnpm check:bindings   # 重新生成 bindings 并校验仓库产物无漂移
pnpm check:icons      # 校验图标类名在 iconify 集合里存在、且产物 CSS 里真有对应规则
```

首次运行前先执行 `pnpm install`；`go build ./...` 依赖 `frontend/dist`（`go:embed`），因此完整构建请走 `pnpm build`，或先执行 `pnpm build:frontend`。

## 目录

```
.
├── main.go                 # 入口：embed 前端产物、注册强类型事件
├── internal/
│   ├── app/                # 应用装配：窗口、服务注册、事件推送、CSP 中间件
│   ├── service/            # 服务层：运行信息 / 设置持久化 / 窗口控制
│   ├── contract/           # 单一契约来源：模型、事件、结构化错误
│   ├── infra/              # 基础设施：日志、设置存储、路径校验
│   └── platform/           # 平台差异：配置与日志目录
├── frontend/
│   ├── src/bridge/         # 唯一接触 bindings 的层：调用封装、事件订阅、错误归一化
│   ├── src/views/          # 首页 / 设置 / 404
│   ├── src/layouts/        # 双轨外壳（侧栏 + 主区）与单栏外壳
│   └── bindings/           # wails3 generate bindings 产物，入库并由 CI 防漂移
├── build/                  # Taskfile 与三平台打包资源（无 ios/android）
└── scripts/                # dev.mjs（端口挑选与清理）、wails.mjs（CLI 版本门禁与透传）、lib/wails.mjs
```

## 架构约定

- **契约单一来源**：Go 的 `internal/contract` 是唯一模型来源；前端类型全部来自生成的 `frontend/bindings`，不手写 DTO。
- **生成产物入库**：`frontend/bindings` 提交到仓库；改 Go 服务后执行 `wails3 generate bindings -clean=true -ts -i` 并提交。本地用 `pnpm check:bindings` 防漂移；CI 将在门禁阶段落地（当前仓库尚无 workflow）。
- **图标单一来源**：所有图标类名集中在 `frontend/src/icons.ts` 的 `APP_ICONS`；UnoCSS 的 `safelist` 由该表生成，`scripts/check-icons.mjs` 在构建期校验「集合里存在 + 产物 CSS 里有规则」，动态拼接类名直接报错。
- **错误回传**：Go 侧统一返回 `*contract.AppError`，经 `application.Options.MarshalError` 序列化为 `{code,message,details}`；前端 `src/bridge/errors.ts` 归一化并附加 15s 超时与取消。
- **事件**：`app.clock` 每秒推送时间与运行时长，`app.status` 在启动完成/设置保存时推送；前端只在 `src/bridge/events.ts` 订阅。
- **安全基线**：资源响应带 CSP、导航与外部链接由 `src/bridge/links.ts` 拦截并交给 `SystemService.OpenExternal`（只放行 http/https/mailto）；`internal/infra/path.go` 负责路径越界与符号链接校验。
- **布局**：外壳满高固定，`html/body/#app` 锁 100%，侧栏不滚动，主区独立滚动。

## 已知未验证项

- Windows NSIS 安装包与 MSIX、代码签名：受限于当前 macOS 构建环境，未实机验证。
- macOS Developer ID 签名与公证：需要证书，模板只保留 `task darwin:sign` / `sign:notarize` 入口。
- Linux WebKitGTK 渲染差异：未在 Linux 实机验证。
- 自动更新：未实现，也无接入点（代码中没有任何 `app.Updater` / `updater.Init` 引用）。
