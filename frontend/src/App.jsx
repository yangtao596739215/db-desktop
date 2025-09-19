import React, { useEffect, useState } from 'react'
import { Layout, Menu } from 'antd'
import { 
  DatabaseOutlined, 
  SettingOutlined, 
  ApiOutlined,
  CloudServerOutlined,
  HddOutlined
} from '@ant-design/icons'
import { useConnectionStore } from './stores/connection'
import Query from './views/Query'
import Settings from './views/Settings'
import './App.css'

const { Sider, Content } = Layout

function App() {
  const [activeTab, setActiveTab] = useState('mysql')
  const { loadConnections } = useConnectionStore()

  useEffect(() => {
    // 加载保存的连接
    loadConnections()
  }, [loadConnections])

  const menuItems = [
    {
      key: 'mysql',
      icon: <CloudServerOutlined />,
      label: 'MySQL',
    },
    {
      key: 'clickhouse',
      icon: <HddOutlined />,
      label: 'ClickHouse',
    },
    {
      key: 'redis',
      icon: <ApiOutlined />,
      label: 'Redis',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
  ]

  const handleMenuClick = ({ key }) => {
    setActiveTab(key)
  }

  // 渲染当前活跃的页面
  const renderActivePage = () => {
    switch (activeTab) {
      case 'mysql':
        return <Query type="mysql" />
      case 'clickhouse':
        return <Query type="clickhouse" />
      case 'redis':
        return <Query type="redis" />
      case 'settings':
        return <Settings />
      default:
        return <Query type="mysql" />
    }
  }

  return (
    <div className="app-container">
      <Layout style={{ minHeight: '100vh' }}>
        {/* 侧边栏 */}
        <Sider width={250} className="sidebar">
          <div className="logo">
            <DatabaseOutlined style={{ fontSize: '24px', marginRight: '10px', color: '#3498db' }} />
            <span>DB Desktop</span>
          </div>
          
          <Menu
            mode="inline"
            selectedKeys={[activeTab]}
            className="sidebar-menu"
            items={menuItems}
            theme="dark"
            onClick={handleMenuClick}
          />
        </Sider>
        
        {/* 主内容区 */}
        <Layout>
          <Content className="main-content">
            {renderActivePage()}
          </Content>
        </Layout>
      </Layout>
    </div>
  )
}

export default App
