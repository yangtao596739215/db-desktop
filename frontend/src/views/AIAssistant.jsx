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
  Popconfirm,
  Input as AntInput,
  Drawer
} from 'antd'
import { 
  SendOutlined, 
  SettingOutlined, 
  RobotOutlined, 
  UserOutlined,
  ClearOutlined,
  CopyOutlined,
  PlusOutlined,
  DeleteOutlined,
  EditOutlined,
  MessageOutlined,
  HistoryOutlined,
  ToolOutlined
} from '@ant-design/icons'
import { useAIAssistantStore } from '../stores/aiAssistant'
import ToolConfirmationCard from '../components/ToolConfirmationCard'

const { TextArea } = Input
const { Title, Text } = Typography

const AIAssistant = () => {
  const [form] = Form.useForm()
  const [messageForm] = Form.useForm()
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [showConversationsDrawer, setShowConversationsDrawer] = useState(false)
  const [newConversationTitle, setNewConversationTitle] = useState('')
  const [editingConversationId, setEditingConversationId] = useState(null)
  const [editingTitle, setEditingTitle] = useState('')
  const messagesEndRef = useRef(null)
  
  const {
    messages,
    isLoading,
    config,
    pendingConfirmCards,
    toolConfirmationLoading,
    conversations,
    currentConversationId,
    conversationsLoading,
    addMessage,
    sendMessage,
    clearMessages,
    updateConfig,
    loadConfig,
    loadPendingCards,
    confirmToolCall,
    clearPendingCards,
    loadConversations,
    createConversation,
    selectConversation,
    deleteConversation,
    updateConversationTitle,
    sendMessageToCurrentConversation
  } = useAIAssistantStore()

  // 加载配置
  useEffect(() => {
    loadConfig()
  }, [loadConfig])

  // 加载会话列表
  useEffect(() => {
    loadConversations()
  }, [loadConversations])

  // 加载待确认的卡片
  useEffect(() => {
    loadPendingCards()
  }, [loadPendingCards])

  // 调试：监听pendingConfirmCards变化
  useEffect(() => {
    console.log('pendingConfirmCards updated:', pendingConfirmCards)
  }, [pendingConfirmCards])

  // 自动滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // 发送消息
  const handleSendMessage = async () => {
    try {
      const values = await messageForm.validateFields()
      if (!values.message.trim()) return

      const userMessage = values.message.trim()
      messageForm.setFieldsValue({ message: '' })
      
      // 如果没有当前会话，创建一个新会话
      if (!currentConversationId) {
        const conversation = await createConversation(userMessage.slice(0, 30) + (userMessage.length > 30 ? '...' : ''))
        setShowConversationsDrawer(false)
      }
      
      // 添加用户消息
      addMessage({
        role: 'user',
        content: userMessage,
        timestamp: new Date()
      })

      // 发送到AI
      if (currentConversationId) {
        await sendMessageToCurrentConversation(userMessage)
      } else {
        await sendMessage(userMessage)
      }
    } catch (error) {
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
      content: '确定要清空所有对话记录吗？',
      onOk: () => {
        clearMessages()
        clearPendingCards()
      }
    })
  }

  // 处理工具确认
  const handleToolApproval = async (toolCallID) => {
    try {
      await confirmToolCall(toolCallID, true)
      message.success('工具执行已确认')
    } catch (error) {
      message.error('确认工具执行失败')
    }
  }

  // 处理工具拒绝
  const handleToolRejection = async (toolCallID) => {
    try {
      await confirmToolCall(toolCallID, false)
      message.info('工具执行已拒绝')
    } catch (error) {
      message.error('拒绝工具执行失败')
    }
  }

  // 会话管理处理函数

  // 创建新会话
  const handleCreateConversation = async () => {
    if (!newConversationTitle.trim()) return
    
    try {
      await createConversation(newConversationTitle.trim())
      setNewConversationTitle('')
      setShowConversationsDrawer(false)
      message.success('会话创建成功')
    } catch (error) {
      message.error('创建会话失败')
    }
  }

  // 选择会话
  const handleSelectConversation = async (conversationId) => {
    try {
      await selectConversation(conversationId)
      setShowConversationsDrawer(false)
    } catch (error) {
      message.error('加载会话失败')
    }
  }

  // 删除会话
  const handleDeleteConversation = async (conversationId) => {
    try {
      await deleteConversation(conversationId)
      message.success('会话已删除')
    } catch (error) {
      message.error('删除会话失败')
    }
  }

  // 开始编辑会话标题
  const handleStartEditTitle = (conversation) => {
    setEditingConversationId(conversation.id)
    setEditingTitle(conversation.title)
  }

  // 保存编辑的标题
  const handleSaveEditTitle = async () => {
    if (!editingTitle.trim()) return
    
    try {
      await updateConversationTitle(editingConversationId, editingTitle.trim())
      setEditingConversationId(null)
      setEditingTitle('')
      message.success('标题已更新')
    } catch (error) {
      message.error('更新标题失败')
    }
  }

  // 取消编辑
  const handleCancelEdit = () => {
    setEditingConversationId(null)
    setEditingTitle('')
  }

  return (
    <div style={{ padding: '24px', height: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* 头部 */}
      <Card style={{ marginBottom: '16px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <RobotOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
            <Title level={3} style={{ margin: 0 }}>AI 助手</Title>
            {currentConversationId && (
              <Text type="secondary" style={{ marginLeft: '8px' }}>
                {conversations.find(c => c.id === currentConversationId)?.title || '当前会话'}
              </Text>
            )}
          </div>
          <Space>
            <Button 
              icon={<HistoryOutlined />} 
              onClick={() => setShowConversationsDrawer(true)}
            >
              历史会话
            </Button>
            <Button 
              icon={<PlusOutlined />} 
              onClick={handleCreateConversation}
              type="primary"
            >
              新建会话
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
            >
              清空对话
            </Button>
          </Space>
        </div>
      </Card>

      {/* 对话区域 */}
      <Card 
        style={{ 
          flex: 1, 
          display: 'flex', 
          flexDirection: 'column',
          overflow: 'hidden'
        }}
        bodyStyle={{ 
          padding: '16px', 
          height: '100%', 
          display: 'flex', 
          flexDirection: 'column' 
        }}
      >
        {/* 消息列表 */}
        <div 
          style={{ 
            flex: 1, 
            overflowY: 'auto', 
            padding: '16px 0',
            maxHeight: 'calc(100vh - 300px)'
          }}
        >
          {messages.length === 0 ? (
            <div style={{ 
              textAlign: 'center', 
              color: '#999', 
              padding: '40px 0' 
            }}>
              <RobotOutlined style={{ fontSize: '48px', marginBottom: '16px' }} />
              <div>开始与AI助手对话吧！</div>
            </div>
          ) : (
            messages.map((msg, index) => (
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
                    backgroundColor: msg.role === 'user' ? '#1890ff' : msg.role === 'tool' ? '#fa8c16' : '#52c41a',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'white',
                    flexShrink: 0
                  }}>
                    {msg.role === 'user' ? <UserOutlined /> : msg.role === 'tool' ? <ToolOutlined /> : <RobotOutlined />}
                  </div>
                  <div style={{
                    flex: 1,
                    backgroundColor: msg.role === 'user' ? '#f0f8ff' : msg.role === 'tool' ? '#fff7e6' : '#f6ffed',
                    padding: '12px 16px',
                    borderRadius: '8px',
                    border: '1px solid #e8e8e8'
                  }}>
                    <div style={{ whiteSpace: 'pre-wrap', marginBottom: '8px' }}>
                      {msg.role === 'tool' && (
                        <div style={{ 
                          fontSize: '12px', 
                          color: '#d46b08', 
                          marginBottom: '4px',
                          fontWeight: 'bold'
                        }}>
                          🔧 工具执行结果
                        </div>
                      )}
                      {msg.content}
                    </div>
                    <div style={{ 
                      fontSize: '12px', 
                      color: '#999',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center'
                    }}>
                      <span>{msg.timestamp.toLocaleTimeString()}</span>
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
            ))
          )}
          {/* 待确认的工具调用卡片 */}
          {pendingConfirmCards.length > 0 && (
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
              {/* 显示卡片确认 */}
              {pendingConfirmCards.map((card) => (
                <ToolConfirmationCard
                  key={card.cardId}
                  toolCall={{
                    toolCallID: card.toolCallId,
                    function: 'unknown', // 卡片中没有函数名
                    arguments: {}, // 卡片中没有参数
                    message: card.showContent
                  }}
                  onApprove={handleToolApproval}
                  onReject={handleToolRejection}
                  isLoading={toolConfirmationLoading}
                />
              ))}
            </div>
          )}

          {isLoading && (
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
          <div ref={messagesEndRef} />
        </div>

        <Divider />

        {/* 输入区域 */}
        <Form form={messageForm} onFinish={handleSendMessage}>
          <div style={{ display: 'flex', gap: '8px' }}>
            <Form.Item 
              name="message" 
              style={{ flex: 1, margin: 0 }}
              rules={[{ required: true, message: '请输入消息' }]}
            >
              <TextArea
                placeholder="输入你的问题..."
                autoSize={{ minRows: 1, maxRows: 4 }}
                onPressEnter={(e) => {
                  if (e.shiftKey) return
                  e.preventDefault()
                  handleSendMessage()
                }}
                disabled={isLoading}
              />
            </Form.Item>
            <Button 
              type="primary" 
              icon={<SendOutlined />} 
              htmlType="submit"
              loading={isLoading}
              disabled={isLoading}
            >
              发送
            </Button>
          </div>
        </Form>
      </Card>

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
          
          <Form.Item
            label="启用流式响应"
            name="stream"
            valuePropName="checked"
            initialValue={true}
          >
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      {/* 会话管理抽屉 */}
      <Drawer
        title="历史会话"
        placement="left"
        width={400}
        open={showConversationsDrawer}
        onClose={() => setShowConversationsDrawer(false)}
        extra={
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={handleCreateConversation}
          >
            新建会话
          </Button>
        }
      >
        {/* 新建会话输入框 */}
        <div style={{ marginBottom: '16px' }}>
          <Space.Compact style={{ width: '100%' }}>
            <AntInput
              placeholder="输入会话标题"
              value={newConversationTitle}
              onChange={(e) => setNewConversationTitle(e.target.value)}
              onPressEnter={handleCreateConversation}
            />
            <Button 
              type="primary" 
              onClick={handleCreateConversation}
              disabled={!newConversationTitle.trim()}
            >
              创建
            </Button>
          </Space.Compact>
        </div>

        {/* 会话列表 */}
        <List
          loading={conversationsLoading}
          dataSource={conversations}
          renderItem={(conversation) => (
            <List.Item
              key={conversation.id}
              style={{
                backgroundColor: conversation.id === currentConversationId ? '#f0f8ff' : 'transparent',
                borderRadius: '6px',
                padding: '8px',
                marginBottom: '4px'
              }}
              actions={[
                <Button
                  key="edit"
                  type="text"
                  icon={<EditOutlined />}
                  onClick={() => handleStartEditTitle(conversation)}
                />,
                <Popconfirm
                  key="delete"
                  title="确定要删除这个会话吗？"
                  onConfirm={() => handleDeleteConversation(conversation.id)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button
                    type="text"
                    danger
                    icon={<DeleteOutlined />}
                  />
                </Popconfirm>
              ]}
            >
              <List.Item.Meta
                avatar={<MessageOutlined />}
                title={
                  editingConversationId === conversation.id ? (
                    <Space.Compact style={{ width: '100%' }}>
                      <AntInput
                        value={editingTitle}
                        onChange={(e) => setEditingTitle(e.target.value)}
                        onPressEnter={handleSaveEditTitle}
                        size="small"
                      />
                      <Button 
                        type="primary" 
                        size="small"
                        onClick={handleSaveEditTitle}
                      >
                        保存
                      </Button>
                      <Button 
                        size="small"
                        onClick={handleCancelEdit}
                      >
                        取消
                      </Button>
                    </Space.Compact>
                  ) : (
                    <div 
                      style={{ cursor: 'pointer' }}
                      onClick={() => handleSelectConversation(conversation.id)}
                    >
                      {conversation.title}
                    </div>
                  )
                }
                description={
                  <div>
                    <div style={{ fontSize: '12px', color: '#999' }}>
                      {conversation.messageCount} 条消息
                    </div>
                    <div style={{ fontSize: '12px', color: '#999' }}>
                      {new Date(conversation.updatedAt).toLocaleString()}
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      </Drawer>

      <style jsx>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  )
}

export default AIAssistant
