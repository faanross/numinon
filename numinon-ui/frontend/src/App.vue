
<template>
  <div id="app">
    <div class="header">
      <h1>🎯 Numinon C2 Client</h1>
      <div class="subtitle">Command & Control Interface</div>
    </div>

    <div class="main-layout">
      <!-- Left Panel: Connection & Events -->
      <div class="left-panel">
        <ConnectionPanel />
      </div>

      <!-- Right Panel: Agents -->
      <div class="right-panel">
        <AgentList />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import ConnectionPanel from './components/ConnectionPanel.vue'
import AgentList from './components/AgentList.vue'
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  background: #1a1a1a;
  color: #fff;
}

#app {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  text-align: center;
  box-shadow: 0 2px 10px rgba(0,0,0,0.5);
}

.header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: bold;
}

.subtitle {
  opacity: 0.9;
  font-size: 12px;
  margin-top: 5px;
}

.main-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.left-panel {
  width: 400px;
  border-right: 1px solid #333;
  overflow-y: auto;
}

.right-panel {
  flex: 1;
  overflow-y: auto;
}

/* Custom Scrollbar */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #1a1a1a;
}

::-webkit-scrollbar-thumb {
  background: #444;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>