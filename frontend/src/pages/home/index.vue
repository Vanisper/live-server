<template>
  <div style="height: 100%;">
    <a-button @click="statAllGroup">获取所有视频流信息</a-button>
    <div style="display: flex;justify-content: center;margin-top: 20px;">
      <a-button type="text" style="width: 100%;">
        <template #icon>
          <icon-plus />
        </template>
      </a-button>
    </div>
    <a-table :columns="columns" :data="historyStore.$state" :bordered="true" :hoverable="true" :stripe="false"
      :loading="false" :show-header="true" :row-selection="{
        type: 'checkbox',
        showCheckedAll: false,
        onlyCurrent: false,
      }" v-model:selected-keys="selectedKeys" row-key="name">
      <template #status="{ record }">
        <div :title="record.status" :style="{
          backgroundColor: record.status == 'success' ? '#009900' : (record.status == 'error' ? '#aa0000' : '#ff8e35')
        }" style="width: 10px;height: 10px;margin: auto;border-radius: 50%;animation: breathing 0.8s infinite;"></div>
      </template>
      <template #operate="{ record }">
        <a-button-group>
          <a-button status="normal" :disabled="!(record.relay && (record.status == 'success'))"
            @click="$modal.info({ title: 'M3U8', width: 'fit-content', content: ModalVideoContent(record.url) })">
            <template #icon>
              <icon-play-circle-fill />
            </template>
            <!-- Use the default slot to avoid extra spaces -->
            <template #default>播放</template>
          </a-button>
          <a-popconfirm content="确定删除此条目?" type="error" @ok="stopStream(record.name, true)" position="tr">
            <a-button :disabled="record.relay && (record.status != 'success')" status="danger">
              <template #icon>
                <icon-delete />
              </template>
            </a-button>
            <!-- <a-button status="danger" @click="stopStream(record.name, true)">
              <template #icon>
                <icon-delete />
              </template>
            </a-button> -->
          </a-popconfirm>
        </a-button-group>
      </template>
    </a-table>

    <div>
      <input type="text" v-model="stream_name">
      <button @click="statGroup">搜索</button>
    </div>
    <div style="display: flex;align-items: center;margin-top: 10px;flex-direction: column;">
      <span style="color:chocolate;" v-if="historyStore.$state.length == 0">暂未拉流</span>
      <a-grid :cols="{ xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }" :colGap="12" :rowGap="16" style="width: 100%;">
        <a-grid-item v-for="( item, index ) in  historyStore.$state " :key="index">
          <div v-if="item.relay" style="display: flex;position: relative;">
            <!-- <video controls muted autoplay :data-origin="item.origin_url" :src="item.url"></video> -->
            <VideoPlayer style="width: 100%;height: 500px;" :options="{
              autoplay: true, muted: true, controls: true,
              sources: [
                { src: item.url, type: 'application/x-mpegURL' }
              ]
            }
              " :data-origin="item.origin_url" :data-src="item.url"></VideoPlayer>
            <button style="background-color: brown;color: aliceblue;position: absolute;"
              @click="stopStream(item.name)">停止</button>
          </div>
          <div
            style="display: flex;justify-content: center;align-items: center;height: 500px;background-color: antiquewhite;"
            v-else>
            <a-spin dot />
          </div>
        </a-grid-item>
        <a-grid-item :span="12" suffix style="display: flex;justify-content: center;">
          <span style="cursor: pointer;" title="添加拉流"><svg @click="addStream" viewBox="0 0 1024 1024" version="1.1"
              xmlns="http://www.w3.org/2000/svg" p-id="4021" width="24" height="24">
              <path fill="#409EFF"
                d="M512 936.915619c-234.672764 0-424.915619-190.243879-424.915619-424.915619S277.327236 87.083357 512 87.083357c234.676857 0 424.916643 190.243879 424.916643 424.915619S746.676857 936.915619 512 936.915619zM724.45781 469.50414 554.491767 469.50414 554.491767 299.546284l-84.983533 0 0 169.957857L299.54219 469.50414l0 84.99172 169.966043 0 0 169.966043 84.983533 0L554.491767 554.49586l169.966043 0L724.45781 469.50414z"
                p-id="4022"></path>
            </svg></span>
        </a-grid-item>
      </a-grid>
    </div>
  </div>
