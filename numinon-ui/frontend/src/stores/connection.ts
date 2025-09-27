import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime'
import { Connect, Disconnect, GetConnectionStatus } from '../../wailsjs/go/main/App'
import { models } from '../../wailsjs/go/models'

export const useConnectionStore = defineStore('connection', () => {
    // State
    // THIS IS THE CORRECTED PART
    const status = ref<models.ConnectionStatus>(new models.ConnectionStatus({
        connected: false,
        serverUrl: 'ws://localhost:8080/client',
        lastPing: new Date(),
        latency: 0,
        error: ''
    }))

    const serverMessages = ref<models.ServerMessage[]>([])
    const pingHistory = ref<{timestamp: Date, latency: number}[]>([])

    // Computed
    const isConnected = computed(() => status.value.connected)
    const averageLatency = computed(() => {
        if (pingHistory.value.length === 0) return 0
        const sum = pingHistory.value.reduce((acc, p) => acc + p.latency, 0)
        return Math.round(sum / pingHistory.value.length)
    })

    // Actions
    async function connect(url?: string) {
        const result = await Connect(url || status.value.serverUrl)
        status.value = result
        return result
    }

    async function disconnect() {
        const result = await Disconnect()
        status.value = result
        serverMessages.value = [] // Clear messages on disconnect
        pingHistory.value = []
        return result
    }

    async function refreshStatus() {
        const result = await GetConnectionStatus()
        status.value = result
        return result
    }

    // Event Listeners
    function setupEventListeners() {
        // Connection status updates
        EventsOn('connection:status', (newStatus: models.ConnectionStatus) => {
            console.log('Received connection:status event:', newStatus)
            status.value = newStatus
        })

        // Connection established
        EventsOn('connection:established', (newStatus: models.ConnectionStatus) => {
            console.log('Connection established!', newStatus)
            status.value = newStatus
        })

        // Connection failed
        EventsOn('connection:failed', (newStatus: models.ConnectionStatus) => {
            console.error('Connection failed:', newStatus.error)
            status.value = newStatus
        })

        // Connection lost
        EventsOn('connection:disconnected', (newStatus: models.ConnectionStatus) => {
            console.log('Connection disconnected')
            status.value = newStatus
        })

        // Ping updates
        EventsOn('connection:ping', (ping: any) => {
            console.log('Ping received:', ping)
            pingHistory.value.push({
                timestamp: new Date(ping.timestamp),
                latency: ping.latency
            })
            // Keep only last 20 pings
            if (pingHistory.value.length > 20) {
                pingHistory.value.shift()
            }
        })

        // Server messages
        EventsOn('server:message', (message: models.ServerMessage) => {
            console.log('Server message:', message)
            serverMessages.value.unshift(message) // Add to beginning
            // Keep only last 50 messages
            if (serverMessages.value.length > 50) {
                serverMessages.value.pop()
            }
        })
    }

    function cleanupEventListeners() {
        EventsOff('connection:status')
        EventsOff('connection:established')
        EventsOff('connection:failed')
        EventsOff('connection:disconnected')
        EventsOff('connection:ping')
        EventsOff('server:message')
    }

    return {
        // State
        status,
        serverMessages,
        pingHistory,

        // Computed
        isConnected,
        averageLatency,

        // Actions
        connect,
        disconnect,
        refreshStatus,
        setupEventListeners,
        cleanupEventListeners
    }
})