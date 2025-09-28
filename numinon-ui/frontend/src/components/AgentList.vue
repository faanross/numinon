<template>
  <div class="agent-list">
    <div class="header">
      <h2>Connected Agents ({{ store.agents.length }})</h2>
      <button @click="store.refreshAgents" :disabled="!store.isConnected">
        Refresh
      </button>
    </div>

    <div v-if="!store.isConnected" class="warning">
      ⚠️ Not connected to server
    </div>

    <div v-else-if="store.agents.length === 0" class="warning">
      No agents connected yet. Waiting for agents...
    </div>

    <div v-else class="agents-grid">
      <div
          v-for="agent in store.agents"
          :key="agent.id"
          @click="selectAgent(agent)"
          :class="['agent-card', { selected: selectedAgent?.id === agent.id }]"
      >
        <div class="agent-status" :class="agent.status"></div>
        <div class="agent-info">
          <div class="agent-name">{{ agent.hostname }}</div>
          <div class="agent-details">
            <span>{{ agent.os }}</span> • <span>{{ agent.ipAddress }}</span>
          </div>
          <div class="agent-id">ID: {{ agent.id }}</div>
        </div>
      </div>
    </div>

    <!-- Command Panel -->
    <div v-if="selectedAgent" class="command-panel">
      <h3>Command Interface - {{ selectedAgent.hostname }}</h3>
      <div class="command-input-group">
        <input
            v-model="commandInput"
            placeholder="Enter command (e.g., whoami, hostname, pwd)"
            @keyup.enter="handleSendCommand"
            class="command-input"
        />
        <button @click="handleSendCommand" class="send-btn">
          Execute →
        </button>
      </div>

      <div v-if="commandResponse" class="command-result" :class="{ success: commandResponse.success }">
        <div class="result-header">
          <span>{{ commandResponse.success ? '✅ Success' : '❌ Failed' }}</span>
        </div>
        <pre class="result-output">{{ commandResponse.output || commandResponse.error }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useConnectionStore } from '../stores/connection';
import { frontend } from '../../wailsjs/go/models';

const store = useConnectionStore();

const selectedAgent = ref<frontend.AgentDTO | null>(null);
const commandResponse = ref<frontend.CommandResponseDTO | null>(null);
const commandInput = ref('whoami');

function selectAgent(agent: frontend.AgentDTO) {
  selectedAgent.value = agent;
  commandResponse.value = null;
}

async function handleSendCommand() {
  if (selectedAgent.value && commandInput.value.trim()) {
    commandResponse.value = await store.sendCommand(
        selectedAgent.value.id,
        commandInput.value.trim(),
        ''
    );
  }
}
</script>

<style scoped>
.agent-list {
  padding: 20px;
  display: flex;
  flex-direction: column;
  height: 100vh;
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
  transition: all 0.3s ease;
}

.header button:hover:not(:disabled) {
  background: #22c55e;
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

.agents-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
  max-height: 300px;
  overflow-y: auto;
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
  border-color: #60a5fa;
  transform: translateY(-2px);
}

.agent-card.selected {
  border-color: #4ade80;
  background: #2a3a2a;
  box-shadow: 0 0 20px rgba(74, 222, 128, 0.3);
}

.agent-status {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.agent-status.online {
  background: #4ade80;
  box-shadow: 0 0 10px #4ade80;
  animation: pulse 2s infinite;
}

.agent-status.offline {
  background: #f87171;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
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
  margin-bottom: 4px;
}

.agent-id {
  color: #666;
  font-size: 11px;
  font-family: monospace;
}

.command-panel {
  flex: 1;
  background: #2a2a2a;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.command-panel h3 {
  color: #fff;
  margin: 0 0 15px 0;
  padding-bottom: 10px;
  border-bottom: 1px solid #444;
}

.command-input-group {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.command-input {
  flex: 1;
  padding: 10px;
  background: #1a1a1a;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-family: monospace;
}

.command-input:focus {
  outline: none;
  border-color: #60a5fa;
}

.send-btn {
  padding: 10px 20px;
  background: #60a5fa;
  color: #000;
  border: none;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s ease;
}

.send-btn:hover {
  background: #3b82f6;
}

.command-result {
  flex: 1;
  background: #1a1a1a;
  border-radius: 8px;
  border: 2px solid #f87171;
  overflow: hidden;
}

.command-result.success {
  border-color: #4ade80;
}

.result-header {
  padding: 10px 15px;
  background: rgba(0,0,0,0.5);
  border-bottom: 1px solid #333;
  font-weight: bold;
}

.result-output {
  padding: 15px;
  margin: 0;
  color: #4ade80;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
  overflow-y: auto;
  max-height: 300px;
}

.command-result:not(.success) .result-output {
  color: #f87171;
}
</style>