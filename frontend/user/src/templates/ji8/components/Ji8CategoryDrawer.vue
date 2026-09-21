<template>
  <Transition name="j8-drawer" :duration="220">
    <div v-if="open" class="j8-drawer">
      <div class="j8-drawer-backdrop" @click="close"></div>
      <div class="j8-drawer-panel" role="dialog" aria-modal="true" :aria-label="t('ji8.categories.title')">
        <header>
          <div>
            <span class="j8-eyebrow">{{ t('ji8.categories.eyebrow') }}</span>
            <h2>{{ t('ji8.categories.title') }}</h2>
          </div>
          <button type="button" class="j8-drawer-close" :aria-label="t('ji8.categories.close')" @click="close"><X /></button>
        </header>
        <Ji8CategoryList :groups="groups" :selected="selected" :total="total" :counts="counts" @select="onSelect" />
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { X } from 'lucide-vue-next'
import type { CategoryGroup } from '../../../utils/category'
import Ji8CategoryList from './Ji8CategoryList.vue'

const props = defineProps<{
  open: boolean
  groups: CategoryGroup[]
  selected: number | null
  total: number
  counts: Record<number, number>
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  select: [id: number | null]
}>()

const { t } = useI18n()

const close = () => emit('update:open', false)

const onSelect = (id: number | null) => {
  emit('select', id)
  close()
}

// 抽屉打开时锁定页面滚动
watch(
  () => props.open,
  (value) => {
    document.body.style.overflow = value ? 'hidden' : ''
  },
)

onUnmounted(() => {
  document.body.style.overflow = ''
})
</script>

<style scoped>
/* 根节点 .j8-drawer 本身没有 transition，Vue 只读根元素的过渡时长，故模板上显式 :duration="220" */
.j8-drawer-enter-active .j8-drawer-panel,
.j8-drawer-leave-active .j8-drawer-panel {
  transition: transform 0.22s ease;
}
.j8-drawer-enter-from .j8-drawer-panel,
.j8-drawer-leave-to .j8-drawer-panel {
  transform: translateX(-100%);
}
.j8-drawer-enter-active .j8-drawer-backdrop,
.j8-drawer-leave-active .j8-drawer-backdrop {
  transition: opacity 0.18s ease;
}
.j8-drawer-enter-from .j8-drawer-backdrop,
.j8-drawer-leave-to .j8-drawer-backdrop {
  opacity: 0;
}
</style>
