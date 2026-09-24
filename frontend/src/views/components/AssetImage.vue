<template>
  <span class="asset-image" :class="{ 'asset-image--small': small }"
    ><img
      v-if="src && !failed"
      :src="src"
      :alt="label"
      loading="lazy"
      @error="failed = true"
    /><span v-else role="img" :aria-label="label" :title="label">{{
      label?.match(/\d+/)?.[0] || label?.slice(0, 2) || '—'
    }}</span></span
  >
</template>
<script setup>
import { ref, watch } from 'vue'
const props = defineProps({ src: String, label: String, small: Boolean })
const failed = ref(false)
watch(
  () => props.src,
  () => {
    failed.value = false
  },
)
</script>
