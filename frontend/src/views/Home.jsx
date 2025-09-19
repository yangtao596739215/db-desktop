import React, { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { 
  Card, 
  Button, 
  Row, 
  Col, 
  Statistic, 
  Steps, 
  Tag, 
  Typography,
  Space
} from 'antd'
import { 
  DatabaseOutlined, 
  LinkOutlined, 
  FileTextOutlined,
  ClockCircleOutlined,
  QuestionCircleOutlined,
  PlayCircleOutlined
} from '@ant-design/icons'
import { useConnectionStore } from '../stores/connection'
import { useQueryStore } from '../stores/query'

const { Title, Paragraph } = Typography
const { Step } = Steps

function Home() {
  const navigate = useNavigate()
  const { 
    connections, 
    connectedConnections, 
    loadConnections, 
    connectToDatabase 
  } = useConnectionStore()
  const { queryHistory } = useQueryStore()

  // 最近连接（取前5个）
  const recentConnections = connections.slice(0, 5)

  // 获取连接类型颜色
  const getConnectionTypeColor = (type) => {
    const colors = {
      mysql: 'blue',
      redis: 'red',
      clickhouse: 'orange'
    }
    return colors[type] || 'default'
  }

  // 连接到数据库
  const handleConnectToDatabase = async (conn) => {
    try {
      if (conn.status === 'connected') {
        // 如果已连接，直接跳转到查询页面
        navigate('/query')
      } else {
        // 如果未连接，先连接再跳转
        await connectToDatabase(conn.id)
        navigate('/query')
      }
    } catch (error) {
      console.error('连接失败:', error)
    }
  }

  useEffect(() => {
    // 加载连接列表
    loadConnections()
  }, [loadConnections])

  return (
    <div className="home" style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
      <Card style={{ marginBottom: '20px', textAlign: 'center' }}>
        <div style={{ padding: '40px 20px' }}>
          <DatabaseOutlined style={{ fontSize: '64px', color: '#409EFF', marginBottom: '20px' }} />
          <Title level={1} style={{ margin: '20px 0 10px', color: '#303133' }}>
            欢迎使用 DB Desktop
          </Title>
          <Paragraph style={{ color: '#606266', fontSize: '16px', margin: 0 }}>
            一个现代化的数据库管理工具，支持 MySQL、Redis 和 ClickHouse
          </Paragraph>
        </div>
      </Card>

      <Row gutter={20} style={{ marginBottom: '20px' }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="已配置连接"
              value={connections.length}
              prefix={<LinkOutlined style={{ color: '#67C23A' }} />}
              valueStyle={{ fontSize: '24px', fontWeight: 'bold' }}
            />
          </Card>
        </Col>
        
        <Col span={8}>
          <Card>
            <Statistic
              title="活跃连接"
              value={connectedConnections().length}
              prefix={<LinkOutlined style={{ color: '#E6A23C' }} />}
              valueStyle={{ fontSize: '24px', fontWeight: 'bold' }}
            />
          </Card>
        </Col>
        
        <Col span={8}>
          <Card>
            <Statistic
              title="查询历史"
              value={queryHistory.length}
              prefix={<FileTextOutlined style={{ color: '#F56C6C' }} />}
              valueStyle={{ fontSize: '24px', fontWeight: 'bold' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={20} style={{ marginBottom: '20px' }}>
        <Col span={12}>
          <Card 
            title={
              <Space>
                <LinkOutlined />
                <span>快速连接</span>
              </Space>
            }
            style={{ height: '120px' }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', height: '60px' }}>
              <p style={{ margin: 0, color: '#606266' }}>配置和管理数据库连接</p>
              <Button type="primary" onClick={() => navigate('/connections')}>
                管理连接
              </Button>
            </div>
          </Card>
        </Col>
        
        <Col span={12}>
          <Card 
            title={
              <Space>
                <PlayCircleOutlined />
                <span>SQL 查询</span>
              </Space>
            }
            style={{ height: '120px' }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', height: '60px' }}>
              <p style={{ margin: 0, color: '#606266' }}>执行 SQL 查询和浏览数据</p>
              <Button type="primary" onClick={() => navigate('/query')}>
                开始查询
              </Button>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 最近连接 */}
      {recentConnections.length > 0 && (
        <Card 
          title={
            <Space>
              <ClockCircleOutlined />
              <span>最近连接</span>
            </Space>
          }
          style={{ marginBottom: '20px' }}
        >
          <div style={{ maxHeight: '200px', overflowY: 'auto' }}>
            {recentConnections.map(conn => (
              <div
                key={conn.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '12px 0',
                  borderBottom: '1px solid #f0f0f0',
                  cursor: 'pointer',
                  transition: 'background-color 0.3s'
                }}
                onClick={() => handleConnectToDatabase(conn)}
                onMouseEnter={(e) => e.target.style.backgroundColor = '#f5f7fa'}
                onMouseLeave={(e) => e.target.style.backgroundColor = 'transparent'}
              >
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 'bold', color: '#303133', marginBottom: '4px' }}>
                    {conn.name}
                  </div>
                  <Space>
                    <Tag color={getConnectionTypeColor(conn.type)}>
                      {conn.type.toUpperCase()}
                    </Tag>
                    <span style={{ color: '#909399', fontSize: '12px' }}>
                      {conn.host}:{conn.port}
                    </span>
                  </Space>
                </div>
                <div style={{ marginLeft: '10px' }}>
                  <Tag color={conn.status === 'connected' ? 'success' : 'default'}>
                    {conn.status === 'connected' ? '已连接' : '未连接'}
                  </Tag>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* 快速开始指南 */}
      <Card 
        title={
          <Space>
            <QuestionCircleOutlined />
            <span>快速开始</span>
          </Space>
        }
        style={{ marginBottom: '20px' }}
      >
        <Steps current={0} size="small">
          <Step title="添加连接" description="配置数据库连接信息" />
          <Step title="测试连接" description="验证连接是否正常" />
          <Step title="开始查询" description="执行 SQL 查询和浏览数据" />
        </Steps>
      </Card>
    </div>
  )
}

export default Home
