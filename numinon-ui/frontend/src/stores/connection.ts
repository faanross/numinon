import {defineStore} from 'pinia';
import {ref, reactive, computed} from 'vue';
import {EventsOn} from '../../wailsjs/runtime';
import {frontend} from '../../wailsjs/go/models';
import {
    Connect,
    Disconnect,
    GetAgents,
    GetConnectionStatus,
    SendCommand
} from '../../wailsjs/go/main/App';

export const useConnectionStore = defineStore('connection', () => {
    // --- State ---
    const status = reactive<frontend.ConnectionStatusDTO>(frontend.ConnectionStatusDTO.createFrom({ connected: false, latency: 0 }));
    const agents = ref<frontend.AgentDTO[]>([]);
    const serverMessages = ref<frontend.ServerMessageDTO[]>([]);
    const pingHistory = ref<{ latency: number, timestamp: string }[]>([]);

    // --- Computed Properties (Getters) ---
    const isConnected = computed(() => status.connected);

    const averageLatency = computed(() => {
        if (pingHistory.value.length === 0) return 0;
        const total = pingHistory.value.reduce((acc, ping) => acc + ping.latency, 0);
        return Math.round(total / pingHistory.value.length);
    });


    // --- Actions ---
    function setupWailsListeners() {
        EventsOn('connection_status', (newStatus: frontend.ConnectionStatusDTO) => {
            Object.assign(status, newStatus);
            // Add new pings to the history for the chart
            if (newStatus.connected) {
                pingHistory.value.push({ latency: newStatus.latency, timestamp: newStatus.lastPing });
                // Keep the history to a reasonable size, e.g., last 50 pings
                if (pingHistory.value.length > 50) {
                    pingHistory.value.shift();
                }
            } else {
                // Clear history on disconnect
                pingHistory.value = [];
            }
        });

        EventsOn('agent_update', (updatedAgents: frontend.AgentDTO[]) => {
            agents.value = updatedAgents;
        });

        EventsOn('server_message', (message: frontend.ServerMessageDTO) => {
            serverMessages.value.unshift(message);
        });
    }

    async function connect(serverUrl: string) {
        const result = await Connect(serverUrl);
        Object.assign(status, result);
    }

    async function disconnect() {
        const result = await Disconnect();
        Object.assign(status, result);
    }

    async function refreshStatus() {
        const result = await GetConnectionStatus();
        Object.assign(status, result);
    }

    async function refreshAgents() {
        agents.value = await GetAgents();
    }

    async function sendCommand(agentId: string, command: string, args: string) {
        return await SendCommand({agentId, command, arguments: args});
    }

    return {
        // State
        status,
        agents,
        serverMessages,
        pingHistory,
        // Computed
        isConnected,
        averageLatency,
        // Actions
        connect,
        disconnect,
        refreshStatus,
        refreshAgents,
        sendCommand,
        setupWailsListeners,
    };
});

