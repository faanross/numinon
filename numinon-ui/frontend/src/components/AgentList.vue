<template>
  <div class="agent-list">
    <div class="header">
      <h2>Agents ({{ agents.length }})</h2>
      <button @click="refreshAgents" :disabled="!isConnected">
        Refresh
      </button>
    </div>

    <div v-if="!isConnected" class="warning">
      Connect to server to view agents
    </div>

    <div v-else class="agents">
      <div
          v-for="agent in agents"
          :key="agent.id"
          class="agent-card"
          :class="{ 'online': agent.status === 'online' }"
          @click="selectAgent(agent)"
      >
        <div class="agent-status"></div>
        <div class="agent-info">
          <div class="agent-name">{{ agent.hostname }}</div>
          <div class="agent-details">
            {{ agent.os }} • {{ agent.ipAddress }}
          </div>
          <div class="agent-lastseen">
            Last seen: {{ formatTime(agent.lastSeen) }}
          </div>
        </div>
        <button
            v-if="selectedAgent?.id === agent.id"
            @click.stop="sendTestCommand"
            class="command-btn"
        >
          Send Command
        </button>
      </div>
    </div>

    <!-- Command Result Display -->
    <div v-if="commandResult" class="command-result" :class="{ 'success': commandResult.success }">
      <h3>Command Result:</h3>
      <pre>{{ commandResult.output || commandResult.error }}</pre>
      <button @click="commandResult = null">Close</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { GetAgents, SendCommand } from '../../wailsjs/go/main/App'
import { models } from '../../wailsjs/go/models'
import { useConnectionStore } from '../stores/connection'
import { storeToRefs } from 'pinia'

const connectionStore = useConnectionStore()
const { isConnected } = storeToRefs(connectionStore)

const agents = ref<models.Agent[]>([])
const selectedAgent = ref<models.Agent | null>(null)
const commandResult = ref<models.CommandResponse | null>(null)

onMounted(() => {
  if (isConnected.value) {
    refreshAgents()
  }
})

async function refreshAgents() {
  if (!isConnected.value) return
  agents.value = await GetAgents()
}

function selectAgent(agent: models.Agent) {
  selectedAgent.value = agent
}

async function sendTestCommand() {
  if (!selectedAgent.value) return

  const request: models.CommandRequest = {
    agentId: selectedAgent.value.id,
    command: 'whoami',
    arguments: ''
  }

  commandResult.value = await SendCommand(request)
}

function formatTime(time: any): string {
  return new Date(time).toLocaleString()
}
</script>

<style scoped>
.agent-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  color: #fff;
  margin: 0;
}

.header button {
  padding: 8px 16px;
  background: #4ade80;
  color: #000;
  border: none;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
}

.header button:disabled {
  background: #444;
  color: #888;
  cursor: not-allowed;
}

.warning {
  padding: 20px;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid #fbbf24;
  border-radius: 8px;
  color: #fbbf24;
  text-align: center;
}

.agents {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.agent-card {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background: #2a2a2a;
  border-radius: 8px;
  border: 2px solid #444;
  cursor: pointer;
  transition: all 0.3s ease;
}

.agent-card:hover {
  border-color: #4ade80;
}

.agent-card.online .agent-status {
  background: #4ade80;
}

.agent-status {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #f87171;
}

.agent-info {
  flex: 1;
}

.agent-name {
  color: #fff;
  font-weight: bold;
  margin-bottom: 4px;
}

.agent-details {
  color: #888;
  font-size: 12px;
}

.agent-lastseen {
  color: #666;
  font-size: 11px;
  margin-top: 4px;
}

.command-btn {
  padding: 6px 12px;
  background: #60a5fa;
  color: #000;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  cursor: pointer;
}

.command-result {
  margin-top: 20px;
  padding: 15px;
  background: #1a1a1a;
  border-radius: 8px;
  border: 2px solid #f87171;
}

.command-result.success {
  border-color: #4ade80;
}

.command-result h3 {
  color: #fff;
  margin: 0 0 10px 0;
  font-size: 14px;
}

.command-result pre {
  color: #ccc;
  font-family: monospace;
  font-size: 12px;
  white-space: pre-wrap;
  margin: 10px 0;
}

.command-result button {
  padding: 6px 12px;
  background: #444;
  color: #fff;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}
</style>