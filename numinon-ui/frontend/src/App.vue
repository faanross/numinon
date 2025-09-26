<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {Greet, GetSystemInfo} from '../wailsjs/go/main/App'

const resultText = ref("Please enter your name below 👇")
const name = ref('')
const systemInfo = ref<any>({})

onMounted(() => {
  // Fetch system info when component mounts
  GetSystemInfo().then(info => {
    systemInfo.value = info
  })
})

function greet() {
  Greet(name.value).then(result => {
    resultText.value = result
  })
}
</script>

<template>
  <main>
    <div id="system-info">
      <h3>System Information:</h3>
      <p>OS: {{systemInfo.os}} ({{systemInfo.arch}})</p>
      <p>Host: {{systemInfo.hostname}}</p>
      <p>Time: {{systemInfo.current_time}}</p>
    </div>

    <div id="result">{{resultText}}</div>
    <div id="input">
      <input v-model="name" @keyup.enter="greet" type="text"/>
      <button @click="greet">Greet</button>
    </div>
  </main>
</template>

<style>
#logo {
  display: block;
  width: 50%;
  height: 50%;
  margin: auto;
  padding: 10% 0 0;
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  background-origin: content-box;
}
</style>
