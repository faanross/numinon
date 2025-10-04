// frontend/src/composables/useWebSocket.ts

import { ref, computed, onMounted, onUnmounted } from 'vue';
import { wsService } from '../services/websocket.service';
import { WSState, WSMessage } from '../types/websocket.types';

/**
 * Vue Composable for WebSocket functionality
 *
 * Usage in component:
 * const { isConnected, state, connect, disconnect, sendMessage } = useWebSocket();
 */
export function useWebSocket() {
    // Reactive state - FIX: Use enum values
    const state = ref<WSState>(WSState.DISCONNECTED);
    const isConnecting = ref(false);
    const error = ref<string | null>(null);
    const latency = ref<number>(0);
    const reconnectAttempt = ref<number>(0);

    // Computed properties
    const isConnected = computed(() =>
        state.value === WSState.CONNECTED || state.value === WSState.AUTHENTICATED
    );

    const isReconnecting = computed(() => state.value === WSState.RECONNECTING);

    /**
     * Connects to WebSocket server
     */
    const connect = async (url: string): Promise<boolean> => {
        if (isConnecting.value || isConnected.value) {
            console.warn('[useWebSocket] Already connected or connecting');
            return false;
        }

        isConnecting.value = true;
        error.value = null;

        try {
            const success = await wsService.connect(url);
            if (success) {
                state.value = WSState.CONNECTED;  // FIX: Use enum
                console.log('[useWebSocket] Connected successfully');
            } else {
                throw new Error('Connection failed');
            }
            return success;
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Connection failed';
            state.value = WSState.ERROR;  // FIX: Use enum
            return false;
        } finally {
            isConnecting.value = false;
        }
    };

    /**
     * Disconnects from WebSocket server
     */
    const disconnect = async (): Promise<boolean> => {
        try {
            const success = await wsService.disconnect();
            if (success) {
                state.value = WSState.DISCONNECTED;  // FIX: Use enum
            }
            return success;
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Disconnect failed';
            return false;
        }
    };

    /**
     * Sends a message through WebSocket
     */
    const sendMessage = async (type: string, payload?: any): Promise<any> => {
        if (!isConnected.value) {
            throw new Error('Not connected');
        }

        return await wsService.send(type, payload);
    };

    /**
     * Executes a task on an agent
     */
    const executeAgentTask = async (
        agentId: string,
        taskType: string,
        parameters: Record<string, any>
    ): Promise<string> => {
        if (!isConnected.value) {
            throw new Error('Not connected');
        }

        return await wsService.executeTask(agentId, taskType, parameters);
    };

    /**
     * Sets up event listeners
     */
    const setupListeners = () => {
        // State changes - FIX: Use enum values in all handlers
        wsService.on('ws:connecting', () => {
            state.value = WSState.CONNECTING;  // FIX: Use enum
            reconnectAttempt.value = 0;
        });

        wsService.on('ws:connected', () => {
            state.value = WSState.CONNECTED;  // FIX: Use enum
            error.value = null;
            reconnectAttempt.value = 0;
        });

        wsService.on('ws:disconnected', () => {
            state.value = WSState.DISCONNECTED;  // FIX: Use enum
        });

        wsService.on('ws:reconnecting', (data: any) => {
            state.value = WSState.RECONNECTING;  // FIX: Use enum
            reconnectAttempt.value = data.attempt;
        });

        wsService.on('ws:error', (err: any) => {
            state.value = WSState.ERROR;  // FIX: Use enum
            error.value = err.error || 'WebSocket error';
        });

        // Latency updates
        wsService.on('ws:latency', (data: any) => {
            latency.value = data.latency;
        });
    };

    /**
     * Cleanup listeners
     */
    const cleanup = () => {
        wsService.off('ws:connecting');
        wsService.off('ws:connected');
        wsService.off('ws:disconnected');
        wsService.off('ws:reconnecting');
        wsService.off('ws:error');
        wsService.off('ws:latency');
    };

    // Lifecycle hooks
    onMounted(() => {
        setupListeners();
    });

    onUnmounted(() => {
        cleanup();
    });

    return {
        // State
        state,
        isConnected,
        isConnecting,
        isReconnecting,
        error,
        latency,
        reconnectAttempt,

        // Methods
        connect,
        disconnect,
        sendMessage,
        executeAgentTask,
    };
}

/**
 * Composable for listening to specific WebSocket messages
 */
export function useWebSocketMessages(messageType: string) {
    const messages = ref<WSMessage[]>([]);
    const latestMessage = ref<WSMessage | null>(null);

    const handleMessage = (msg: WSMessage) => {
        messages.value.push(msg);
        latestMessage.value = msg;

        // Keep only last 100 messages
        if (messages.value.length > 100) {
            messages.value = messages.value.slice(-100);
        }
    };

    onMounted(() => {
        wsService.on(`message:${messageType}`, handleMessage);
    });

    onUnmounted(() => {
        wsService.off(`message:${messageType}`);
    });

    return {
        messages,
        latestMessage,
    };
}