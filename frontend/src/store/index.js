import { createStore } from "vuex";
import ws from "./websocket";
import ui from './ui';

export default createStore({
    state: {},
    mutations: {},
    actions: {},
    modules: { ws, ui },
});