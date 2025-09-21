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
  InputNumber
} from 'antd'
import { 
  SendOutlined, 
  SettingOutlined, 
  RobotOutlined, 
  UserOutlined,
  ClearOutlined,
  CopyOutlined
} from '@ant-design/icons'
import { useAIAssistantStore } from '../stores/aiAssistant'

const { TextArea } = Input
const { Title, Text } = Typography

const AIAssistantSimple = () => {
  const [form] = Form.useForm()
  const [messageForm] = Form.useForm()
  const [showConfigModal, setShowConfigModal] = useState(false)
  const messagesEndRef = useRef(null)
  
  const {
    messages,
    isLoading,
    config,
    addMessage,
    sendMessage,
    clearMessages,
    updateConfig,
    loadConfig
  } = useAIAssistantStore()

  // 加载配置
  useEffect(() => {
    loadConfig()
  }, [loadConfig])

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
      
      // 添加用户消息
      addMessage({
        role: 'user',
        content: userMessage,
        timestamp: new Date()
      })

      // 发送到AI
      await sendMessage(userMessage)
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
      onOk: clearMessages
    })
  }

  return (
    <div style={{ padding: '24px', height: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* 头部 */}
      <Card style={{ marginBottom: '16px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <RobotOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
            <Title level={3} style={{ margin: 0 }}>AI 助手 (简化版)</Title>
          </div>
          <Space>
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

      <style jsx>{`
        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  )
}

export default AIAssistantSimple
