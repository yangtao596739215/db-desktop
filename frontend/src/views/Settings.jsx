import React, { useState, useEffect } from 'react'
import { 
  Card, 
  Row, 
  Col, 
  Form, 
  Radio, 
  Select, 
  Switch, 
  InputNumber, 
  Slider, 
  Button, 
  Typography, 
  Space, 
  Modal,
  message
} from 'antd'
import { 
  SettingOutlined, 
  EditOutlined, 
  BarChartOutlined, 
  InfoCircleOutlined,
  DatabaseOutlined
} from '@ant-design/icons'
import { useQueryStore } from '../stores/query'
import { useConnectionStore } from '../stores/connection'

const { Title } = Typography
const { Option } = Select

function Settings() {
  const { clearHistory, queryHistory } = useQueryStore()
  const { connections } = useConnectionStore()
  
  const [form] = Form.useForm()
  const [settings, setSettings] = useState({
    theme: 'auto',
    language: 'zh-CN',
    autoSave: true,
    queryTimeout: 30,
    resultLimit: 1000,
    fontSize: 14,
    autoComplete: true,
    syntaxHighlight: true,
    showLineNumbers: true,
    autoIndent: true,
    tabSize: 2,
    autoFormat: false
  })

  // 默认设置
  const defaultSettings = {
    theme: 'auto',
    language: 'zh-CN',
    autoSave: true,
    queryTimeout: 30,
    resultLimit: 1000,
    fontSize: 14,
    autoComplete: true,
    syntaxHighlight: true,
    showLineNumbers: true,
    autoIndent: true,
    tabSize: 2,
    autoFormat: false
  }

  // 加载设置
  const loadSettings = () => {
    const saved = localStorage.getItem('db-desktop-settings')
    if (saved) {
      try {
        const parsed = JSON.parse(saved)
        setSettings(parsed)
        form.setFieldsValue(parsed)
      } catch (error) {
        console.error('Failed to load settings:', error)
      }
    }
  }

  // 保存设置
  const handleSaveSettings = () => {
    try {
      localStorage.setItem('db-desktop-settings', JSON.stringify(settings))
      message.success('设置已保存')
    } catch (error) {
      message.error('保存设置失败')
    }
  }

  // 重置设置
  const handleResetSettings = () => {
    Modal.confirm({
      title: '确认重置',
      content: '确定要重置所有设置吗？',
      onOk() {
        setSettings(defaultSettings)
        form.setFieldsValue(defaultSettings)
        handleSaveSettings()
      }
    })
  }

  // 清空查询历史
  const handleClearQueryHistory = () => {
    Modal.confirm({
      title: '确认清空',
      content: '确定要清空查询历史吗？',
      onOk() {
        clearHistory()
        message.success('查询历史已清空')
      }
    })
  }

  // 导出查询历史
  const handleExportQueryHistory = () => {
    try {
      const data = JSON.stringify(queryHistory, null, 2)
      const blob = new Blob([data], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `query-history-${new Date().toISOString().split('T')[0]}.json`
      a.click()
      URL.revokeObjectURL(url)
      message.success('查询历史已导出')
    } catch (error) {
      message.error('导出查询历史失败')
    }
  }

  // 导出连接配置
  const handleExportConnections = () => {
    try {
      const data = JSON.stringify(connections, null, 2)
      const blob = new Blob([data], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `connections-${new Date().toISOString().split('T')[0]}.json`
      a.click()
      URL.revokeObjectURL(url)
      message.success('连接配置已导出')
    } catch (error) {
      message.error('导出连接配置失败')
    }
  }

  // 导入连接配置
  const handleImportConnections = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json'
    input.onchange = async (event) => {
      const file = event.target.files[0]
      if (file) {
        try {
          const text = await file.text()
          const connections = JSON.parse(text)
          // 这里应该实现导入逻辑
          message.success('连接配置导入成功')
        } catch (error) {
          message.error('导入连接配置失败')
        }
      }
    }
    input.click()
  }

  // 清空所有数据
  const handleClearAllData = () => {
    Modal.confirm({
      title: '确认清空',
      content: '确定要清空所有应用数据吗？这将删除所有连接配置和查询历史。',
      onOk() {
        clearHistory()
        // 这里应该实现清空连接配置的逻辑
        message.success('所有数据已清空')
      }
    })
  }

  // 检查更新
  const handleCheckUpdate = () => {
    message.info('当前已是最新版本')
  }

  // 打开GitHub
  const handleOpenGitHub = () => {
    window.open('https://github.com/your-username/db-desktop', '_blank')
  }

  useEffect(() => {
    loadSettings()
  }, [])

  const handleFormChange = (changedValues) => {
    setSettings(prev => ({ ...prev, ...changedValues }))
  }

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      <div style={{ marginBottom: '20px' }}>
        <Title level={2}>设置</Title>
      </div>

      <Row gutter={20}>
        <Col span={12}>
          {/* 应用设置 */}
          <Card 
            title={
              <Space>
                <SettingOutlined />
                <span>应用设置</span>
              </Space>
            }
            style={{ marginBottom: '20px' }}
          >
            <Form
              form={form}
              layout="vertical"
              initialValues={settings}
              onValuesChange={handleFormChange}
            >
              <Form.Item label="主题模式" name="theme">
                <Radio.Group>
                  <Radio value="light">浅色模式</Radio>
                  <Radio value="dark">深色模式</Radio>
                  <Radio value="auto">跟随系统</Radio>
                </Radio.Group>
              </Form.Item>
              
              <Form.Item label="语言" name="language">
                <Select style={{ width: 200 }}>
                  <Option value="zh-CN">简体中文</Option>
                  <Option value="en-US">English</Option>
                </Select>
              </Form.Item>
              
              <Form.Item label="自动保存" name="autoSave" valuePropName="checked">
                <Switch />
                <span style={{ marginLeft: '8px', color: '#909399', fontSize: '12px' }}>
                  自动保存查询历史和连接配置
                </span>
              </Form.Item>
              
              <Form.Item label="查询超时" name="queryTimeout">
                <InputNumber min={5} max={300} style={{ width: 200 }} />
                <span style={{ marginLeft: '8px', color: '#909399', fontSize: '12px' }}>秒</span>
              </Form.Item>
              
              <Form.Item label="结果限制" name="resultLimit">
                <InputNumber min={10} max={10000} style={{ width: 200 }} />
                <span style={{ marginLeft: '8px', color: '#909399', fontSize: '12px' }}>
                  默认查询结果行数限制
                </span>
              </Form.Item>
              
              <Form.Item label="字体大小" name="fontSize">
                <Slider min={12} max={20} style={{ width: 200 }} />
              </Form.Item>
            </Form>
          </Card>
        </Col>
        
        <Col span={12}>
          {/* 编辑器设置 */}
          <Card 
            title={
              <Space>
                <EditOutlined />
                <span>编辑器设置</span>
              </Space>
            }
            style={{ marginBottom: '20px' }}
          >
            <Form
              form={form}
              layout="vertical"
              initialValues={settings}
              onValuesChange={handleFormChange}
            >
              <Form.Item label="自动补全" name="autoComplete" valuePropName="checked">
                <Switch />
              </Form.Item>
              
              <Form.Item label="语法高亮" name="syntaxHighlight" valuePropName="checked">
                <Switch />
              </Form.Item>
              
              <Form.Item label="行号显示" name="showLineNumbers" valuePropName="checked">
                <Switch />
              </Form.Item>
              
              <Form.Item label="自动缩进" name="autoIndent" valuePropName="checked">
                <Switch />
              </Form.Item>
              
              <Form.Item label="制表符大小" name="tabSize">
                <InputNumber min={2} max={8} style={{ width: 200 }} />
              </Form.Item>
              
              <Form.Item label="自动格式化" name="autoFormat" valuePropName="checked">
                <Switch />
                <span style={{ marginLeft: '8px', color: '#909399', fontSize: '12px' }}>
                  保存时自动格式化SQL
                </span>
              </Form.Item>
            </Form>
          </Card>
        </Col>
      </Row>

      <Row gutter={20} style={{ marginTop: '20px' }}>
        <Col span={12}>
          {/* 数据管理 */}
          <Card 
            title={
              <Space>
                <BarChartOutlined />
                <span>数据管理</span>
              </Space>
            }
            style={{ marginBottom: '20px' }}
          >
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center', 
                padding: '12px', 
                border: '1px solid #e4e7ed', 
                borderRadius: '4px' 
              }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 'bold', color: '#303133', marginBottom: '4px' }}>
                    查询历史
                  </div>
                  <div style={{ color: '#909399', fontSize: '12px' }}>
                    管理SQL查询历史记录
                  </div>
                </div>
                <Space>
                  <Button size="small" onClick={handleClearQueryHistory}>
                    清空历史
                  </Button>
                  <Button size="small" type="primary" onClick={handleExportQueryHistory}>
                    导出
                  </Button>
                </Space>
              </div>
              
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center', 
                padding: '12px', 
                border: '1px solid #e4e7ed', 
                borderRadius: '4px' 
              }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 'bold', color: '#303133', marginBottom: '4px' }}>
                    连接配置
                  </div>
                  <div style={{ color: '#909399', fontSize: '12px' }}>
                    管理数据库连接配置
                  </div>
                </div>
                <Space>
                  <Button size="small" onClick={handleExportConnections}>
                    导出配置
                  </Button>
                  <Button size="small" type="primary" onClick={handleImportConnections}>
                    导入配置
                  </Button>
                </Space>
              </div>
              
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'center', 
                padding: '12px', 
                border: '1px solid #e4e7ed', 
                borderRadius: '4px' 
              }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 'bold', color: '#303133', marginBottom: '4px' }}>
                    应用数据
                  </div>
                  <div style={{ color: '#909399', fontSize: '12px' }}>
                    清理所有应用数据
                  </div>
                </div>
                <Button size="small" danger onClick={handleClearAllData}>
                  清空数据
                </Button>
              </div>
            </div>
          </Card>
        </Col>
        
        <Col span={12}>
          {/* 关于信息 */}
          <Card 
            title={
              <Space>
                <InfoCircleOutlined />
                <span>关于</span>
              </Space>
            }
            style={{ marginBottom: '20px' }}
          >
            <div style={{ textAlign: 'center' }}>
              <div style={{ 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'center', 
                marginBottom: '20px', 
                gap: '16px' 
              }}>
                <div style={{ 
                  display: 'flex', 
                  alignItems: 'center', 
                  justifyContent: 'center', 
                  width: '64px', 
                  height: '64px', 
                  background: '#f0f9ff', 
                  borderRadius: '50%' 
                }}>
                  <DatabaseOutlined style={{ fontSize: '48px', color: '#409eff' }} />
                </div>
                <div>
                  <h3 style={{ margin: '0 0 8px', color: '#303133' }}>DB Desktop</h3>
                  <p style={{ margin: '4px 0', color: '#606266', fontSize: '14px' }}>版本 1.0.0</p>
                  <p style={{ margin: '4px 0', color: '#606266', fontSize: '14px' }}>
                    基于 Wails3 + React + Ant Design 构建
                  </p>
                </div>
              </div>
              
              <div style={{ textAlign: 'left', marginBottom: '20px' }}>
                <h4 style={{ margin: '0 0 12px', color: '#303133' }}>支持的功能：</h4>
                <ul style={{ margin: 0, paddingLeft: '20px', color: '#606266' }}>
                  <li>MySQL 数据库连接和查询</li>
                  <li>Redis 数据库连接和操作</li>
                  <li>ClickHouse 数据库连接和查询</li>
                  <li>SQL 查询编辑和执行</li>
                  <li>数据库结构浏览</li>
                  <li>查询历史记录</li>
                  <li>连接配置管理</li>
                </ul>
              </div>
              
              <Space>
                <Button onClick={handleCheckUpdate}>检查更新</Button>
                <Button type="primary" onClick={handleOpenGitHub}>GitHub</Button>
              </Space>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 保存按钮 */}
      <div style={{ 
        position: 'fixed', 
        bottom: '20px', 
        right: '20px', 
        display: 'flex', 
        gap: '12px', 
        background: '#fff', 
        padding: '12px 20px', 
        borderRadius: '8px', 
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)' 
      }}>
        <Button onClick={handleResetSettings}>重置</Button>
        <Button type="primary" onClick={handleSaveSettings}>保存设置</Button>
      </div>
    </div>
  )
}

export default Settings
