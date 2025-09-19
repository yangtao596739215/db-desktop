import { create } from 'zustand'
import { 
  ExecuteQuery,
  ExecuteQueryWithLimit,
  GetDatabases,
  GetTables,
  GetTableInfo,
  GetTableData,
  GetDatabaseInfo,
  FormatQuery,
  ValidateQuery
} from '../wailsjs/go/main/App'

export const useQueryStore = create((set, get) => ({
  // 状态 - 按数据库类型分组
  queryStates: {
    mysql: {
      queryHistory: [],
      currentQuery: '',
      queryResult: null,
      databases: [],
      tables: [],
      currentTable: null,
      loading: false,
      error: null,
      activeConnection: null,
    },
    clickhouse: {
      queryHistory: [],
      currentQuery: '',
      queryResult: null,
      databases: [],
      tables: [],
      currentTable: null,
      loading: false,
      error: null,
      activeConnection: null,
    },
    redis: {
      queryHistory: [],
      currentQuery: '',
      queryResult: null,
      databases: [],
      tables: [],
      currentTable: null,
      loading: false,
      error: null,
      activeConnection: null,
    }
  },
  currentType: 'mysql', // 当前活跃的数据库类型

  // 辅助函数
  getCurrentState: () => {
    const { queryStates, currentType } = get()
    return queryStates[currentType] || queryStates.mysql
  },

  updateCurrentState: (updates) => {
    const { queryStates, currentType } = get()
    set({
      queryStates: {
        ...queryStates,
        [currentType]: {
          ...queryStates[currentType],
          ...updates
        }
      }
    })
  },

  // 计算属性
  recentQueries: () => {
    const currentState = get().getCurrentState()
    return currentState.queryHistory.slice(-10).reverse()
  },

  // 获取当前状态属性 - 使用函数形式
  getQueryHistory: () => get().getCurrentState().queryHistory,
  getCurrentQuery: () => get().getCurrentState().currentQuery,
  getQueryResult: () => get().getCurrentState().queryResult,
  getDatabases: () => get().getCurrentState().databases,
  getTables: () => get().getCurrentState().tables,
  getCurrentTable: () => get().getCurrentState().currentTable,
  getLoading: () => get().getCurrentState().loading,
  getError: () => get().getCurrentState().error,
  getActiveConnection: () => get().getCurrentState().activeConnection,

  // 设置当前类型
  setCurrentType: (type) => {
    set({ currentType: type })
  },

  // 执行查询
  executeQuery: async (connectionId, query, limit = null) => {
    console.log('executeQuery called with:', { connectionId, query, limit })
    
    if (!connectionId || !query.trim()) {
      console.error('Missing connectionId or query:', { connectionId, query })
      throw new Error('连接ID和查询语句不能为空')
    }

    try {
      get().updateCurrentState({ loading: true, error: null })
      
      console.log('Calling backend ExecuteQuery...')
      let result
      if (limit) {
        result = await ExecuteQueryWithLimit(connectionId, query, limit)
      } else {
        result = await ExecuteQuery(connectionId, query)
      }
      
      console.log('Backend ExecuteQuery result:', result)
      
      // 添加到历史记录
      get().addToHistory(query, result)
      
      get().updateCurrentState({ queryResult: result, loading: false })
      return result
    } catch (err) {
      console.error('executeQuery error:', err)
      get().updateCurrentState({ 
        error: err.message || '查询执行失败', 
        loading: false 
      })
      throw err
    }
  },

  // 获取数据库列表
  loadDatabases: async (connectionId) => {
    try {
      get().updateCurrentState({ loading: true, error: null })
      const result = await GetDatabases(connectionId)
      get().updateCurrentState({ databases: result || [], loading: false })
      return result
    } catch (err) {
      get().updateCurrentState({ 
        error: err.message || '获取数据库列表失败', 
        loading: false 
      })
      throw err
    }
  },

  // 获取表列表
  loadTables: async (connectionId, database) => {
    console.log('loadTables 调用:', { connectionId, database })
    try {
      get().updateCurrentState({ loading: true, error: null })
      const result = await GetTables(connectionId, database)
      console.log('GetTables 后端返回结果:', result)
      get().updateCurrentState({ tables: result || [], loading: false })
      return result
    } catch (err) {
      console.error('loadTables 错误:', err)
      get().updateCurrentState({ 
        error: err.message || '获取表列表失败', 
        loading: false 
      })
      throw err
    }
  },

  // 获取表信息
  loadTableInfo: async (connectionId, database, table) => {
    try {
      get().updateCurrentState({ loading: true, error: null })
      const result = await GetTableInfo(connectionId, database, table)
      get().updateCurrentState({ currentTable: result, loading: false })
      return result
    } catch (err) {
      get().updateCurrentState({ 
        error: err.message || '获取表信息失败', 
        loading: false 
      })
      throw err
    }
  },

  // 获取表数据
  loadTableData: async (connectionId, database, table, limit = 100, offset = 0) => {
    try {
      get().updateCurrentState({ loading: true, error: null })
      const result = await GetTableData(connectionId, database, table, limit, offset)
      get().updateCurrentState({ queryResult: result, loading: false })
      return result
    } catch (err) {
      get().updateCurrentState({ 
        error: err.message || '获取表数据失败', 
        loading: false 
      })
      throw err
    }
  },

  // 获取数据库信息
  loadDatabaseInfo: async (connectionId) => {
    try {
      get().updateCurrentState({ loading: true, error: null })
      const result = await GetDatabaseInfo(connectionId)
      get().updateCurrentState({ loading: false })
      return result
    } catch (err) {
      get().updateCurrentState({ 
        error: err.message || '获取数据库信息失败', 
        loading: false 
      })
      throw err
    }
  },

  // 格式化查询
  formatQuery: async (connectionId, query) => {
    try {
      const result = await FormatQuery(connectionId, query)
      return result
    } catch (err) {
      console.error('格式化查询失败:', err)
      return query
    }
  },

  // 验证查询
  validateQuery: async (connectionId, query) => {
    try {
      await ValidateQuery(connectionId, query)
      return true
    } catch (err) {
      get().updateCurrentState({ error: err.message || '查询验证失败' })
      throw err
    }
  },

  // 添加到历史记录
  addToHistory: (query, result) => {
    const historyItem = {
      id: Date.now(),
      query: query.trim(),
      timestamp: new Date(),
      success: !result.error,
      executionTime: result.time,
      rowCount: result.count
    }
    
    const currentState = get().getCurrentState()
    const newHistory = [...currentState.queryHistory, historyItem]
    
    // 限制历史记录数量
    if (newHistory.length > 100) {
      newHistory.splice(0, newHistory.length - 100)
    }
    
    get().updateCurrentState({ queryHistory: newHistory })
  },

  // 清空查询结果
  clearResult: () => {
    get().updateCurrentState({ queryResult: null })
  },

  // 清空历史记录
  clearHistory: () => {
    get().updateCurrentState({ queryHistory: [] })
  },

  // 设置当前查询
  setCurrentQuery: (query) => {
    get().updateCurrentState({ currentQuery: query })
  },

  // 设置活动连接
  setActiveConnection: (connection) => {
    get().updateCurrentState({ activeConnection: connection })
  },

  // 清除错误
  clearError: () => {
    get().updateCurrentState({ error: null })
  },

  // 重置所有状态
  resetState: () => {
    const { queryStates, currentType } = get()
    set({
      queryStates: {
        ...queryStates,
        [currentType]: {
          queryHistory: [],
          currentQuery: '',
          queryResult: null,
          databases: [],
          tables: [],
          currentTable: null,
          loading: false,
          error: null,
          activeConnection: null
        }
      }
    })
  },

  // 重置查询相关状态（保留历史记录）
  resetQueryState: () => {
    get().updateCurrentState({
      currentQuery: '',
      queryResult: null,
      databases: [],
      tables: [],
      currentTable: null,
      loading: false,
      error: null,
      activeConnection: null
    })
  }
}))