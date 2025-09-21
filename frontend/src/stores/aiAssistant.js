import { create } from 'zustand'
import { 
  SendMessage,
  GetAIConfig,
  UpdateAIConfig,
  CreateConversation,
  ListConversations,
  GetConversationHistory,
  DeleteConversation,
  UpdateConversation,
  ConfirmToolCall
} from '../wailsjs/go/app/App'

export const useAIAssistantStore = create((set, get) => ({
  // 状态
  messages: [],
  isLoading: false,
  error: null,
  config: {
    apiKey: '',
    temperature: 0.7,
    stream: true
  },
  // 工具确认相关状态
  pendingConfirmCards: [],
  toolConfirmationLoading: false,
  // 会话管理状态
  conversations: [],
  currentConversationId: null,
  conversationsLoading: false,

  // 添加消息
  addMessage: (message) => {
    set(state => ({
      messages: [...state.messages, message]
    }))
  },

  // 发送消息（使用普通响应）
  sendMessage: async (content) => {
    try {
      set({ isLoading: true, error: null })
      
      // 确保有会话ID，如果没有则先创建一个
      let { currentConversationId } = get()
      if (!currentConversationId) {
        const conversation = await CreateConversation(content.substring(0, 20))
        currentConversationId = conversation.id
        set({ currentConversationId })
      }
      
      // 使用普通方法发送消息
      const response = await SendMessage(currentConversationId, content)
      
      // 添加AI回复到消息列表
      set(state => ({
        messages: [...state.messages, {
          role: 'assistant',
          content: response.content || '抱歉，没有收到回复',
          timestamp: new Date()
        }]
      }))
      
      set({ isLoading: false })
    } catch (err) {
      console.error('sendMessage error:', err)
      set({ 
        error: err.message || '发送消息失败',
        isLoading: false 
      })
      
      // 添加错误消息
      set(state => ({
        messages: [...state.messages, {
          role: 'assistant',
          content: `抱歉，发生了错误：${err.message || '发送消息失败'}`,
          timestamp: new Date()
        }]
      }))
    }
  },

  // 清空消息
  clearMessages: () => {
    set({ messages: [] })
  },

  // 加载配置
  loadConfig: async () => {
    try {
      const config = await GetAIConfig()
      console.log('Loaded AI config:', config)
      set({ config })
    } catch (err) {
      console.error('loadConfig error:', err)
      // 设置默认配置
      const defaultConfig = {
        apiKey: '',
        temperature: 0.7,
        stream: true
      }
      set({ config: defaultConfig })
    }
  },

  // 更新配置
  updateConfig: async (newConfig) => {
    try {
      await UpdateAIConfig(newConfig)
      set({ config: newConfig })
    } catch (err) {
      console.error('updateConfig error:', err)
      throw err
    }
  },


  // 确认工具调用（通过卡片）
  confirmToolCall: async (toolCallID, approved) => {
    try {
      set({ toolConfirmationLoading: true })
      
      // 找到对应的卡片
      const { pendingConfirmCards } = get()
      const card = pendingConfirmCards.find(card => card.toolCallId === toolCallID)
      
      if (!card) {
        console.error('No confirmation card found for tool call:', toolCallID)
        console.error('Available cards:', pendingConfirmCards)
        throw new Error('No confirmation card found for tool call')
      }
      
      // 调用后端的工具确认方法
      await ConfirmToolCall(card.cardId, approved)
      
      // 从待确认列表中移除
      set(state => ({
        pendingConfirmCards: state.pendingConfirmCards.filter(card => card.toolCallId !== toolCallID)
      }))
      
      set({ toolConfirmationLoading: false })
      return { success: true }
    } catch (err) {
      console.error('confirmToolCall error:', err)
      set({ toolConfirmationLoading: false })
      throw err
    }
  },


  // 会话管理方法

  // 加载会话列表
  loadConversations: async () => {
    try {
      set({ conversationsLoading: true })
      const conversations = await ListConversations()
      set({ conversations, conversationsLoading: false })
    } catch (err) {
      console.error('loadConversations error:', err)
      set({ conversationsLoading: false })
    }
  },

  // 创建新会话
  createConversation: async (title) => {
    try {
      const conversation = await CreateConversation(title)
      set(state => ({
        conversations: [conversation, ...state.conversations],
        currentConversationId: conversation.id
      }))
      return conversation
    } catch (err) {
      console.error('createConversation error:', err)
      throw err
    }
  },

  // 选择会话
  selectConversation: async (conversationId) => {
    try {
      const messages = await GetConversationHistory(conversationId)
      set({
        currentConversationId: conversationId,
        messages: messages.map(msg => ({
          role: msg.role,
          content: msg.content,
          timestamp: new Date(msg.createdAt),
          isToolResult: msg.isToolResult
        }))
      })
    } catch (err) {
      console.error('selectConversation error:', err)
      throw err
    }
  },

  // 删除会话
  deleteConversation: async (conversationId) => {
    try {
      await DeleteConversation(conversationId)
      set(state => {
        const newConversations = state.conversations.filter(c => c.id !== conversationId)
        const newCurrentId = state.currentConversationId === conversationId ? null : state.currentConversationId
        const newMessages = state.currentConversationId === conversationId ? [] : state.messages
        return {
          conversations: newConversations,
          currentConversationId: newCurrentId,
          messages: newMessages
        }
      })
    } catch (err) {
      console.error('deleteConversation error:', err)
      throw err
    }
  },

  // 更新会话标题
  updateConversationTitle: async (conversationId, newTitle) => {
    try {
      const conversation = get().conversations.find(c => c.id === conversationId)
      if (conversation) {
        conversation.title = newTitle
        await UpdateConversation(conversation)
        set(state => ({
          conversations: state.conversations.map(c => 
            c.id === conversationId ? conversation : c
          )
        }))
      }
    } catch (err) {
      console.error('updateConversationTitle error:', err)
      throw err
    }
  },

  // 发送消息到当前会话
  sendMessageToCurrentConversation: async (content) => {
    const { currentConversationId } = get()
    if (!currentConversationId) {
      throw new Error('No conversation selected')
    }
    return get().sendMessageToConversation(currentConversationId, content)
  },

  // 发送消息到指定会话（使用普通响应）
  sendMessageToConversation: async (conversationId, content) => {
    try {
      set({ isLoading: true, error: null })
      
      // 使用普通方法发送消息
      const response = await SendMessage(conversationId, content)
      
      // 添加AI回复到消息列表
      set(state => ({
        messages: [...state.messages, {
          role: 'assistant',
          content: response.content || '抱歉，没有收到回复',
          timestamp: new Date()
        }]
      }))
      
      set({ isLoading: false })

      // 更新会话列表中的消息数量
      set(state => ({
        conversations: state.conversations.map(c => 
          c.id === conversationId 
            ? { ...c, messageCount: c.messageCount + 2, updatedAt: new Date().toISOString() }
            : c
        )
      }))
    } catch (err) {
      console.error('sendMessageToConversation error:', err)
      set({ 
        error: err.message || '发送消息失败',
        isLoading: false 
      })
      
      // 添加错误消息
      set(state => ({
        messages: [...state.messages, {
          role: 'assistant',
          content: `抱歉，发生了错误：${err.message || '发送消息失败'}`,
          timestamp: new Date()
        }]
      }))
    }
  }
}))

