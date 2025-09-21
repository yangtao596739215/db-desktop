import React, { useState, useRef, useEffect } from 'react'
import { 
  Card, 
  Input, 
  Button, 
  Form, 
  message, 
  Space, 
  Typography, 
  Divider,
  Modal,
  Switch,
  InputNumber,
  List,
  Row,
  Col,
  Tooltip,
  Popconfirm,
  Empty,
  Badge
} from 'antd'
import { 
  SendOutlined, 
  SettingOutlined, 
  RobotOutlined, 
  UserOutlined,
  ClearOutlined,
  CopyOutlined,
  PlusOutlined,
  MessageOutlined,
  DeleteOutlined,
  EditOutlined,
  ClockCircleOutlined
} from '@ant-design/icons'
import { useAIAssistantStore } from '../stores/aiAssistant'

const { TextArea } = Input
const { Title, Text } = Typography

// 解析卡片消息内容，提取工具调用信息
const parseCardMessage = (content) => {
  // 简单的解析逻辑，可以根据实际需要调整
  // 这里假设卡片消息格式为 "执行Redis命令: `command`" 等
  const redisMatch = content.match(/执行Redis命令: `(.+)`/)
  const mysqlMatch = content.match(/执行MySQL查询: `(.+)`/)
  const clickhouseMatch = content.match(/执行ClickHouse查询: `(.+)`/)
  
  if (redisMatch) {
    return {
      toolCallID: 'redis_' + Date.now(),
      function: 'execute_redis_command',
      arguments: { command: redisMatch[1] }
    }
  } else if (mysqlMatch) {
    return {
      toolCallID: 'mysql_' + Date.now(),
      function: 'execute_mysql_query',
      arguments: { query: mysqlMatch[1] }
    }
  } else if (clickhouseMatch) {
    return {
      toolCallID: 'clickhouse_' + Date.now(),
      function: 'execute_clickhouse_query',
      arguments: { query: clickhouseMatch[1] }
    }
  }
  
  return {
    toolCallID: 'unknown',
    function: 'unknown',
    arguments: {}
  }
}

