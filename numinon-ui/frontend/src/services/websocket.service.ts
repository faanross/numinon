import { EventsOn, EventsOff } from '../../wailsjs/runtime';
import {
    ConnectWebSocket,
    DisconnectWebSocket,
    SendWebSocketMessage,
    GetWebSocketStatus,
    ExecuteAgentTask
} from '../../wailsjs/go/main/App';
import type { WSMessage, WSState, WSConfig } from '../types/websocket.types';

/**
 * WebSocketService - Manages WebSocket communication through Wails backend
 *
 * Architecture Note: The actual WebSocket connection lives in Go.
 * This service is a facade that:
 * 1. Calls Go functions via Wails bindings
 * 2. Listens to events emitted by Go
 * 3. Provides a clean API for Vue components
 */
export class WebSocketService {
    private state: WSState = 'disconnected';
    private eventHandlers: Map<string, Function> = new Map();
    private messageQueue: WSMessage[] = [];

    constructor() {
        this.setupEventListeners();
    }

    /**
     * Sets up listeners for WebSocket events from Go backend
     */
    private setupEventListeners(): void {
        // Connection state events
        this.on('ws:connecting', () => {
            this.state = 'connecting';
            console.log('[WS] Connecting...');
        });

        this.on('ws:connected', (data: any) => {
            this.state = 'connected';
            console.log('[WS] Connected:', data);
            this.processQueuedMessages();
        });

        this.on('ws:disconnected', () => {
            this.state = 'disconnected';
            console.log('[WS] Disconnected');
        });

        this.on('ws:reconnecting', (data: any) => {
            this.state = 'reconnecting';
            console.log(`[WS] Reconnecting... Attempt ${data.attempt}`);
        });

        this.on('ws:error', (error: any) => {
            this.state = 'error';
            console.error('[WS] Error:', error);
        });

        // Message events - dynamically listen for all message types
        const messageTypes = [
            'agent:list',
            'agent:update',
            'task:result',
            'notification'
        ];

        messageTypes.forEach(type => {
            this.on(`ws:message:${type}`, (msg: WSMessage) => {
                this.handleIncomingMessage(msg);
            });
        });
    }

    /**
     * Connects to WebSocket server
     */
    async connect(url: string): Promise<boolean> {
        try {
            const result = await ConnectWebSocket(url);
            return result.success;
        } catch (error) {
            console.error('[WS] Connect failed:', error);
            return false;
        }
    }

    /**
     * Disconnects from WebSocket server
     */
    async disconnect(): Promise<boolean> {
        try {
            const result = await DisconnectWebSocket();
            return result.success;
        } catch (error) {
            console.error('[WS] Disconnect failed:', error);
            return false;
        }
    }

    /**
     * Sends a message through WebSocket
     */
    async send(type: string, payload?: any): Promise<any> {
        // Queue if not connected
        if (!this.isConnected()) {
            this.queueMessage({ type, payload } as WSMessage);
            throw new Error('Not connected - message queued');
        }

        try {
            const result = await SendWebSocketMessage(type, payload);
            if (!result.success) {
                throw new Error(result.error);
            }
            return result.message;
        } catch (error) {
            console.error('[WS] Send failed:', error);
            this.queueMessage({ type, payload } as WSMessage);
            throw error;
        }
    }

    /**
     * Executes a task on an agent
     */
    async executeTask(
        agentId: string,
        taskType: string,
        parameters: Record<string, any>
    ): Promise<string> {
        try {
            const result = await ExecuteAgentTask(agentId, taskType, parameters);
            if (!result.success) {
                throw new Error(result.error);
            }
            return result.taskId; // Return task ID for tracking
        } catch (error) {
            console.error('[WS] Task execution failed:', error);
            throw error;
        }
    }

    /**
     * Gets current WebSocket status
     */
    async getStatus(): Promise<any> {
        return await GetWebSocketStatus();
    }

    /**
     * Checks if connected
     */
    isConnected(): boolean {
        return this.state === 'connected' || this.state === 'authenticated';
    }

    /**
     * Gets current state
     */
    getState(): WSState {
        return this.state;
    }

    /**
     * Handles incoming messages from backend
     */
    private handleIncomingMessage(message: WSMessage): void {
        console.log('[WS] Received message:', message.type, message);

        // Emit to any local listeners
        this.emit(`message:${message.type}`, message);
    }

    /**
     * Queues a message for sending when reconnected
     */
    private queueMessage(message: WSMessage): void {
        this.messageQueue.push(message);

        // Limit queue size
        if (this.messageQueue.length > 100) {
            this.messageQueue.shift(); // Remove oldest
        }
    }

    /**
     * Processes queued messages after reconnection
     */
    private async processQueuedMessages(): Promise<void> {
        const queue = [...this.messageQueue];
        this.messageQueue = [];

        for (const msg of queue) {
            try {
                await this.send(msg.type, msg.payload);
            } catch (error) {
                console.error('[WS] Failed to send queued message:', error);
                break; // Stop on first failure
            }
        }
    }

    /**
     * Registers an event listener
     */
    on(event: string, handler: Function): void {
        EventsOn(event, handler);
        this.eventHandlers.set(event, handler);
    }

    /**
     * Removes an event listener
     */
    off(event: string): void {
        const handler = this.eventHandlers.get(event);
        if (handler) {
            EventsOff(event);
            this.eventHandlers.delete(event);
        }
    }

    /**
     * Emits an event (local only, not to backend)
     */
    private emit(event: string, data?: any): void {
        const handler = this.eventHandlers.get(event);
        if (handler) {
            handler(data);
        }
    }

    /**
     * Cleanup on service destruction
     */
    destroy(): void {
        // Remove all event listeners
        this.eventHandlers.forEach((handler, event) => {
            EventsOff(event);
        });
        this.eventHandlers.clear();
    }
}

// Export singleton instance
export const wsService = new WebSocketService();