</template>

// https://github.com/vuejs/core/issues/7799
// 解决了setup组件中没有显式定义组件name的问题  路由缓存的include、exclude属性匹配的是组件的name属性
<script lang="ts">
import { PageEnum } from "@/enums/pageEnum";
import { onMounted, ref, watch, h } from "vue";
import { computed } from "vue";

export default {
  name: PageEnum.BASE_HOME_NAME,
};
</script>
<script lang="ts" setup>
//@ts-ignore
import { HttpApiStatGroup, HttpApiStatAllGroup, PullRtsp2PushRtspList, PullRtsp2PushRtspStop, PullRtsp2PushRtspStart, CheckUrl } from "@wailsjs/go/main/App";
import { EventsOn } from "@wailsjs/runtime";
import VideoPlayer from "@/components/video/VideoPlayer.vue";
import { TableColumnData, Message } from "@arco-design/web-vue";
import { useHistoryStore } from "@/stores/modules/useHistory.storer";

const historyStore = useHistoryStore();
const ModalVideoContent = (url: string) => {
  return () => h(VideoPlayer, {
    style: "width: 800px;height: 500px;",
    options: {
      autoplay: true, muted: true, controls: true,
      sources: [
        { src: url, type: 'application/x-mpegURL' }
      ]
    },
    "data-origin": url,
    "data-src": url
  })
}
const selectedKeys = computed({
  get() {
    return historyStore.$state.filter((item) => item.relay).map((item) => item.key);
  },
  set(val) {
    // 将指定的key设置为选中状态
    historyStore.$state.forEach((item) => {
      item.relay = val.includes(item.key);
      // 将不选中的key停止拉流
      if (!val.includes(item.key)) {
        stopStream(item.name);
      }
    })

    return historyStore.$state.filter((item) => item.relay).map((item) => item.key);
  }
});
// 监听勾选变化以及播放列表
watch([selectedKeys, historyStore.$state], ([keys, history]) => {
  // 遍历勾选列表
  keys.forEach((item) => {
    // 当前勾选上的视频流信息
    const historyItem = historyStore.getHistoryItem(item);

    if (historyItem && historyItem.status == "stop") {
      PullRtsp2PushRtspStart(historyItem.origin_url, historyItem.name, 1, 0).then((res) => {
        // 轮询CheckUrl
        const timer = setInterval(() => {
          const item = historyStore.$state.find((item) => item.name == res.name)
          CheckUrl(res.hls_url).then(() => {
            item && (item.status = "success");
            clearInterval(timer);
          }).catch((err) => {
            Message.warning("轮询中: " + err);
            item && (item.status = "error");
          });
        }, 2000);
      }).catch((err) => {
        Message.error(err);
        historyItem.status = "error";
        historyItem.relay = false;
      });
    }
  })
  console.log(history);
}, {
  immediate: true
})

const statAllGroup = async () => {
  historyStore.$state.forEach((res) => {
    // 检测hls地址是否可用
    CheckUrl(res.url).then(() => {
      historyStore.$state.find((item) => item.name == res.name)!.relay = true;
    }).catch(() => {
      historyStore.$state.find((item) => item.name == res.name)!.relay = false;
    })
  })

  const res = await PullRtsp2PushRtspList();
  if (res && res.length != 0) {
    res.forEach((item) => {
      if (historyStore.$state.find((i) => i.name == item.name)) {
        return;
      }
      historyStore.addHistory({
        key: item.name,
        name: item.name,
        url: item.hls_url,
        origin_url: item.origin_url,
        status: "success",
        relay: true,
      });
    });
  }
  // location.reload();
};
const stream_name = ref("");
const statGroup = async () => {
  const res = await HttpApiStatGroup(stream_name.value);
  console.log(stream_name.value, res.data);
};

