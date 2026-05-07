<template>
    <a-row>
        <a-col :span="6">
            <GameHistoryList :puuid="curuuid" :page-size="9" :selected-game-id="rankDetailGameId"
                @game-change="goGame" />
        </a-col>
        <a-col :span="18">
            <RankDetail :gameId="rankDetailGameId" :puuid="curuuid" @checkout-puuid="checkoutUser" />
        </a-col>
    </a-row>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { useStore } from 'vuex'
import RankDetail from './RankDetail.vue';
import GameHistoryList from './components/GameHistoryList.vue';
const rankDetailGameId = ref(0)
let { getters } = useStore()
const selfuuid = computed(() => getters['ws/getUuid'])
const curuuid = ref('')
curuuid.value = selfuuid.value

watch(() => selfuuid.value, (newId, oldId) => {
    curuuid.value = newId
})

const goGame = (gameId) => {
    rankDetailGameId.value = gameId
}
const checkoutUser = (puuid) => {
    curuuid.value = puuid
};
</script>
