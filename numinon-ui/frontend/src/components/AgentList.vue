<script setup lang="ts">
import { ref } from 'vue';
import { useConnectionStore } from '../stores/connection';

// IMPORTANT: Change the import here to the 'frontend' namespace.
import { frontend } from '../../wailsjs/go/models';

const store = useConnectionStore();

// BEFORE: const selectedAgent = ref<models.Agent | null>(null);
// AFTER: Use the AgentDTO type.
const selectedAgent = ref<frontend.AgentDTO | null>(null);

// BEFORE: const commandResponse = ref<models.CommandResponse | null>(null);
// AFTER: Use the CommandResponseDTO type.
const commandResponse = ref<frontend.CommandResponseDTO | null>(null);

const commandInput = ref('whoami');

// BEFORE: function selectAgent(agent: models.Agent) { ... }
// AFTER: The agent parameter is now an AgentDTO.
function selectAgent(agent: frontend.AgentDTO) {
  selectedAgent.value = agent;
  commandResponse.value = null; // Clear previous response
}

async function handleSendCommand() {
  if (selectedAgent.value) {
    commandResponse.value = await store.sendCommand(
        selectedAgent.value.id,
        commandInput.value,
        '' // Assuming arguments might be added later
    );
  }
}
</script>

<template>
  <div class="agent-list-panel">
    <h2>Agents</h2>
    <ul v-if="store.agents.length > 0">
      <li
          v-for="agent in store.agents"
          :key="agent.id"
          @click="selectAgent(agent)"
          :class="{ selected: selectedAgent?.id === agent.id }"
      >
        <span>{{ agent.hostname }} ({{ agent.ipAddress }})</span>
        <span>{{ agent.os }}</span>
      </li>
    </ul>
    <p v-else>No agents connected.</p>

    <div v-if="selectedAgent" class="command-section">
      <h3>Send Command to {{ selectedAgent.hostname }}</h3>
      <input v-model="commandInput" placeholder="Enter command" />
      <button @click="handleSendCommand">Send</button>
      <div v-if="commandResponse" class="response-display">
        <h4>Response:</h4>
        <pre>{{ commandResponse.output || commandResponse.error }}</pre>
      </div>
    </div>
  </div>
</template>

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