const addStream = async () => {
  const url = window.prompt("请输入拉流地址", "rtsp://admin:xxbADMIN@112.4.158.2:10001/ch1/main/av_stream");
  if (!url) {
    return;
  }
  // 判断url是否是rtmp/rtsp地址
  if (!url.startsWith("rtmp://") && !url.startsWith("rtsp://")) {
    alert("请输入正确的rtmp/rtsp地址");
    return;
  }
  const name = window.prompt("请输入拉流名称", "111");
  if (!name) {
    return;
  }

  PullRtsp2PushRtspStart(url, name, 1, 0).then((res) => {
    historyStore.addHistory({
      key: res.name,
      name: res.name,
      url: res.hls_url,
      origin_url: res.origin_url,
      status: "stop",
      relay: false,
    })
    // 轮询CheckUrl
    const timer = setInterval(() => {
      const item = historyStore.$state.find((item) => item.name == res.name);
      CheckUrl(res.hls_url).then(() => {
        if (item) {
          item.relay = true;
          item.status = "success";
        }
        clearInterval(timer);
      }).catch((err) => {
        item && (item.status = "error");
        Message.warning(err);
      });
    }, 2000);
  }).catch((err) => {
    Message.error(err);
  });
};

const stopStream = async (name: string, isDelItem?: boolean) => {
  isDelItem = isDelItem || false;
  await PullRtsp2PushRtspStop(name);
  const item = historyStore.getHistoryItem(name);
  item && (item.status = "stop", item.relay = false);
  if (isDelItem) {
    historyStore.removeHistoryItem(name);
  }
};

const columns: TableColumnData[] = [
  {
    title: "名称",
    dataIndex: "name",
    align: "center"
  },
  {
    title: "hls",
    dataIndex: "url",
    align: "center"
  },
  {
    title: "源",
    dataIndex: "origin_url",
    align: "center"
  },
  {
    title: "状态",
    slotName: "status",
    align: "center"
  },
  {
    title: "拉流情况",
    dataIndex: "relay",
    align: "center"
  }, {
    title: "操作",
    slotName: "operate",
    align: "center"
  }
];
onMounted(async () => {
  EventsOn("stream_disconnected", (res: string) => {
    Message.warning(res + "断流了");
    const item = historyStore.getHistoryItem(res);
    item && (item.status = "error", item.relay = false);
  })
  // historyStore.$state.forEach((res) => {
  //   // 检测hls地址是否可用
  //   CheckUrl(res.url).then(() => {
  //     historyStore.$state.find((item) => item.name == res.name)!.relay = true;
  //   }).catch(() => {
  //     historyStore.$state.find((item) => item.name == res.name)!.relay = false;
  //   })
  // })

  // const res = await PullRtsp2PushRtspList();
  // if (res && res.length != 0) {
  //   res.forEach((item) => {
  //     if (historyStore.$state.find((i) => i.name == item.name)) {
  //       return;
  //     }
  //     historyStore.addHistory({
  //       key: item.name,
  //       name: item.name,
  //       url: item.hls_url,
  //       origin_url: item.origin_url,
  //       status: "success",
  //       relay: true,
  //     });
  //   });
  // }
  // await statAllGroup();
  // HttpApiStatAllGroup().then((res) => {
  //   if (res.data.groups && res.data.groups.length != 0) {
  //     res.data.groups.forEach((item) => {
  //       if (streamList.value.find((i) => i.name == item.stream_name)) {
  //         return;
  //       }
  //       streamList.value.push({
  //         name: item.stream_name,
  //         url: `http://127.0.0.0:8080/hls/${item.stream_name}.m3u8`,
  //         origin_url: "",
  //         status: "success",
  //         relay: true,
  //       });
  //     });
  //   }
  // });
})

</script>

<style lang="less">
@keyframes breathing {
  0% {
    opacity: 0.2;
    transform: scale(1.1);
  }

  50% {
    opacity: 1;
    transform: scale(1.5);
  }

  100% {
    opacity: 0.2;
    transform: scale(1.1);
  }
}
</style>
