<template>
    <video ref="video" class="video-js"></video>
</template>

<script lang="ts" setup>
import videojs from 'video.js';
import { onMounted, onUnmounted, ref } from 'vue';
const video = ref<HTMLVideoElement>();
const player = ref<any>();
const props = defineProps({
    options: {
        type: Object,
        default: () => ({}),
    },
})
onMounted(() => {
    player.value = videojs(video.value!, props.options, () => {
        console.log('player is ready')
    })
})
onUnmounted(() => {
    player.value?.dispose()
})
</script>

<style lang="less" scoped>
/* styles here */
</style>