import React, { useState, useEffect } from 'react'
import { 
  Table, 
  Card, 
  Button, 
  Select, 
  Input, 
  Space, 
  Tag, 
  Typography, 
  message, 
  Popconfirm,
  Statistic,
  Row,
  Col,
  Tooltip,
  Modal,
  Alert,
  Empty
} from 'antd'
import { 
  ReloadOutlined, 
  DeleteOutlined, 
  SearchOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DatabaseOutlined
} from '@ant-design/icons'
import { 
  GetQueryHistory,
  GetQueryHistoryByDBType,
  GetQueryHistoryStats,
  ClearQueryHistory,
  RetryQuery
} from '../wailsjs/go/app/App'

const { Title } = Typography
const { Option } = Select
const { Search } = Input

function QueryHistory() {
  const [queryHistory, setQueryHistory] = useState([])
  const [loading, setLoading] = useState(false)
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0
  })
  const [dbTypeFilter, setDbTypeFilter] = useState('all')
  const [searchText, setSearchText] = useState('')
  const [stats, setStats] = useState({})
  const [retryLoading, setRetryLoading] = useState({})
  const [retryResult, setRetryResult] = useState(null)
  const [showResultModal, setShowResultModal] = useState(false)

  // 加载查询历史
  const loadQueryHistory = async (page = 1, pageSize = 20, dbType = 'all') => {
    setLoading(true)
    try {
      let history
      if (dbType === 'all') {
        history = await GetQueryHistory(pageSize, (page - 1) * pageSize)
      } else {
        history = await GetQueryHistoryByDBType(dbType, pageSize, (page - 1) * pageSize)
      }
      
      setQueryHistory(history || [])
      setPagination(prev => ({
        ...prev,
        current: page,
        pageSize: pageSize,
        total: (history || []).length === pageSize ? page * pageSize + 1 : (history || []).length
      }))
    } catch (error) {
      console.error('Failed to load query history:', error)
      message.error('加载查询历史失败')
    } finally {
      setLoading(false)
    }
  }

  // 加载统计信息
  const loadStats = async () => {
    try {
      const statsData = await GetQueryHistoryStats()
      setStats(statsData || {})
    } catch (error) {
      console.error('Failed to load stats:', error)
    }
  }

  // 清空查询历史
  const clearHistory = async () => {
    try {
      await ClearQueryHistory()
      message.success('查询历史已清空')
      loadQueryHistory(pagination.current, pagination.pageSize, dbTypeFilter)
      loadStats()
    } catch (error) {
      console.error('Failed to clear history:', error)
      message.error('清空查询历史失败')
    }
  }

  // 搜索查询
  const handleSearch = (value) => {
    setSearchText(value)
    // 这里可以实现客户端搜索，或者调用后端搜索API
  }

  // 数据库类型过滤
  const handleDbTypeChange = (value) => {
    setDbTypeFilter(value)
    loadQueryHistory(1, pagination.pageSize, value)
  }

  // 再试一次功能
  const handleRetryQuery = async (historyId) => {
    setRetryLoading(prev => ({ ...prev, [historyId]: true }))
    try {
      const result = await RetryQuery(historyId)
      setRetryResult(result)
      setShowResultModal(true)
      message.success('查询重新执行成功')
      // 刷新查询历史
      loadQueryHistory(pagination.current, pagination.pageSize, dbTypeFilter)
    } catch (error) {
      console.error('Retry query failed:', error)
      message.error('重新执行失败: ' + error.message)
    } finally {
      setRetryLoading(prev => ({ ...prev, [historyId]: false }))
    }
  }

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
      sorter: (a, b) => a.id - b.id,
    },
    {
      title: '数据库类型',
      dataIndex: 'dbType',
      key: 'dbType',
      width: 120,
      render: (dbType) => {
        const colors = {
          mysql: 'blue',
          redis: 'red',
          clickhouse: 'green'
        }
        const icons = {
          mysql: '🐬',
          redis: '🔴',
          clickhouse: '⚡'
        }
        return (
          <Tag color={colors[dbType] || 'default'}>
            {icons[dbType] || '📊'} {dbType.toUpperCase()}
          </Tag>
        )
      },
      filters: [
        { text: 'MySQL', value: 'mysql' },
        { text: 'Redis', value: 'redis' },
        { text: 'ClickHouse', value: 'clickhouse' },
      ],
      onFilter: (value, record) => record.dbType === value,
    },
    {
      title: '连接名称',
      dataIndex: 'connectionName',
      key: 'connectionName',
      width: 150,
      render: (name) => (
        <Tag color="purple" icon={<DatabaseOutlined />}>
          {name}
        </Tag>
      ),
    },
    {
      title: '查询语句',
      dataIndex: 'query',
      key: 'query',
      ellipsis: {
        showTitle: false,
      },
      render: (query) => (
        <Tooltip placement="topLeft" title={query}>
          <code style={{ 
            background: '#f5f5f5', 
            padding: '2px 4px', 
            borderRadius: '3px',
            fontSize: '12px',
            maxWidth: '300px',
            display: 'block',
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis'
          }}>
            {query}
          </code>
        </Tooltip>
      ),
    },
    {
      title: '执行时间',
      dataIndex: 'executionTime',
      key: 'executionTime',
      width: 120,
      render: (time) => (
        <Space>
          <ClockCircleOutlined />
          {time}ms
        </Space>
      ),
      sorter: (a, b) => a.executionTime - b.executionTime,
    },
    {
      title: '结果行数',
      dataIndex: 'resultRows',
      key: 'resultRows',
      width: 100,
      sorter: (a, b) => a.resultRows - b.resultRows,
    },
    {
      title: '状态',
      dataIndex: 'success',
      key: 'success',
      width: 100,
      render: (success, record) => (
        <Space>
          {success ? (
            <Tag icon={<CheckCircleOutlined />} color="success">
              成功
            </Tag>
          ) : (
            <Tag icon={<CloseCircleOutlined />} color="error">
              失败
            </Tag>
          )}
        </Space>
      ),
      filters: [
        { text: '成功', value: true },
        { text: '失败', value: false },
      ],
      onFilter: (value, record) => record.success === value,
    },
    {
      title: '执行时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (date) => new Date(date).toLocaleString('zh-CN'),
      sorter: (a, b) => new Date(a.createdAt) - new Date(b.createdAt),
    },
    {
      title: '错误信息',
      dataIndex: 'error',
      key: 'error',
      ellipsis: {
        showTitle: false,
      },
      render: (error) => error ? (
        <Tooltip placement="topLeft" title={error}>
          <span style={{ color: '#ff4d4f' }}>
            {error.length > 50 ? error.substring(0, 50) + '...' : error}
          </span>
        </Tooltip>
      ) : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Space>
          <Button
            type="primary"
            size="small"
            loading={retryLoading[record.id]}
            onClick={() => handleRetryQuery(record.id)}
            icon={<ReloadOutlined />}
          >
            再试一次
          </Button>
        </Space>
      ),
    },
  ]

  // 分页处理
  const handleTableChange = (pagination, filters, sorter) => {
    loadQueryHistory(pagination.current, pagination.pageSize, dbTypeFilter)
  }

  // 初始化加载
  useEffect(() => {
    loadQueryHistory()
    loadStats()
  }, [])

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>
        <DatabaseOutlined style={{ marginRight: '8px' }} />
        SQL 执行历史
      </Title>

      {/* 统计信息 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总查询数"
              value={stats.totalQueries || 0}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
        {stats.statsByType && Object.entries(stats.statsByType).map(([dbType, data]) => (
          <Col span={6} key={dbType}>
            <Card>
              <Statistic
                title={`${dbType.toUpperCase()} 查询`}
                value={data.count || 0}
                suffix={`成功: ${data.successCount || 0}`}
              />
            </Card>
          </Col>
        ))}
      </Row>

      {/* 操作栏 */}
      <Card style={{ marginBottom: '16px' }}>
        <Space wrap>
          <Select
            value={dbTypeFilter}
            onChange={handleDbTypeChange}
            style={{ width: 120 }}
            placeholder="数据库类型"
          >
            <Option value="all">全部</Option>
            <Option value="mysql">MySQL</Option>
            <Option value="redis">Redis</Option>
            <Option value="clickhouse">ClickHouse</Option>
          </Select>

          <Search
            placeholder="搜索查询语句"
            allowClear
            onSearch={handleSearch}
            style={{ width: 300 }}
            prefix={<SearchOutlined />}
          />

          <Button
            icon={<ReloadOutlined />}
            onClick={() => loadQueryHistory(pagination.current, pagination.pageSize, dbTypeFilter)}
            loading={loading}
          >
            刷新
          </Button>

          <Popconfirm
            title="确定要清空所有查询历史吗？"
            onConfirm={clearHistory}
            okText="确定"
            cancelText="取消"
          >
            <Button
              icon={<DeleteOutlined />}
              danger
            >
              清空历史
            </Button>
          </Popconfirm>
        </Space>
      </Card>

      {/* 查询历史表格 */}
      <Card>
        <Table
          columns={columns}
          dataSource={queryHistory}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条记录`,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          onChange={handleTableChange}
          scroll={{ x: 1200 }}
          size="small"
        />
      </Card>

      {/* 执行结果模态框 */}
      <Modal
        title="查询执行结果"
        open={showResultModal}
        onCancel={() => setShowResultModal(false)}
        width={800}
        footer={[
          <Button key="close" onClick={() => setShowResultModal(false)}>
            关闭
          </Button>
        ]}
      >
        {retryResult ? (
          <div>
            {retryResult.error ? (
              <Alert
                message="查询执行失败"
                description={retryResult.error}
                type="error"
                showIcon
              />
            ) : (
              <div>
                <div style={{ marginBottom: '16px' }}>
                  <Space>
                    <Tag color="blue">执行时间: {retryResult.time}ms</Tag>
                    <Tag color="green">结果行数: {retryResult.count}</Tag>
                  </Space>
                </div>
                
                {retryResult.rows && retryResult.rows.length > 0 ? (
                  <Table
                    dataSource={retryResult.rows.map((row, index) => ({ ...row, key: index }))}
                    columns={retryResult.columns?.map((column, index) => ({
                      title: column,
                      dataIndex: index,
                      key: index,
                      ellipsis: true
                    })) || []}
                    bordered
                    size="small"
                    scroll={{ y: 400 }}
                    pagination={false}
                  />
                ) : (
                  <Empty description="查询结果为空" />
                )}
              </div>
            )}
          </div>
        ) : (
          <Empty description="暂无执行结果" />
        )}
      </Modal>
    </div>
  )
}

export default QueryHistory
