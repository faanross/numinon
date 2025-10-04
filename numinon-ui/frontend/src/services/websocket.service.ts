// frontend/src/services/websocket.service.ts

import { EventsOn, EventsOff } from '../../wailsjs/runtime';
import {
    ConnectWebSocket,
    DisconnectWebSocket,
    SendWebSocketMessage,
    GetWebSocketStatus,
    ExecuteAgentTask
} from '../../wailsjs/go/main/App';
import { WSMessage, WSState } from '../types/websocket.types';

/**
 * WebSocketService - Manages WebSocket communication through Wails backend
 *
 * Teaching Note: The key issue was using string literals instead of enum values.
 * TypeScript enums are actual objects at runtime, not just types!
 */
export class WebSocketService {
    private state: WSState = WSState.DISCONNECTED;  // FIX: Use enum value
    private eventHandlers: Map<string, (...args: any[]) => void> = new Map();  // FIX: Proper type
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
            this.state = WSState.CONNECTING;  // FIX: Use enum
            console.log('[WS] Connecting...');
        });

        this.on('ws:connected', (data: any) => {
            this.state = WSState.CONNECTED;  // FIX: Use enum
            console.log('[WS] Connected:', data);
            this.processQueuedMessages();
        });

        this.on('ws:disconnected', () => {
            this.state = WSState.DISCONNECTED;  // FIX: Use enum
            console.log('[WS] Disconnected');
        });

        this.on('ws:reconnecting', (data: any) => {
            this.state = WSState.RECONNECTING;  // FIX: Use enum
            console.log(`[WS] Reconnecting... Attempt ${data.attempt}`);
        });

        this.on('ws:error', (error: any) => {
            this.state = WSState.ERROR;  // FIX: Use enum
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
            this.queueMessage({
                id: crypto.randomUUID(),  // FIX: Add required fields
                type: type as any,
                timestamp: new Date().toISOString(),
                payload
            });
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
            this.queueMessage({
                id: crypto.randomUUID(),
                type: type as any,
                timestamp: new Date().toISOString(),
                payload
            });
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
        return this.state === WSState.CONNECTED || this.state === WSState.AUTHENTICATED;
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
     * FIX: Proper typing for Wails event handlers
     */
    on(event: string, handler: (...data: any[]) => void): void {
        // Wrap handler to match Wails signature
        const wrappedHandler = (...args: any[]) => {
            handler(...args);
        };

        EventsOn(event, wrappedHandler);
        this.eventHandlers.set(event, wrappedHandler);
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
        // Find all handlers that match this event pattern
        this.eventHandlers.forEach((handler, key) => {
            if (key === event) {
                handler(data);
            }
        });
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