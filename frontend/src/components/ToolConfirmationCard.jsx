import React from 'react'
import { Card, Button, Space, Typography, Tag, Divider } from 'antd'
import { 
  CheckCircleOutlined, 
  CloseCircleOutlined, 
  ToolOutlined,
  DatabaseOutlined,
  CodeOutlined
} from '@ant-design/icons'

const { Text, Title } = Typography

const ToolConfirmationCard = ({ 
  toolCall, 
  onApprove, 
  onReject, 
  isLoading = false 
}) => {
  const getToolIcon = (functionName) => {
    switch (functionName) {
      case 'execute_redis_command':
        return <DatabaseOutlined style={{ color: '#ff4d4f' }} />
      case 'execute_mysql_query':
        return <DatabaseOutlined style={{ color: '#1890ff' }} />
      case 'execute_clickhouse_query':
        return <DatabaseOutlined style={{ color: '#52c41a' }} />
      default:
        return <ToolOutlined style={{ color: '#722ed1' }} />
    }
  }

  const getToolName = (functionName) => {
    switch (functionName) {
      case 'execute_redis_command':
        return 'Redis命令'
      case 'execute_mysql_query':
        return 'MySQL查询'
      case 'execute_clickhouse_query':
        return 'ClickHouse查询'
      default:
        return functionName
    }
  }

  const getToolColor = (functionName) => {
    switch (functionName) {
      case 'execute_redis_command':
        return 'red'
      case 'execute_mysql_query':
        return 'blue'
      case 'execute_clickhouse_query':
        return 'green'
      default:
        return 'purple'
    }
  }

  const formatArguments = (args) => {
    if (!args) return {}
    
    // 格式化参数显示
    const formatted = {}
    Object.keys(args).forEach(key => {
      const value = args[key]
      if (typeof value === 'string' && value.length > 100) {
        formatted[key] = value.substring(0, 100) + '...'
      } else {
        formatted[key] = value
      }
    })
    return formatted
  }

  return (
    <Card
      size="small"
      style={{
        marginBottom: '12px',
        border: '1px solid #d9d9d9',
        borderRadius: '8px',
        backgroundColor: '#fafafa'
      }}
      bodyStyle={{ padding: '12px' }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: '12px' }}>
        {/* 工具图标 */}
        <div style={{
          width: '32px',
          height: '32px',
          borderRadius: '50%',
          backgroundColor: '#fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          border: '1px solid #d9d9d9',
          flexShrink: 0
        }}>
          {getToolIcon(toolCall.function)}
        </div>

        {/* 内容区域 */}
        <div style={{ flex: 1, minWidth: 0 }}>
          {/* 标题 */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
            <Title level={5} style={{ margin: 0 }}>
              {toolCall.message}
            </Title>
            <Tag color={getToolColor(toolCall.function)}>
              {getToolName(toolCall.function)}
            </Tag>
          </div>

          {/* 参数详情 */}
          <div style={{ marginBottom: '12px' }}>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              工具ID: <Text code>{toolCall.toolCallID}</Text>
            </Text>
            <br />
            <Text type="secondary" style={{ fontSize: '12px' }}>
              函数: <Text code>{toolCall.function}</Text>
            </Text>
          </div>

          {/* 参数显示 */}
          {toolCall.arguments && Object.keys(toolCall.arguments).length > 0 && (
            <div style={{ marginBottom: '12px' }}>
              <Text strong style={{ fontSize: '12px' }}>参数:</Text>
              <div style={{ 
                marginTop: '4px',
                padding: '8px',
                backgroundColor: '#fff',
                border: '1px solid #e8e8e8',
                borderRadius: '4px',
                fontFamily: 'monospace',
                fontSize: '12px'
              }}>
                {Object.entries(formatArguments(toolCall.arguments)).map(([key, value]) => (
                  <div key={key} style={{ marginBottom: '2px' }}>
                    <Text strong>{key}:</Text> <Text code>{JSON.stringify(value)}</Text>
                  </div>
                ))}
              </div>
            </div>
          )}

          <Divider style={{ margin: '8px 0' }} />

          {/* 操作按钮 */}
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button
              type="primary"
              danger
              size="small"
              icon={<CloseCircleOutlined />}
              onClick={() => onReject(toolCall.toolCallID)}
              loading={isLoading}
              disabled={isLoading}
            >
              拒绝
            </Button>
            <Button
              type="primary"
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={() => onApprove(toolCall.toolCallID)}
              loading={isLoading}
              disabled={isLoading}
            >
              确认执行
            </Button>
          </div>
        </div>
      </div>
    </Card>
  )
}

export default ToolConfirmationCard
