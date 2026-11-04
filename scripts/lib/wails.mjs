// Wails3 CLI 解析与版本门禁：所有入口（dev / build / package / bindings）共用。
// 低版本（如 v3.0.0-alpha.3）缺少 -config、-port、-clean、-ts 等参数，
// 且生成的 bindings 与本项目契约不兼容，因此这里直接中止，不做参数降级。
import { spawnSync } from 'node:child_process'
import process from 'node:process'

export const MIN_WAILS_VERSION = '3.1.0'
export const MIN_WAILS_VERSION_TAG = `v${MIN_WAILS_VERSION}`
const INSTALL_COMMAND = 'go install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0'

/** 解析要使用的 wails3 二进制：优先 WAILS3_BIN，其次 PATH 中的 wails3。 */
export function wailsBin() {
  return process.env.WAILS3_BIN?.trim() || 'wails3'
}

/** 从 `wails3 version` 输出中提取语义化版本（保留预发布后缀）。 */
function parseVersion(output) {
  const match = output.match(/v(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?/)
  if (!match) {
    return null
  }
  return {
    raw: match[0],
    core: [Number(match[1]), Number(match[2]), Number(match[3])],
    // 无预发布后缀时视为正式版，排序高于同号预发布版。
    prerelease: match[4] ?? null,
  }
}

/** 语义化版本比较：返回 -1 / 0 / 1。 */
function compareVersion(left, right) {
  for (let index = 0; index < 3; index += 1) {
    if (left.core[index] !== right.core[index]) {
      return left.core[index] > right.core[index] ? 1 : -1
    }
  }
  if (left.prerelease === right.prerelease) {
    return 0
  }
  if (left.prerelease === null) {
    return 1
  }
  if (right.prerelease === null) {
    return -1
  }
  return left.prerelease > right.prerelease ? 1 : -1
}

/** 读取版本；二进制不存在或输出无法解析时抛出带中文提示的错误。 */
export function ensureWailsCli() {
  const bin = wailsBin()
  // wails3 version 把版本号写到 stderr（stdout 为空），因此两个通道都要收。
  const result = spawnSync(bin, ['version'], { encoding: 'utf8' })
  if (result.error) {
    const reason = result.error.code === 'ENOENT' ? '找不到可执行文件' : `执行失败（${result.error.message}）`
    throw new Error(
      [
        `[cli] ${reason}：${bin}`,
        `[cli] 本项目要求 Wails3 CLI >= ${MIN_WAILS_VERSION_TAG}，请安装：${INSTALL_COMMAND}`,
        '[cli] 也可以用 WAILS3_BIN=/path/to/wails3 指定自定义二进制。',
      ].join('\n'),
    )
  }
  if (result.status !== 0) {
    throw new Error(
      [
        `[cli] \`${bin} version\` 执行失败（exit=${result.status ?? 'unknown'}）：${`${result.stderr}`.trim()}`,
        `[cli] 本项目要求 >= ${MIN_WAILS_VERSION_TAG}，请安装：${INSTALL_COMMAND}`,
      ].join('\n'),
    )
  }

  const output = `${result.stdout ?? ''}${result.stderr ?? ''}`
  const version = parseVersion(output)
  if (!version) {
    throw new Error(
      [
        `[cli] 无法从 \`${bin} version\` 的输出解析版本：${`${output}`.trim()}`,
        `[cli] 本项目要求 >= ${MIN_WAILS_VERSION_TAG}，请安装：${INSTALL_COMMAND}`,
      ].join('\n'),
    )
  }

  const required = { core: [3, 1, 0], prerelease: null }
  if (compareVersion(version, required) < 0) {
    throw new Error(
      [
        `[cli] Wails3 CLI 版本过低：当前 ${version.raw}（${bin}），要求 >= ${MIN_WAILS_VERSION_TAG}。`,
        '[cli] 低版本缺少 -config / -port / -clean / -ts 等参数（典型报错：flag provided but not defined: -config），',
        '[cli] 且生成的 bindings 与本项目契约不兼容，因此不会自动降级参数。',
        `[cli] 升级命令：${INSTALL_COMMAND}`,
        '[cli] 或指定自定义二进制：WAILS3_BIN=/path/to/wails3。',
      ].join('\n'),
    )
  }

  return { bin, version: version.raw }
}

/** 统一的失败出口：打印中文提示并以非 0 退出。 */
export function abortWith(message) {
  console.error(message)
  process.exit(1)
}
