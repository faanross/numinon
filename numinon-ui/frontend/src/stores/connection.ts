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
    const status = reactive<frontend.ConnectionStatusDTO>(
        frontend.ConnectionStatusDTO.createFrom({
            connected: false,
            latency: 0,
            serverUrl: '',
            lastPing: new Date().toISOString(),
            error: ''
        })
    );
    const agents = ref<frontend.AgentDTO[]>([]);
    
    const serverMessages = ref<frontend.ServerMessageDTO[]>([]);

    const pingHistory = ref<{ latency: number, timestamp: string }[]>([]);

    // --- Computed Properties ---
    const isConnected = computed(() => status.connected);

    const averageLatency = computed(() => {
        if (pingHistory.value.length === 0) return 0;
        const total = pingHistory.value.reduce((acc, ping) => acc + ping.latency, 0);
        return Math.round(total / pingHistory.value.length);
    });

    // --- Actions ---
    function setupWailsListeners() {
        console.log('Setting up Wails event listeners...');

        // Listen for connection status updates
        EventsOn('connection_status', (newStatus: frontend.ConnectionStatusDTO) => {
            console.log('Received connection_status:', newStatus);
            Object.assign(status, newStatus);

            // Track ping history for the chart
            if (newStatus.connected && newStatus.latency > 0) {
                pingHistory.value.push({
                    latency: newStatus.latency,
                    timestamp: newStatus.lastPing
                });
                // Keep only last 50 pings
                if (pingHistory.value.length > 50) {
                    pingHistory.value.shift();
                }
            } else if (!newStatus.connected) {
                // Clear history on disconnect
                pingHistory.value = [];
            }
        });

        // Listen for agent updates
        EventsOn('agent_update', (updatedAgents: frontend.AgentDTO[]) => {
            console.log('Received agent_update:', updatedAgents);
            agents.value = updatedAgents;
        });

        // Listen for server messages
        EventsOn('server_message', (message: frontend.ServerMessageDTO) => {
            console.log('Received server_message:', message);
            serverMessages.value.unshift(message);
            // Keep only last 100 messages
            if (serverMessages.value.length > 100) {
                serverMessages.value.pop();
            }
        });
    }

    async function connect(serverUrl: string) {
        console.log('Connecting to:', serverUrl);
        const result = await Connect(serverUrl);
        Object.assign(status, result);

        // If connection successful, fetch agents
        if (result.connected) {
            console.log('Connection successful, fetching agents...');
            await refreshAgents();
        }

        return status;
    }

    async function disconnect() {
        console.log('Disconnecting...');
        const result = await Disconnect();
        Object.assign(status, result);
        // Clear agents on disconnect
        agents.value = [];
        return status;
    }

    async function refreshStatus() {
        const result = await GetConnectionStatus();
        Object.assign(status, result);
    }

    async function refreshAgents() {
        console.log('Refreshing agents...');
        try {
            const result = await GetAgents();
            console.log('Got agents:', result);
            agents.value = result;
        } catch (error) {
            console.error('Failed to refresh agents:', error);
        }
    }

    async function sendCommand(agentId: string, command: string, args: string) {
        console.log(`Sending command '${command}' to agent ${agentId}`);
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