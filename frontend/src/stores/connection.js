import { create } from 'zustand'
import { 
  ListConnections, 
  AddConnection, 
  UpdateConnection, 
  DeleteConnection,
  TestConnection,
  Connect,
  Disconnect,
  GetConnectionStatus
} from '../wailsjs/go/app/App'

export const useConnectionStore = create((set, get) => ({
  // 状态
  connections: [],
  activeConnection: null,
  loading: false,
  error: null,

  // 计算属性
  connectedConnections: () => get().connections.filter(conn => conn.status === 'connected'),

  // 加载连接列表
  loadConnections: async () => {
    try {
      set({ loading: true, error: null })
      const result = await ListConnections()
      const connections = result || []
      
      // 检查连接状态
      for (const conn of connections) {
        const status = await GetConnectionStatus(conn.id)
        conn.status = status.status
        conn.lastPing = status.lastPing
        conn.message = status.message
      }
      
      set({ connections, loading: false })
    } catch (err) {
      set({ 
        error: err.message || '加载连接失败', 
        loading: false 
      })
      console.error('Failed to load connections:', err)
    }
  },

  // 添加连接
  addConnection: async (config) => {
    console.log('addConnection called with config:', config)
    try {
      set({ loading: true, error: null })
      console.log('Calling AddConnection...')
      await AddConnection(config)
      console.log('AddConnection successful, loading connections...')
      await get().loadConnections()
      console.log('addConnection completed successfully')
    } catch (err) {
      console.error('addConnection error:', err)
      set({ 
        error: err.message || '添加连接失败', 
        loading: false 
      })
      throw err
    }
  },

  // 更新连接
  updateConnection: async (config) => {
    try {
      set({ loading: true, error: null })
      await UpdateConnection(config)
      await get().loadConnections()
    } catch (err) {
      set({ 
        error: err.message || '更新连接失败', 
        loading: false 
      })
      throw err
    }
  },

  // 删除连接
  deleteConnection: async (id) => {
    try {
      set({ loading: true, error: null })
      await DeleteConnection(id)
      await get().loadConnections()
      
      // 如果删除的是当前活动连接，清空活动连接
      const { activeConnection } = get()
      if (activeConnection?.id === id) {
        set({ activeConnection: null })
      }
    } catch (err) {
      set({ 
        error: err.message || '删除连接失败', 
        loading: false 
      })
      throw err
    }
  },

  // 测试连接
  testConnection: async (config) => {
    try {
      set({ loading: true, error: null })
      await TestConnection(config)
      set({ loading: false })
      return true
    } catch (err) {
      set({ 
        error: err.message || '连接测试失败', 
        loading: false 
      })
      throw err
    }
  },

  // 连接数据库
  connectToDatabase: async (id) => {
    try {
      set({ loading: true, error: null })
      await Connect(id)
      
      // 更新连接状态
      const status = await GetConnectionStatus(id)
      const { connections } = get()
      const conn = connections.find(c => c.id === id)
      if (conn) {
        conn.status = status.status
        conn.lastPing = status.lastPing
        conn.message = status.message
      }
      
      set({ activeConnection: conn, loading: false })
    } catch (err) {
      set({ 
        error: err.message || '连接失败', 
        loading: false 
      })
      throw err
    }
  },

  // 断开连接
  disconnectFromDatabase: async (id) => {
    try {
      set({ loading: true, error: null })
      await Disconnect(id)
      
      // 更新连接状态
      const status = await GetConnectionStatus(id)
      const { connections } = get()
      const conn = connections.find(c => c.id === id)
      if (conn) {
        conn.status = status.status
        conn.lastPing = status.lastPing
        conn.message = status.message
      }
      
      // 如果断开的是当前活动连接，清空活动连接
      const { activeConnection } = get()
      if (activeConnection?.id === id) {
        set({ activeConnection: null })
      }
      
      set({ loading: false })
    } catch (err) {
      set({ 
        error: err.message || '断开连接失败', 
        loading: false 
      })
      throw err
    }
  },

  // 设置活动连接
  setActiveConnection: (conn) => {
    set({ activeConnection: conn })
  },

  // 清除错误
  clearError: () => {
    set({ error: null })
  }
}))