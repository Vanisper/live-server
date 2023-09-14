import { Message } from "@arco-design/web-vue";
import { defineStore } from "pinia";

interface IStream {
    key: string;
    name: string;
    url: string;
    origin_url: string;
    status: "stop" | "success" | "error";
    relay: boolean;
}

export const useHistoryStore = defineStore("history", {
    state: () => [] as IStream[],
    persist: true,
    getters: {
        // historyList: (state) => {
        //     return state
        // }
    },
    actions: {
        addHistory(stream: IStream) {
            // 判断name是否已经存在
            const index = this.$state.findIndex(item => item.name === stream.name)

            if (index == -1) {
                this.$state.push(stream)
            } else {
                Message.error("确保直播源名称唯一")
            }
        },
        setHistoryList(list: IStream[]) {
            this.$state = list
        },
        getHistoryList() {
            return this.$state
        },
        getHistoryItem(key: string) {
            return this.$state.find(item => item.name === key)
        },
        removeHistoryItem(key: string) {
            const index = this.$state.findIndex(item => item.name === key)
            this.$state.splice(index, 1)
        }
    }
})