import React, { useState, useEffect, useRef, useCallback } from 'react'
import { 
  Card, 
  Button, 
  Row, 
  Col, 
  Modal, 
  Tag, 
  Table, 
  Alert, 
  Empty,
  Space,
  Typography,
  Input,
  Dropdown,
  Menu
} from 'antd'
import { 
  LinkOutlined, 
  MenuOutlined, 
  DatabaseOutlined, 
  FolderOutlined,
  TableOutlined,
  FileTextOutlined,
  EditOutlined,
  PlayCircleOutlined,
  BarChartOutlined,
  PlusOutlined,
  SettingOutlined,
  DownOutlined
} from '@ant-design/icons'
import { useConnectionStore } from '../stores/connection'
import { useQueryStore } from '../stores/query'
import { message } from 'antd'
import ConnectionManager from '../components/ConnectionManager'

const { TextArea } = Input
const { Title } = Typography

function Query({ type = 'mysql' }) {
  const { 
    connections, 
    connectedConnections, 
    connectToDatabase,
    autoSelectConnection,
    loadConnections
  } = useConnectionStore()
  const { 
    getActiveConnection,
    getCurrentQuery,
    getQueryResult,
    getDatabases,
    getTables,
    getLoading,
    setCurrentQuery,
    setActiveConnection,
    executeQuery: executeQueryAction,
    loadDatabases,
    loadTables,
    formatQuery: formatQueryAction,
    resetQueryState,
    setCurrentType
  } = useQueryStore()

  // 获取当前状态
  const activeConnection = getActiveConnection()
  const currentQuery = getCurrentQuery()
  const queryResult = getQueryResult()
  const databases = getDatabases()
  const tables = getTables()
  const loading = getLoading()

  const [showConnectionDialog, setShowConnectionDialog] = useState(false)
  const [showConnectionManager, setShowConnectionManager] = useState(false)
  const [selectedConnection, setSelectedConnection] = useState(null)
  const [selectedDatabase, setSelectedDatabase] = useState('')
  const [selectedTable, setSelectedTable] = useState('')
  
  // 分割线相关状态
  const [queryPanelHeight, setQueryPanelHeight] = useState(300) // 查询面板固定高度（像素）
  const [isDragging, setIsDragging] = useState(false)
  const containerRef = useRef(null)

  // 根据类型过滤连接
  const filteredConnections = connections.filter(conn => conn.type === type)

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

  // 格式化时间
  const formatTime = (duration) => {
    if (duration < 1000) {
      return `${duration}ms`
    } else {
      return `${(duration / 1000).toFixed(2)}s`
    }
  }

  // 选择数据库
  const handleSelectDatabase = async (database) => {
    console.log('选择数据库:', database, '连接ID:', activeConnection?.id)
    setSelectedDatabase(database)
    setSelectedTable('')
    
    try {
      const result = await loadTables(activeConnection.id, database)
      console.log('加载表列表结果:', result)
    } catch (error) {
      console.error('加载表列表失败:', error)
      message.error('加载表列表失败: ' + error.message)
    }
  }

  // 选择表
  const handleSelectTable = async (table) => {
    setSelectedTable(table.name)
    
    try {
      // 生成查询语句
      const query = `SELECT * FROM \`${selectedDatabase}\`.\`${table.name}\` LIMIT 100`
      setCurrentQuery(query)
      
      // 执行查询
      await executeQueryAction(activeConnection.id, query)
    } catch (error) {
      message.error('加载表数据失败')
    }
  }

  // 执行查询
  const handleExecuteQuery = async () => {
    if (!activeConnection) {
      message.warning('请先选择连接')
      return
    }
    
    if (!currentQuery.trim()) {
      message.warning('请输入查询语句')
      return
    }
    
    try {
      console.log('执行查询:', { connectionId: activeConnection.id, query: currentQuery })
      const result = await executeQueryAction(activeConnection.id, currentQuery)
      console.log('查询结果:', result)
    } catch (error) {
      console.error('查询执行失败:', error)
      message.error('查询执行失败: ' + error.message)
    }
  }

  // 格式化查询
  const handleFormatQuery = async () => {
    if (!activeConnection) {
      message.warning('请先选择连接')
      return
    }
    
    try {
      const formatted = await formatQueryAction(activeConnection.id, currentQuery)
      setCurrentQuery(formatted)
    } catch (error) {
      message.error('格式化失败')
    }
  }

  // 确认连接选择
  const handleConfirmConnection = async () => {
    if (selectedConnection) {
      try {
        // 先建立连接
        await connectToDatabase(selectedConnection.id)
        // 然后设置为活动连接（同时更新queryStore）
        setActiveConnection(selectedConnection)
        setShowConnectionDialog(false)
        message.success('连接成功')
      } catch (error) {
        message.error('连接失败: ' + (error.message || '未知错误'))
      }
    }
  }

  // 加载数据库列表
  const handleLoadDatabases = async () => {
    if (!activeConnection) return
    
    try {
      await loadDatabases(activeConnection.id)
    } catch (error) {
      message.error('加载数据库列表失败')
    }
  }

  // 键盘快捷键处理
  const handleKeyDown = (event) => {
    if (event.ctrlKey && event.key === 'Enter') {
      event.preventDefault()
      handleExecuteQuery()
    }
  }


  // 拖拽处理函数
  const handleMouseDown = useCallback((e) => {
    e.preventDefault()
    setIsDragging(true)
    
    const startY = e.clientY
    const startHeight = queryPanelHeight
    
    const handleMouseMove = (e) => {
      const deltaY = e.clientY - startY
      const newHeight = Math.max(200, Math.min(600, startHeight + deltaY))
      setQueryPanelHeight(newHeight)
    }
    
    const handleMouseUp = () => {
      setIsDragging(false)
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
    
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }, [queryPanelHeight])

  // 重置分割线高度
  const resetSplitter = () => {
    setQueryPanelHeight(300) // 重置为 300px 固定高度
  }

  useEffect(() => {
    // 设置当前数据库类型
    setCurrentType(type)
    console.log(`Switched to ${type} page, current activeConnection:`, activeConnection)
  }, [type, setCurrentType])

  // 初始化时加载连接并自动选择默认连接
  useEffect(() => {
    const initializeConnections = async () => {
      try {
        console.log(`Initializing ${type} page...`)
        
        // 加载连接列表
        await loadConnections()
        console.log('Connections loaded, checking for auto-selection...')
        
        // 短暂延迟确保连接状态已更新
        setTimeout(() => {
          console.log(`Checking auto-selection for ${type}:`)
          console.log('Current activeConnection:', activeConnection)
          
          // 检查是否需要自动选择连接
          const needAutoSelection = !activeConnection || activeConnection.type !== type
          
          if (needAutoSelection) {
            console.log(`Need auto-selection for ${type} (reason: ${!activeConnection ? 'no connection' : 'type mismatch'})`)
            const defaultConnection = autoSelectConnection(type)
            if (defaultConnection) {
              console.log(`Auto-selected ${type} connection:`, defaultConnection.name)
              // 同时设置到queryStore中
              setActiveConnection(defaultConnection)
            } else {
              console.log(`No connected ${type} database found for auto-selection`)
            }
          } else {
            console.log('Active connection matches page type:', activeConnection.name, activeConnection.type)
          }
        }, 100)
      } catch (error) {
        console.error('Failed to initialize connections:', error)
      }
    }

    initializeConnections()
  }, [type]) // 只在type变化时执行

  useEffect(() => {
    // 如果有活动连接，加载数据库列表
    if (activeConnection) {
      handleLoadDatabases()
    } else {
      setSelectedDatabase('')
      setSelectedTable('')
    }
  }, [activeConnection])

  const columns = queryResult?.columns?.map((column, index) => ({
    title: column,
    dataIndex: `col_${index}`,
    key: `col_${index}`,
    width: 150,
    ellipsis: true,
    render: (text) => {
      if (text === null || text === undefined) return <span style={{ color: '#999' }}>NULL</span>
      return String(text)
    }
  })) || []
  
  // 调试信息
  if (queryResult) {
    console.log('查询结果详情:', {
      columns: queryResult.columns,
      rows: queryResult.rows,
      count: queryResult.count,
      time: queryResult.time,
      error: queryResult.error
    })
    console.log('生成的列配置:', columns)
  }
  
  // 调试表数据
  console.log('当前表数据:', {
    selectedDatabase,
    tablesCount: tables.length,
    tables: tables.map(t => t.name),
    activeConnection: activeConnection?.id,
    currentType: type
  })

  return (
    <div 
      className="query-page" 
      style={{ 
        padding: '20px', 
        height: 'calc(100vh - 60px)', 
        overflow: 'hidden',
        userSelect: isDragging ? 'none' : 'auto'
      }}
    >
      {/* 连接选择器 */}
      <div style={{ marginBottom: '20px' }}>
        <Card>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '16px', fontWeight: 'bold' }}>
              <LinkOutlined style={{ color: '#409eff' }} />
              {activeConnection ? (
                <span>
                  {activeConnection.name} ({activeConnection.type.toUpperCase()})
                </span>
              ) : (
                <span style={{ color: '#909399' }}>请先选择一个连接</span>
              )}
            </div>
            <Space>
              <Button 
                type="primary" 
                onClick={() => setShowConnectionDialog(true)}
                disabled={filteredConnections.length === 0}
              >
                选择连接
              </Button>
              <Dropdown
                overlay={
                  <Menu>
                    <Menu.Item 
                      key="add" 
                      icon={<PlusOutlined />}
                      onClick={() => setShowConnectionManager(true)}
                    >
                      添加连接
                    </Menu.Item>
                    <Menu.Item 
                      key="manage" 
                      icon={<SettingOutlined />}
                      onClick={() => setShowConnectionManager(true)}
                    >
                      管理连接
                    </Menu.Item>
                  </Menu>
                }
                trigger={['click']}
              >
                <Button icon={<DownOutlined />}>
                  更多
                </Button>
              </Dropdown>
            </Space>
          </div>
        </Card>
      </div>

      <Row gutter={20} style={{ height: 'calc(100% - 100px)' }}>
        {/* 左侧：数据库浏览器 */}
        <Col span={6}>
          <Card 
            title={
              <Space>
                <MenuOutlined />
                <span>数据库浏览器</span>
              </Space>
            }
            style={{ height: '100%', overflow: 'auto' }}
          >
            {!activeConnection ? (
              <Empty description="请先选择连接" />
            ) : (
              <div style={{ height: 'calc(100% - 40px)', overflowY: 'auto' }}>
                {/* 数据库列表 */}
                <div style={{ marginBottom: '20px' }}>
                  <div style={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    gap: '8px', 
                    fontWeight: 'bold', 
                    color: '#303133', 
                    marginBottom: '12px', 
                    paddingBottom: '8px', 
                    borderBottom: '1px solid #e4e7ed' 
                  }}>
                    <DatabaseOutlined style={{ color: '#409eff' }} />
                    <span>数据库</span>
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                    {databases.map(db => (
                      <div
                        key={db}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '8px',
                          padding: '8px 12px',
                          borderRadius: '4px',
                          cursor: 'pointer',
                          transition: 'all 0.3s',
                          backgroundColor: selectedDatabase === db ? '#e6f7ff' : 'transparent',
                          color: selectedDatabase === db ? '#409eff' : 'inherit'
                        }}
                        onClick={() => handleSelectDatabase(db)}
                        onMouseEnter={(e) => {
                          if (selectedDatabase !== db) {
                            e.target.style.backgroundColor = '#f5f7fa'
                          }
                        }}
                        onMouseLeave={(e) => {
                          if (selectedDatabase !== db) {
                            e.target.style.backgroundColor = 'transparent'
                          }
                        }}
                      >
                        <FolderOutlined style={{ color: selectedDatabase === db ? '#409eff' : '#909399' }} />
                        <span>{db}</span>
                      </div>
                    ))}
                  </div>
                </div>
                
                {/* 表列表 */}
                {selectedDatabase && (
                  <div>
                    <div style={{ 
                      display: 'flex', 
                      alignItems: 'center', 
                      gap: '8px', 
                      fontWeight: 'bold', 
                      color: '#303133', 
                      marginBottom: '12px', 
                      paddingBottom: '8px', 
                      borderBottom: '1px solid #e4e7ed' 
                    }}>
                      <TableOutlined style={{ color: '#409eff' }} />
                      <span>表 ({tables.length})</span>
                    </div>
                    <div style={{ 
                      display: 'flex', 
                      flexDirection: 'column', 
                      gap: '4px',
                      maxHeight: '300px',
                      overflowY: 'auto'
                    }}>
                      {tables.length === 0 ? (
                        <div style={{ 
                          padding: '20px', 
                          textAlign: 'center', 
                          color: '#909399',
                          fontSize: '14px'
                        }}>
                          {loading ? '加载中...' : '暂无表数据'}
                        </div>
                      ) : (
                        tables.map(table => (
                          <div
                            key={table.name}
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: '8px',
                              padding: '8px 12px',
                              borderRadius: '4px',
                              cursor: 'pointer',
                              transition: 'all 0.3s',
                              backgroundColor: selectedTable === table.name ? '#e6f7ff' : 'transparent',
                              color: selectedTable === table.name ? '#409eff' : 'inherit'
                            }}
                            onClick={() => handleSelectTable(table)}
                            onMouseEnter={(e) => {
                              if (selectedTable !== table.name) {
                                e.target.style.backgroundColor = '#f5f7fa'
                              }
                            }}
                            onMouseLeave={(e) => {
                              if (selectedTable !== table.name) {
                                e.target.style.backgroundColor = 'transparent'
                              }
                            }}
                          >
                            <FileTextOutlined style={{ color: selectedTable === table.name ? '#409eff' : '#909399' }} />
                            <span>{table.name}</span>
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                )}
              </div>
            )}
          </Card>
        </Col>
        
        {/* 右侧：查询编辑器和结果区域 */}
        <Col span={18}>
          <div 
            ref={containerRef}
            style={{ 
              height: '100%', 
              display: 'flex', 
              flexDirection: 'column',
              position: 'relative'
            }}
          >
            {/* SQL 查询编辑器 */}
            <Card 
              title={
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontWeight: 'bold' }}>
                  <EditOutlined style={{ color: '#409eff' }} />
                  <span>SQL 查询</span>
                  <div style={{ marginLeft: 'auto', display: 'flex', gap: '8px' }}>
                    <Button 
                      size="small" 
                      onClick={handleFormatQuery}
                      disabled={!activeConnection}
                    >
                      格式化
                    </Button>
                    <Button 
                      type="primary" 
                      size="small" 
                      onClick={handleExecuteQuery}
                      loading={loading}
                      disabled={!activeConnection}
                      icon={<PlayCircleOutlined />}
                    >
                      执行
                    </Button>
                  </div>
                </div>
              }
              style={{ 
                height: `${queryPanelHeight}px`, 
                marginBottom: '0',
                display: 'flex',
                flexDirection: 'column',
                flexShrink: 0
              }}
              bodyStyle={{ 
                flex: 1, 
                display: 'flex', 
                flexDirection: 'column',
                padding: '16px'
              }}
            >
              <TextArea
                value={currentQuery}
                onChange={(e) => setCurrentQuery(e.target.value)}
                placeholder="请输入 SQL 查询语句..."
                style={{
                  height: '100%',
                  fontFamily: 'Monaco, Menlo, Consolas, monospace',
                  fontSize: '14px',
                  lineHeight: '1.5',
                  resize: 'none'
                }}
                onKeyDown={handleKeyDown}
              />
            </Card>
            
            {/* 可拖动的分割线 */}
            <div
              style={{
                height: '12px',
                backgroundColor: isDragging ? '#409eff' : '#f0f0f0',
                cursor: 'ns-resize',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                position: 'relative',
                transition: isDragging ? 'none' : 'background-color 0.2s',
                borderTop: '1px solid #e4e7ed',
                borderBottom: '1px solid #e4e7ed',
                userSelect: 'none',
                zIndex: 10
              }}
              onMouseDown={handleMouseDown}
            >
              <div
                style={{
                  width: '60px',
                  height: '6px',
                  backgroundColor: isDragging ? '#409eff' : '#c0c4cc',
                  borderRadius: '3px',
                  transition: isDragging ? 'none' : 'background-color 0.2s',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  pointerEvents: 'none'
                }}
              >
                <div
                  style={{
                    width: '20px',
                    height: '2px',
                    backgroundColor: isDragging ? '#fff' : '#909399',
                    borderRadius: '1px'
                  }}
                />
              </div>
              {/* 重置按钮 */}
              <Button
                size="small"
                type="text"
                style={{
                  position: 'absolute',
                  right: '10px',
                  fontSize: '12px',
                  color: isDragging ? '#409eff' : '#909399',
                  padding: '0 8px',
                  height: '20px',
                  lineHeight: '20px',
                  zIndex: 11
                }}
                onMouseDown={(e) => e.stopPropagation()}
                onClick={(e) => {
                  e.stopPropagation()
                  resetSplitter()
                }}
              >
                重置
              </Button>
            </div>
            
            {/* 查询结果区域 */}
            <div style={{ flex: 1, minHeight: '200px', overflow: 'auto' }}>
              {queryResult ? (
                <Card 
                  title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontWeight: 'bold' }}>
                      <BarChartOutlined style={{ color: '#409eff' }} />
                      <span>查询结果</span>
                      <div style={{ marginLeft: 'auto', display: 'flex', gap: '8px' }}>
                        {queryResult.count !== undefined && (
                          <Tag>共 {queryResult.count} 行</Tag>
                        )}
                        {queryResult.time && (
                          <Tag>耗时 {formatTime(queryResult.time)}</Tag>
                        )}
                      </div>
                    </div>
                  }
                  style={{ 
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column'
                  }}
                  bodyStyle={{ 
                    flex: 1, 
                    display: 'flex', 
                    flexDirection: 'column',
                    padding: '16px'
                  }}
                >
                  {queryResult.error ? (
                    <Alert
                      message={queryResult.error}
                      type="error"
                      showIcon
                    />
                  ) : !queryResult.rows || queryResult.rows.length === 0 ? (
                    <Empty description="查询结果为空" />
                  ) : (
                    <div style={{ 
                      flex: 1, 
                      overflow: 'auto'
                    }}>
                      <Table 
                        dataSource={queryResult.rows.map((row, rowIndex) => {
                          const rowData = { key: rowIndex }
                          row.forEach((cell, colIndex) => {
                            rowData[`col_${colIndex}`] = cell
                          })
                          return rowData
                        })} 
                        columns={columns}
                        bordered 
                        size="small"
                        scroll={{ y: 400 }}
                        pagination={{
                          pageSize: 100,
                          showSizeChanger: true,
                          showQuickJumper: true,
                          showTotal: (total, range) => `显示 ${range[0]}-${range[1]} 条，共 ${total} 条`
                        }}
                      />
                    </div>
                  )}
                </Card>
              ) : (
                <Card 
                  title={
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontWeight: 'bold' }}>
                      <BarChartOutlined style={{ color: '#409eff' }} />
                      <span>查询结果</span>
                    </div>
                  }
                  style={{ 
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column'
                  }}
                  bodyStyle={{ 
                    flex: 1, 
                    display: 'flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    padding: '40px 16px'
                  }}
                >
                  <Empty 
                    description="执行查询以查看结果" 
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                  />
                </Card>
              )}
            </div>
          </div>
        </Col>
      </Row>

      {/* 连接选择对话框 */}
      <Modal
        title="选择连接"
        open={showConnectionDialog}
        onCancel={() => setShowConnectionDialog(false)}
        onOk={handleConfirmConnection}
        width={600}
        okText="确定"
        cancelText="取消"
        okButtonProps={{ disabled: !selectedConnection }}
      >
        <div style={{ maxHeight: '400px', overflowY: 'auto' }}>
          {filteredConnections.map(conn => (
            <div
              key={conn.id}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '12px',
                border: '1px solid #e4e7ed',
                borderRadius: '4px',
                marginBottom: '8px',
                cursor: 'pointer',
                transition: 'all 0.3s',
                backgroundColor: selectedConnection?.id === conn.id ? '#e6f7ff' : 'transparent',
                borderColor: selectedConnection?.id === conn.id ? '#409eff' : '#e4e7ed'
              }}
              onClick={() => setSelectedConnection(conn)}
              onMouseEnter={(e) => {
                if (selectedConnection?.id !== conn.id) {
                  e.target.style.borderColor = '#409eff'
                  e.target.style.backgroundColor = '#f5f7fa'
                }
              }}
              onMouseLeave={(e) => {
                if (selectedConnection?.id !== conn.id) {
                  e.target.style.borderColor = '#e4e7ed'
                  e.target.style.backgroundColor = 'transparent'
                }
              }}
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
              <div>
                <Tag color={getStatusType(conn.status)}>
                  {getStatusText(conn.status)}
                </Tag>
              </div>
            </div>
          ))}
        </div>
      </Modal>

      {/* 连接管理对话框 */}
      <ConnectionManager
        type={type}
        visible={showConnectionManager}
        onClose={() => setShowConnectionManager(false)}
      />
    </div>
  )
}

export default Query