const AIAssistantFixed = () => {
  const [form] = Form.useForm()
  const [messageForm] = Form.useForm()
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [showConversations, setShowConversations] = useState(true)
  const [editingConversationId, setEditingConversationId] = useState(null)
  const [editingTitle, setEditingTitle] = useState('')
  const messagesEndRef = useRef(null)
  
  const {
    messages,
    isLoading,
    config,
    pendingConfirmCards,
    toolConfirmationLoading,
    currentStreamingMessage,
    isStreaming,
    conversations,
    currentConversationId,
    conversationsLoading,
    addMessage,
    sendMessage,
    clearMessages,
    updateConfig,
    loadConfig,
    confirmToolCall,
    initializeEventListeners,
    cleanupEventListeners,
    loadConversations,
    createConversation,
    selectConversation,
    deleteConversation,
    updateConversationTitle
  } = useAIAssistantStore()

  // 加载配置和初始化事件监听器
  useEffect(() => {
    loadConfig()
    initializeEventListeners()
    loadConversations()
    
    // 清理函数
    return () => {
      cleanupEventListeners()
    }
  }, [loadConfig, initializeEventListeners, cleanupEventListeners, loadConversations])

  // 自动滚动到底部
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ 
        behavior: 'smooth', 
        block: 'end' 
      })
    }
  }, [messages, currentStreamingMessage])

  // 发送消息
  const handleSendMessage = async () => {
    try {
      const values = await messageForm.validateFields()
      if (!values.message.trim()) return

      const userMessage = values.message.trim()
      messageForm.setFieldsValue({ message: '' })
      
      // 添加用户消息
      addMessage({
        role: 'user',
        content: userMessage,
        timestamp: new Date()
      })

      // 发送到AI
      await sendMessage(userMessage)
    } catch (error) {
      console.error('Send message error:', error)
      message.error('发送消息失败')
    }
  }

  // 保存配置
  const handleSaveConfig = async () => {
    try {
      const values = await form.validateFields()
      await updateConfig(values)
      setShowConfigModal(false)
      message.success('配置保存成功')
    } catch (error) {
      console.error('Save config error:', error)
      message.error('保存配置失败')
    }
  }

  // 复制消息
  const handleCopyMessage = (content) => {
    navigator.clipboard.writeText(content)
    message.success('已复制到剪贴板')
  }

  // 清空对话
  const handleClearMessages = () => {
    Modal.confirm({
      title: '确认清空',
      content: '确定要清空当前会话的所有对话记录吗？',
      onOk: () => {
        clearMessages()
      }
    })
  }

  // 创建新会话
  const handleCreateConversation = async () => {
    try {
      const title = `新对话 ${new Date().toLocaleString()}`
      const conversation = await createConversation(title)
      message.success('新会话创建成功')
      await selectConversation(conversation.id)
    } catch (error) {
      console.error('Create conversation error:', error)
      message.error('创建会话失败')
    }
  }

  // 选择会话
  const handleSelectConversation = async (conversationId) => {
    try {
      await selectConversation(conversationId)
    } catch (error) {
      console.error('Select conversation error:', error)
      message.error('加载会话失败')
    }
  }

  // 删除会话
  const handleDeleteConversation = async (conversationId) => {
    try {
      await deleteConversation(conversationId)
      message.success('会话删除成功')
    } catch (error) {
      console.error('Delete conversation error:', error)
      message.error('删除会话失败')
    }
  }

  // 编辑会话标题
  const handleEditTitle = (conversationId, currentTitle) => {
    setEditingConversationId(conversationId)
    setEditingTitle(currentTitle)
  }

  // 保存会话标题
  const handleSaveTitle = async (conversationId) => {
    try {
      await updateConversationTitle(conversationId, editingTitle)
      setEditingConversationId(null)
      setEditingTitle('')
      message.success('标题更新成功')
    } catch (error) {
      console.error('Update title error:', error)
      message.error('更新标题失败')
    }
  }

  // 取消编辑标题
  const handleCancelEdit = () => {
    setEditingConversationId(null)
    setEditingTitle('')
  }

  // 处理工具确认
  const handleToolApproval = async (toolCallID) => {
    try {
      await confirmToolCall(toolCallID, true)
      message.success('工具执行已确认')
    } catch (error) {
      console.error('Tool approval error:', error)
      message.error('确认工具执行失败')
    }
  }

  // 处理工具拒绝
  const handleToolRejection = async (toolCallID) => {
    try {
      await confirmToolCall(toolCallID, false)
      message.info('工具执行已拒绝')
    } catch (error) {
      console.error('Tool rejection error:', error)
      message.error('拒绝工具执行失败')
    }
  }

  // 简化的工具确认卡片组件
  const SimpleToolConfirmationCard = ({ toolCall }) => (
    <Card
      size="small"
      style={{
        marginBottom: '12px',
        border: '1px solid #d9d9d9',
        borderRadius: '8px',
        backgroundColor: '#fafafa'
      }}
    >
      <div style={{ padding: '12px' }}>
        <div style={{ marginBottom: '8px' }}>
          <Text strong>{toolCall.message}</Text>
        </div>
        <div style={{ marginBottom: '8px' }}>
          <Text type="secondary">工具: {toolCall.function}</Text>
        </div>
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
          <Button
            type="primary"
            danger
            size="small"
            onClick={() => handleToolRejection(toolCall.toolCallID)}
            loading={toolConfirmationLoading}
            disabled={toolConfirmationLoading}
          >
            拒绝
          </Button>
          <Button
            type="primary"
            size="small"
            onClick={() => handleToolApproval(toolCall.toolCallID)}
            loading={toolConfirmationLoading}
            disabled={toolConfirmationLoading}
          >
            确认执行
          </Button>
        </div>
      </div>
    </Card>
  )

  return (
    <div style={{ 
      height: '100vh', 
      display: 'flex', 
      flexDirection: 'column',
      padding: '24px'
    }}>
      {/* 头部 */}
      <Card style={{ marginBottom: '16px', flexShrink: 0 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <RobotOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
            <Title level={3} style={{ margin: 0 }}>AI 助手</Title>
            {currentConversationId && (
              <Badge 
                count={conversations.find(c => c.id === currentConversationId)?.messageCount || 0}
                style={{ backgroundColor: '#52c41a' }}
              />
            )}
          </div>
          <Space>
            <Button 
              icon={<MessageOutlined />} 
              onClick={() => setShowConversations(!showConversations)}
              type={showConversations ? 'primary' : 'default'}
            >
              {showConversations ? '隐藏会话' : '显示会话'}
            </Button>
            <Button 
              icon={<SettingOutlined />} 
              onClick={() => setShowConfigModal(true)}
            >
              配置
            </Button>
            <Button 
              icon={<ClearOutlined />} 
              onClick={handleClearMessages}
              danger
              disabled={!currentConversationId}
            >
              清空对话
            </Button>
          </Space>
        </div>
      </Card>

      {/* 主内容区域 */}
      <div style={{ 
        flex: 1, 
        display: 'flex', 
        gap: '16px',
        minHeight: 0
      }}>
        {/* 会话列表 */}
        {showConversations && (
          <div style={{ 
            width: '300px', 
            flexShrink: 0,
            display: 'flex',
            flexDirection: 'column'
          }}>
            <Card 
              title={
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span>会话列表</span>
                  <Button 
                    type="primary" 
                    size="small"
                    icon={<PlusOutlined />}
                    onClick={handleCreateConversation}
                    loading={conversationsLoading}
                  >
                    新建
                  </Button>
                </div>
              }
              style={{ 
                height: '100%',
                display: 'flex',
                flexDirection: 'column'
              }}
              bodyStyle={{ 
                padding: '0',
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden'
              }}
            >
              <div 
                className="conversation-list-container"
                style={{ 
                  flex: 1,
                  overflowY: 'auto',
                  padding: '12px'
                }}
              >
              {conversationsLoading ? (
                <div style={{ textAlign: 'center', padding: '20px' }}>
                  <div>加载中...</div>
                </div>
              ) : conversations.length === 0 ? (
                <Empty 
                  description="暂无会话" 
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  style={{ marginTop: '40px' }}
                >
                  <Button type="primary" onClick={handleCreateConversation}>
                    创建第一个会话
                  </Button>
                </Empty>
              ) : (
                <List
                  dataSource={conversations}
                  renderItem={(conversation) => (
                    <List.Item
                      style={{
                        padding: '8px 12px',
                        marginBottom: '8px',
                        borderRadius: '6px',
                        cursor: 'pointer',
                        backgroundColor: currentConversationId === conversation.id ? '#e6f7ff' : 'transparent',
                        border: currentConversationId === conversation.id ? '1px solid #1890ff' : '1px solid #f0f0f0',
                        transition: 'all 0.3s'
                      }}
                      onClick={() => handleSelectConversation(conversation.id)}
                      onMouseEnter={(e) => {
                        if (currentConversationId !== conversation.id) {
                          e.target.style.backgroundColor = '#f5f5f5'
                        }
                      }}
                      onMouseLeave={(e) => {
                        if (currentConversationId !== conversation.id) {
                          e.target.style.backgroundColor = 'transparent'
                        }
                      }}
                    >
                      <div style={{ width: '100%' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                          <div style={{ flex: 1, minWidth: 0 }}>
                            {editingConversationId === conversation.id ? (
                              <Input
                                size="small"
                                value={editingTitle}
                                onChange={(e) => setEditingTitle(e.target.value)}
                                onPressEnter={() => handleSaveTitle(conversation.id)}
                                onBlur={() => handleSaveTitle(conversation.id)}
                                autoFocus
                                style={{ marginBottom: '4px' }}
                              />
                            ) : (
                              <div
                                style={{
                                  fontWeight: currentConversationId === conversation.id ? 'bold' : 'normal',
                                  fontSize: '14px',
                                  marginBottom: '4px',
                                  overflow: 'hidden',
                                  textOverflow: 'ellipsis',
                                  whiteSpace: 'nowrap'
                                }}
                                title={conversation.title}
                              >
                                {conversation.title}
                              </div>
                            )}
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                              <Text type="secondary" style={{ fontSize: '12px' }}>
                                <ClockCircleOutlined style={{ marginRight: '4px' }} />
                                {new Date(conversation.updatedAt).toLocaleString()}
                              </Text>
                              <Badge 
                                count={conversation.messageCount} 
                                size="small"
                                style={{ backgroundColor: '#f0f0f0', color: '#666' }}
                              />
                            </div>
                          </div>
                          <div style={{ display: 'flex', gap: '4px', marginLeft: '8px' }}>
                            <Tooltip title="编辑标题">
                              <Button
                                type="text"
                                size="small"
                                icon={<EditOutlined />}
                                onClick={(e) => {
                                  e.stopPropagation()
                                  handleEditTitle(conversation.id, conversation.title)
                                }}
                                style={{ width: '24px', height: '24px', padding: 0 }}
                              />
                            </Tooltip>
                            <Popconfirm
                              title="确定删除这个会话吗？"
                              description="删除后无法恢复"
                              onConfirm={(e) => {
                                e.stopPropagation()
                                handleDeleteConversation(conversation.id)
                              }}
                              okText="删除"
                              cancelText="取消"
                              okType="danger"
                            >
                              <Button
                                type="text"
                                size="small"
                                danger
                                icon={<DeleteOutlined />}
                                onClick={(e) => e.stopPropagation()}
                                style={{ width: '24px', height: '24px', padding: 0 }}
                              />
                            </Popconfirm>
                          </div>
                        </div>
                      </div>
                    </List.Item>
                  )}
                />
              )}
              </div>
            </Card>
          </div>
        )}

        {/* 对话区域 */}
        <div style={{ 
          flex: 1,
          display: 'flex',
          flexDirection: 'column',
          minWidth: 0
        }}>
          <Card 
            style={{ 
              height: '100%',
              display: 'flex', 
              flexDirection: 'column'
            }}
            bodyStyle={{ 
              padding: 0,
              flex: 1, 
              display: 'flex', 
              flexDirection: 'column',
              overflow: 'hidden'
            }}
          >
            {!currentConversationId ? (
              <div style={{ 
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                height: '100%',
                textAlign: 'center',
                color: '#999',
                padding: '40px'
              }}>
                <MessageOutlined style={{ fontSize: '64px', marginBottom: '16px', color: '#d9d9d9' }} />
                <Title level={4} style={{ color: '#999', marginBottom: '8px' }}>
                  选择一个会话开始对话
                </Title>
                <Text type="secondary">
                  从左侧选择一个现有会话，或创建新会话开始与AI助手对话
                </Text>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />}
                  onClick={handleCreateConversation}
                  style={{ marginTop: '16px' }}
                >
                  创建新会话
                </Button>
              </div>
            ) : (
              <>
                {/* 消息列表 - 正常顺序显示，最新消息在底部，支持滑动 */}
                <div 
                  className="message-container"
                  style={{ 
                    flex: 1, 
                    overflowY: 'auto', 
                    padding: '16px',
                    minHeight: 0,
                    display: 'flex',
                    flexDirection: 'column'
                  }}
                >
                  {/* 显示消息 - 如果没有消息则显示欢迎界面 */}
                  {messages.length === 0 && !isLoading && !isStreaming && (!pendingConfirmCards || pendingConfirmCards.length === 0) ? (
                    <div style={{ 
                      textAlign: 'center', 
                      color: '#999', 
                      padding: '40px 0'
                    }}>
                      <RobotOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
                      <div>开始与AI助手对话吧！</div>
                    </div>
                  ) : (
                    <>
                      {/* 历史消息 - 正序显示 */}
                      {messages.map((msg, index) => {
                        // 如果是卡片消息，显示确认卡片
                        if (msg.role === 'card') {
                          const cardInfo = parseCardMessage(msg.content)
                          return (
                            <div key={index} style={{ marginBottom: '16px' }}>
                              <SimpleToolConfirmationCard
                                toolCall={{
                                  toolCallID: cardInfo.toolCallID || 'unknown',
                                  function: cardInfo.function || 'unknown',
                                  arguments: cardInfo.arguments || {},
                                  message: msg.content
                                }}
                              />
                            </div>
                          )
                        }
                        
                        // 普通消息
                        return (
                          <div key={index} style={{ marginBottom: '16px' }}>
                            <div style={{ 
                              display: 'flex', 
                              alignItems: 'flex-start', 
                              gap: '12px',
                              flexDirection: msg.role === 'user' ? 'row-reverse' : 'row'
                            }}>
                              <div style={{
                                width: '32px',
                                height: '32px',
                                borderRadius: '50%',
                                backgroundColor: msg.role === 'user' ? '#1890ff' : '#52c41a',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                color: 'white',
                                flexShrink: 0
                              }}>
                                {msg.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
                              </div>
                              <div style={{
                                flex: 1,
                                backgroundColor: msg.role === 'user' ? '#f0f8ff' : '#f6ffed',
                                padding: '12px 16px',
                                borderRadius: '8px',
                                border: '1px solid #e8e8e8'
                              }}>
                                <div style={{ whiteSpace: 'pre-wrap', marginBottom: '8px' }}>
                                  {msg.content}
                                </div>
                                <div style={{ 
                                  fontSize: '12px', 
                                  color: '#999',
                                  display: 'flex',
                                  justifyContent: 'space-between',
                                  alignItems: 'center'
                                }}>
                                  <span>{msg.timestamp?.toLocaleTimeString()}</span>
                                  {msg.role === 'assistant' && (
                                    <Button 
                                      type="text" 
                                      size="small" 
                                      icon={<CopyOutlined />}
                                      onClick={() => handleCopyMessage(msg.content)}
                                    />
                                  )}
                                </div>
                              </div>
                            </div>
                          </div>
                        )
                      })}

                      {/* 待确认的工具调用卡片 - 显示在消息后面 */}
                      {pendingConfirmCards && pendingConfirmCards.length > 0 && (
                        <div style={{ marginBottom: '16px' }}>
                          <div style={{ 
                            marginBottom: '8px',
                            padding: '8px 12px',
                            backgroundColor: '#fff7e6',
                            border: '1px solid #ffd591',
                            borderRadius: '6px',
                            fontSize: '14px',
                            color: '#d46b08'
                          }}>
                            <strong>⚠️ 需要确认的工具执行</strong>
                          </div>
                          {pendingConfirmCards.map((card) => (
                            <SimpleToolConfirmationCard
                              key={card.cardId}
                              toolCall={{
                                toolCallID: card.toolCallId,
                                function: 'unknown',
                                arguments: {},
                                message: card.showContent
                              }}
                            />
                          ))}
                        </div>
                      )}

                      {/* 加载状态 - 显示在所有内容后面 */}
                      {isLoading && !isStreaming && (
                        <div style={{ 
                          display: 'flex', 
                          alignItems: 'center', 
                          gap: '12px',
                          marginBottom: '16px'
                        }}>
                          <div style={{
                            width: '32px',
                            height: '32px',
                            borderRadius: '50%',
                            backgroundColor: '#52c41a',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: 'white'
                          }}>
                            <RobotOutlined />
                          </div>
                          <div style={{
                            backgroundColor: '#f6ffed',
                            padding: '12px 16px',
                            borderRadius: '8px',
                            border: '1px solid #e8e8e8'
                          }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                              <div style={{ 
                                width: '16px', 
                                height: '16px', 
                                border: '2px solid #52c41a',
                                borderTop: '2px solid transparent',
                                borderRadius: '50%',
                                animation: 'spin 1s linear infinite'
                              }} />
                              <span>AI正在思考中...</span>
                            </div>
                          </div>
                        </div>
                      )}

                      {/* 流式消息 - 显示在最后 */}
                      {isStreaming && currentStreamingMessage && (
                        <div style={{ 
                          display: 'flex', 
                          alignItems: 'flex-start', 
                          gap: '12px',
                          marginBottom: '16px'
                        }}>
                          <div style={{
                            width: '32px',
                            height: '32px',
                            borderRadius: '50%',
                            backgroundColor: '#52c41a',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: 'white',
                            flexShrink: 0
                          }}>
                            <RobotOutlined />
                          </div>
                          <div style={{
                            flex: 1,
                            backgroundColor: '#f6ffed',
                            padding: '12px 16px',
                            borderRadius: '8px',
                            border: '1px solid #e8e8e8'
                          }}>
                            <div style={{ whiteSpace: 'pre-wrap', marginBottom: '8px' }}>
                              {currentStreamingMessage}
                              <span style={{ 
                                display: 'inline-block',
                                width: '8px',
                                height: '16px',
                                backgroundColor: '#52c41a',
                                marginLeft: '2px',
                                animation: 'blink 1s infinite'
                              }} />
                            </div>
                            <div style={{ 
                              fontSize: '12px', 
                              color: '#999',
                              display: 'flex',
                              alignItems: 'center',
                              gap: '8px'
                            }}>
                              <div style={{ 
                                width: '12px', 
                                height: '12px', 
                                border: '2px solid #52c41a',
                                borderTop: '2px solid transparent',
                                borderRadius: '50%',
                                animation: 'spin 1s linear infinite'
                              }} />
                              <span>正在接收消息...</span>
                            </div>
                          </div>
                        </div>
                      )}
                    </>
                  )}
                  
                  {/* 滚动到底部的占位元素 */}
                  <div ref={messagesEndRef} style={{ height: '1px' }} />
                </div>

                {/* 输入区域 - 固定在底部 */}
                <div style={{ 
                  padding: '16px',
                  borderTop: '1px solid #f0f0f0',
                  backgroundColor: '#fafafa',
                  flexShrink: 0,
                  position: 'sticky',
                  bottom: 0,
                  zIndex: 1
                }}>
                  <Form form={messageForm} onFinish={handleSendMessage}>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <Form.Item 
                        name="message" 
                        style={{ flex: 1, margin: 0 }}
                        rules={[{ required: true, message: '请输入消息' }]}
                      >
                        <TextArea
                          placeholder="输入你的问题... (Ctrl+Enter发送)"
                          autoSize={{ minRows: 1, maxRows: 4 }}
                          onPressEnter={(e) => {
                            if (e.ctrlKey) {
                              e.preventDefault()
                              handleSendMessage()
                            }
                          }}
                          disabled={isLoading || !currentConversationId}
                        />
                      </Form.Item>
                      <Button 
                        type="primary" 
                        icon={<SendOutlined />} 
                        htmlType="submit"
                        loading={isLoading}
                        disabled={isLoading || !currentConversationId}
                      >
                        发送
                      </Button>
                    </div>
                  </Form>
                </div>
              </>
            )}
          </Card>
        </div>
      </div>

      {/* 配置模态框 */}
      <Modal
        title="AI 助手配置"
        open={showConfigModal}
        onCancel={() => setShowConfigModal(false)}
        onOk={handleSaveConfig}
        width={600}
        okText="保存"
        cancelText="取消"
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={config}
        >
          <Form.Item
            label="千问API Key"
            name="apiKey"
            rules={[{ required: true, message: '请输入千问API Key' }]}
          >
            <Input.Password placeholder="请输入千问API Key" />
          </Form.Item>
          
          
          <Form.Item
            label="温度"
            name="temperature"
            initialValue={0.7}
          >
            <InputNumber 
              min={0} 
              max={2} 
              step={0.1}
              style={{ width: '100%' }}
              placeholder="控制回复的随机性"
            />
          </Form.Item>
          
        </Form>
      </Modal>

      <style jsx>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
        
        @keyframes blink {
          0%, 50% { opacity: 1; }
          51%, 100% { opacity: 0; }
        }
        
        /* 优化滚动条样式 */
        .message-container::-webkit-scrollbar,
        .conversation-list-container::-webkit-scrollbar {
          width: 8px;
        }
        
        .message-container::-webkit-scrollbar-track,
        .conversation-list-container::-webkit-scrollbar-track {
          background: #f1f1f1;
          border-radius: 4px;
        }
        
        .message-container::-webkit-scrollbar-thumb,
        .conversation-list-container::-webkit-scrollbar-thumb {
          background: #c1c1c1;
          border-radius: 4px;
        }
        
        .message-container::-webkit-scrollbar-thumb:hover,
        .conversation-list-container::-webkit-scrollbar-thumb:hover {
          background: #a8a8a8;
        }
        
        /* 平滑滚动 */
        .message-container,
        .conversation-list-container {
          scroll-behavior: smooth;
        }
      `}</style>
    </div>
  )
}

export default AIAssistantFixed
