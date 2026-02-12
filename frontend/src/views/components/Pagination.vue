<template>
    <div>
        <a-button @click="prevPage" :disabled="current === 1" shape="circle" :icon="h(LeftOutlined)" />
        <span class="page-number">{{ current }}</span>
        <a-button @click="nextPage" :disabled="current === total" shape="circle" :icon="h(RightOutlined)" />
    </div>
</template>

<script setup>
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue';
import { h } from 'vue';
const props = defineProps({
    total: {
        type: Number,
        required: true
    },
    current: {
        type: Number,
        required: true
    }
});

const emit = defineEmits(['page-change']);

const prevPage = () => {
    if (props.current > 1) {
        emit('page-change', props.current - 1);
    }
};

const nextPage = () => {
    if (props.current < props.total) {
        emit('page-change', props.current + 1);
    }
};
</script>

<style scoped>
button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.page-number {
    /* width: 40px; */
    /* height: 24px; */
    /* 设置边框 */
    font-size: 18px;
    /* 调整字体大小 */
    border-radius: 4px;
    /* 设置边框圆角 */
    padding-left: 10px;
    padding-right: 10px;
}
</style>