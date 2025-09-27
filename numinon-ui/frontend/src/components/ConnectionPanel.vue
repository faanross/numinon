<template>
  <div class="connection-panel">
    <!-- Connection Status Card -->
    <div class="status-card" :class="{ connected: isConnected, disconnected: !isConnected }">
      <div class="status-header">
        <div class="status-indicator" :class="{ 'online': isConnected }"></div>
        <span class="status-text">{{ isConnected ? 'Connected' : 'Disconnected' }}</span>
        <span v-if="isConnected" class="latency">{{ status.latency }}ms</span>
      </div>

      <!-- Connection Controls -->
      <div class="connection-controls">
        <input
            v-model="serverUrl"
            placeholder="Server URL"
            :disabled="isConnected"
            @keyup.enter="handleConnect"
        />
        <button
            @click="handleConnect"
            :class="{ 'danger': isConnected }"
        >
          {{ isConnected ? 'Disconnect' : 'Connect' }}
        </button>
      </div>

      <!-- Error Display -->
      <div v-if="status.error" class="error-message">
        {{ status.error }}
      </div>
    </div>

    <!-- Ping History Chart -->
    <div v-if="isConnected" class="ping-chart">
      <h3>Latency History (avg: {{ averageLatency }}ms)</h3>
      <div class="chart">
        <div
            v-for="(ping, index) in pingHistory"
            :key="index"
            class="ping-bar"
            :style="{ height: `${Math.min(100, (ping.latency / 200) * 100)}%` }"
            :title="`${ping.latency}ms at ${new Date(ping.timestamp).toLocaleTimeString()}`"
        ></div>
      </div>
    </div>

    <!-- Server Messages -->
    <div class="messages-panel">
      <h3>Server Events ({{ serverMessages.length }})</h3>
      <div class="messages-list">
        <div
            v-for="(msg, index) in serverMessages.slice(0, 10)"
            :key="index"
            class="message"
            :class="`message-${msg.type.split(':')[0]}`"
        >
          <span class="message-time">
            {{ new Date(msg.timestamp).toLocaleTimeString() }}
          </span>
          <span class="message-type">{{ msg.type }}</span>
          <span class="message-preview">
            {{ JSON.stringify(msg.payload).substring(0, 50) }}...
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConnectionStore } from '../stores/connection'
import { storeToRefs } from 'pinia'

const connectionStore = useConnectionStore()
// storeToRefs correctly pulls all state and computed properties from the updated store
const { status, isConnected, serverMessages, pingHistory, averageLatency } = storeToRefs(connectionStore)

const serverUrl = ref('ws://localhost:8080/client')

onMounted(() => {
  // Use the correct function name from the store
  connectionStore.setupWailsListeners()
  // Get initial status on component load
  connectionStore.refreshStatus()
})

async function handleConnect() {
  if (isConnected.value) {
    await connectionStore.disconnect()
  } else {
    await connectionStore.connect(serverUrl.value)
  }
}
</script>

<style scoped>
.connection-panel {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.status-card {
  background: #2a2a2a;
  border-radius: 8px;
  padding: 15px;
  border: 2px solid #444;
  transition: all 0.3s ease;
}

.status-card.connected {
  border-color: #4ade80;
  box-shadow: 0 0 10px rgba(74, 222, 128, 0.3);
}

.status-card.disconnected {
  border-color: #f87171;
}

.status-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
}

.status-indicator {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #f87171;
  transition: all 0.3s ease;
}

.status-indicator.online {
  background: #4ade80;
  box-shadow: 0 0 10px #4ade80;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.5; }
  100% { opacity: 1; }
}

.status-text {
  font-weight: bold;
  color: #fff;
}

.latency {
  margin-left: auto;
  color: #4ade80;
  font-family: monospace;
}

.connection-controls {
  display: flex;
  gap: 10px;
}

.connection-controls input {
  flex: 1;
  padding: 8px 12px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
}

.connection-controls button {
  padding: 8px 20px;
  background: #4ade80;
  color: #000;
  border: none;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s ease;
}

.connection-controls button:hover {
  background: #22c55e;
}

.connection-controls button.danger {
  background: #f87171;
}

.connection-controls button.danger:hover {
  background: #ef4444;
}

.error-message {
  margin-top: 10px;
  padding: 10px;
  background: rgba(248, 113, 113, 0.1);
  border: 1px solid #f87171;
  border-radius: 4px;
  color: #f87171;
}

.ping-chart {
  background: #2a2a2a;
  border-radius: 8px;
  padding: 15px;
}

.ping-chart h3 {
  margin: 0 0 10px 0;
  color: #fff;
  font-size: 14px;
}

.chart {
  height: 60px;
  display: flex;
  align-items: flex-end;
  gap: 2px;
  padding: 5px;
  background: #1a1a1a;
  border-radius: 4px;
}

.ping-bar {
  flex: 1;
  background: linear-gradient(to top, #4ade80, #22c55e);
  border-radius: 2px 2px 0 0;
  transition: height 0.3s ease;
  min-height: 2px;
}

.messages-panel {
  background: #2a2a2a;
  border-radius: 8px;
  padding: 15px;
}

.messages-panel h3 {
  margin: 0 0 10px 0;
  color: #fff;
  font-size: 14px;
}

.messages-list {
  max-height: 200px;
  overflow-y: auto;
}

.message {
  display: flex;
  gap: 10px;
  padding: 8px;
  background: #1a1a1a;
  border-radius: 4px;
  margin-bottom: 5px;
  font-size: 12px;
  border-left: 3px solid #444;
}

.message-agent {
  border-left-color: #4ade80;
}

.message-task {
  border-left-color: #60a5fa;
}

.message-time {
  color: #888;
  font-family: monospace;
}

.message-type {
  color: #fbbf24;
  font-weight: bold;
}

.message-preview {
  color: #ccc;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>