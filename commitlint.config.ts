export default {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // 允许中文 subject，不做大小写限制。
    'subject-case': [0],
    'header-max-length': [2, 'always', 100],
  },
}
