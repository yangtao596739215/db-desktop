import React, { useState, useEffect } from 'react'
import { 
  Card, 
  Button, 
  Table, 
  Modal, 
  Form, 
  Input, 
  Select, 
  InputNumber, 
  Tag, 
  message, 
  Popconfirm,
  Space,
  Typography,
  Row,
  Col,
  Statistic
} from 'antd'
import { 
  PlusOutlined, 
  DatabaseOutlined, 
  EditOutlined, 
  DeleteOutlined,
  PlayCircleOutlined,
  StopOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined
} from '@ant-design/icons'
import { useConnectionStore } from '../stores/connection'
import dayjs from 'dayjs'

const { Title } = Typography
const { Option } = Select

function Connections({ type = null }) {
  const {
    connections,
    loading,
    error,
    loadConnections,
    addConnection,
    updateConnection,
    deleteConnection,
    testConnection,
    connectToDatabase,
    disconnectFromDatabase
  } = useConnectionStore()

  const [showAddDialog, setShowAddDialog] = useState(false)
  const [editingConnection, setEditingConnection] = useState(null)
  const [form] = Form.useForm()

  // 根据类型过滤连接
  const filteredConnections = type 
    ? connections.filter(conn => conn.type === type)
    : connections

  // 数据库类型默认端口
  const defaultPorts = {
    mysql: 3306,
    redis: 6379,
    clickhouse: 8123
  }

  // 获取连接类型颜色
  const getConnectionTypeColor = (type) => {
    const colors = {
      mysql: 'blue',
      redis: 'red',
      clickhouse: 'orange'
    }
    return colors[type] || 'default'
  }

  // 获取状态类型
  const getStatusType = (status) => {
    const types = {
      connected: 'success',
      error: 'error',
      disconnected: 'default'
    }
    return types[status] || 'default'
  }

  // 获取状态文本
  const getStatusText = (status) => {
    const texts = {
      connected: '已连接',
      error: '连接错误',
      disconnected: '未连接'
    }
    return texts[status] || '未知'
  }

  // 获取状态图标
  const getStatusIcon = (status) => {
    switch (status) {
      case 'connected':
        return <CheckCircleOutlined />
      case 'error':
        return <ExclamationCircleOutlined />
      default:
        return <InfoCircleOutlined />
    }
  }

  // 格式化时间
  const formatTime = (time) => {
    return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
  }

  // 数据库类型改变时的处理
  const onTypeChange = (selectedType) => {
    form.setFieldsValue({ port: defaultPorts[selectedType] })
    if (selectedType === 'redis') {
      form.setFieldsValue({ database: '' })
    }
  }

  // 重置表单
  const resetForm = () => {
    form.resetFields()
    setEditingConnection(null)
  }

  // 编辑连接
  const handleEdit = (conn) => {
    setEditingConnection(conn)
    form.setFieldsValue({
      ...conn,
      sslMode: conn.sslMode || ''
    })
    setShowAddDialog(true)
  }

  // 测试连接
  const handleTestConnection = async () => {
    try {
      const values = await form.validateFields()
      await testConnection(values)
      message.success('连接测试成功')
    } catch (error) {
      message.error(error.message || '连接测试失败')
    }
  }

  // 保存连接
  const handleSave = async () => {
    try {
      const values = await form.validateFields()
      const config = {
        ...values,
        id: editingConnection?.id || undefined
      }
      
      if (editingConnection) {
        await updateConnection(config)
        message.success('连接更新成功')
      } else {
        await addConnection(config)
        message.success('连接添加成功')
      }
      
      setShowAddDialog(false)
      resetForm()
    } catch (error) {
      message.error(error.message || '保存失败')
    }
  }

  // 切换连接状态
  const handleToggleConnection = async (conn) => {
    try {
      if (conn.status === 'connected') {
        await disconnectFromDatabase(conn.id)
        message.success('连接已断开')
      } else {
        await connectToDatabase(conn.id)
        message.success('连接成功')
      }
    } catch (error) {
      message.error(error.message || '操作失败')
    }
  }

  // 删除连接
  const handleDelete = async (id) => {
    try {
      await deleteConnection(id)
      message.success('连接已删除')
    } catch (error) {
      message.error(error.message || '删除失败')
    }
  }

  useEffect(() => {
    loadConnections()
  }, [loadConnections])

  const columns = [
    {
      title: '连接名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <Space>
          <DatabaseOutlined style={{ color: '#409eff' }} />
          <span style={{ fontWeight: 'bold' }}>{text}</span>
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type) => (
        <Tag color={getConnectionTypeColor(type)}>
          {type.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: '主机',
      dataIndex: 'host',
      key: 'host',
      render: (host, record) => (
        <span style={{ fontFamily: 'Monaco, Menlo, monospace' }}>
          {host}:{record.port}
        </span>
      ),
    },
    {
      title: '数据库',
      dataIndex: 'database',
      key: 'database',
      render: (database) => database || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status, record) => (
        <Space>
          <Tag 
            color={getStatusType(status)}
            icon={getStatusIcon(status)}
          >
            {getStatusText(status)}
          </Tag>
          {record.lastPing && (
            <span style={{ color: '#909399', fontSize: '12px' }}>
              最后检查: {formatTime(record.lastPing)}
            </span>
          )}
        </Space>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button
            type={record.status === 'connected' ? 'default' : 'primary'}
            size="small"
            icon={record.status === 'connected' ? <StopOutlined /> : <PlayCircleOutlined />}
            onClick={() => handleToggleConnection(record)}
          >
            {record.status === 'connected' ? '断开' : '连接'}
          </Button>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个连接吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const getPageTitle = () => {
    if (type) {
      const typeNames = {
        mysql: 'MySQL',
        redis: 'Redis',
        clickhouse: 'ClickHouse'
      }
      return `${typeNames[type]} 连接管理`
    }
    return '连接管理'
  }

  const getPageDescription = () => {
    if (type) {
      const descriptions = {
        mysql: '管理 MySQL 数据库连接',
        redis: '管理 Redis 缓存连接',
        clickhouse: '管理 ClickHouse 分析数据库连接'
      }
      return descriptions[type]
    }
    return '管理所有数据库连接'
  }

  return (
    <div className="connections" style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      <div style={{ marginBottom: '20px' }}>
        <Title level={2}>{getPageTitle()}</Title>
        <p style={{ color: '#666', marginBottom: '20px' }}>{getPageDescription()}</p>
        
        <Row gutter={16} style={{ marginBottom: '20px' }}>
          <Col span={6}>
            <Card>
              <Statistic
                title="总连接数"
                value={filteredConnections.length}
                prefix={<DatabaseOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="已连接"
                value={filteredConnections.filter(c => c.status === 'connected').length}
                valueStyle={{ color: '#3f8600' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                title="连接错误"
                value={filteredConnections.filter(c => c.status === 'error').length}
                valueStyle={{ color: '#cf1322' }}
                prefix={<ExclamationCircleOutlined />}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => setShowAddDialog(true)}
                style={{ width: '100%', height: '60px' }}
              >
                添加连接
              </Button>
            </Card>
          </Col>
        </Row>
      </div>

      <Card loading={loading}>
        {filteredConnections.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '60px 20px', color: '#909399' }}>
            <DatabaseOutlined style={{ fontSize: '64px', marginBottom: '20px' }} />
            <p style={{ fontSize: '16px', marginBottom: '20px' }}>暂无连接配置</p>
            <Button type="primary" onClick={() => setShowAddDialog(true)}>
              添加第一个连接
            </Button>
          </div>
        ) : (
          <Table
            columns={columns}
            dataSource={filteredConnections}
            rowKey="id"
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条记录`,
            }}
          />
        )}
      </Card>

      <Modal
        title={editingConnection ? '编辑连接' : '添加连接'}
        open={showAddDialog}
        onCancel={() => {
          setShowAddDialog(false)
          resetForm()
        }}
        onOk={handleSave}
        width={600}
        okText="保存"
        cancelText="取消"
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            type: type || 'mysql',
            host: 'localhost',
            port: defaultPorts[type || 'mysql'],
            timeout: 30,
            maxConns: 10
          }}
        >
          <Form.Item
            label="连接名称"
            name="name"
            rules={[{ required: true, message: '请输入连接名称' }]}
          >
            <Input placeholder="请输入连接名称" />
          </Form.Item>
          
          <Form.Item
            label="数据库类型"
            name="type"
            rules={[{ required: true, message: '请选择数据库类型' }]}
          >
            <Select placeholder="请选择数据库类型" onChange={onTypeChange} disabled={!!type}>
              <Option value="mysql">MySQL</Option>
              <Option value="redis">Redis</Option>
              <Option value="clickhouse">ClickHouse</Option>
            </Select>
          </Form.Item>
          
          <Form.Item
            label="主机地址"
            name="host"
            rules={[{ required: true, message: '请输入主机地址' }]}
          >
            <Input placeholder="请输入主机地址" />
          </Form.Item>
          
          <Form.Item
            label="端口"
            name="port"
            rules={[{ required: true, message: '请输入端口号' }]}
          >
            <InputNumber
              min={1}
              max={65535}
              placeholder="请输入端口号"
              style={{ width: '100%' }}
            />
          </Form.Item>
          
          <Form.Item
            label="用户名"
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>
          
          <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
          
          <Form.Item
            label="数据库名"
            name="database"
            rules={[{ required: true, message: '请输入数据库名' }]}
            hidden={form.getFieldValue('type') === 'redis'}
          >
            <Input placeholder="请输入数据库名" />
          </Form.Item>
          
          <Form.Item
            label="SSL模式"
            name="sslMode"
            hidden={!['mysql', 'clickhouse'].includes(form.getFieldValue('type'))}
          >
            <Select placeholder="请选择SSL模式">
              {form.getFieldValue('type') === 'mysql' ? (
                <>
                  <Option value="">禁用</Option>
                  <Option value="preferred">首选</Option>
                  <Option value="required">必需</Option>
                </>
              ) : (
                <>
                  <Option value="false">禁用</Option>
                  <Option value="true">启用</Option>
                </>
              )}
            </Select>
          </Form.Item>
          
          <Form.Item label="连接超时">
            <InputNumber
              name="timeout"
              min={1}
              max={300}
              placeholder="连接超时时间(秒)"
              style={{ width: '100%' }}
            />
          </Form.Item>
          
          <Form.Item label="最大连接数">
            <InputNumber
              name="maxConns"
              min={1}
              max={100}
              placeholder="最大连接数"
              style={{ width: '100%' }}
            />
          </Form.Item>
        </Form>
        
        <div style={{ textAlign: 'right', marginTop: '16px' }}>
          <Space>
            <Button onClick={handleTestConnection}>
              测试连接
            </Button>
          </Space>
        </div>
      </Modal>
    </div>
  )
}

export default Connections
