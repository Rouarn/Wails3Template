<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'

const { t } = useI18n()
const router = useRouter()

function goHome(): void {
  void router.push('/')
}
</script>

<template>
  <div class="max-w-lg px-6">
    <n-result status="404" :title="t('notFound.title')" :description="t('notFound.desc')">
      <template #footer>
        <n-button type="primary" @click="goHome">{{ t('common.backHome') }}</n-button>
      </template>
    </n-result>
  </div>
</template>
