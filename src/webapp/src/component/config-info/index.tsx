import React from "react";
import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs';
import 'prismjs/components/prism-yaml';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-javascript';
import 'prismjs/themes/prism.css'; //Example style, you can use another
import API from '../../utils/api';
import { 
  Button, 
  Card, 
  Form, 
  Input, 
  InputNumber, 
  Row, 
  Col, 
  Icon,
  Select,
  Tooltip,
  message
} from "antd";
import './config-info.css';

const { Option } = Select;
const api = new API();

interface Props {

}

interface IState {
  config: any;
  parsedConfig: any;
  editMode: 'gui' | 'text';
  loading: boolean;
  selectedPlatform: string;
}

// 平台配置列表 - 包含所有支持的直播平台
const PLATFORM_OPTIONS = [
  { key: 'bilibili', name: '哔哩哔哩', domain: 'live.bilibili.com' },
  { key: 'douyin', name: '抖音', domain: 'live.douyin.com' },
  { key: 'douyu', name: '斗鱼', domain: 'www.douyu.com' },
  { key: 'huya', name: '虎牙', domain: 'www.huya.com' },
  { key: 'kuaishou', name: '快手', domain: 'live.kuaishou.com' },
  { key: 'yy', name: 'YY直播', domain: 'www.yy.com' },
  { key: 'acfun', name: 'AcFun', domain: 'live.acfun.cn' },
  { key: 'lang', name: '浪live', domain: 'www.lang.live' },
  { key: 'missevan', name: '猫耳', domain: 'fm.missevan.com' },
  { key: 'openrec', name: 'OpenRec', domain: 'www.openrec.tv' },
  { key: 'weibolive', name: '微博直播', domain: 'weibo.com' },
  { key: 'xiaohongshu', name: '小红书', domain: 'www.xiaohongshu.com' },
  { key: 'yizhibo', name: '一直播', domain: 'www.yizhibo.com' },
  { key: 'hongdoufm', name: '克拉克拉', domain: 'www.hongdoufm.com' },
  { key: 'zhanqi', name: '战旗', domain: 'www.zhanqi.tv' },
  { key: 'cc', name: 'CC直播', domain: 'cc.163.com' },
  { key: 'twitch', name: 'Twitch', domain: 'www.twitch.tv' },
  { key: 'qq', name: '企鹅电竞', domain: 'egame.qq.com' },
  { key: 'huajiao', name: '花椒', domain: 'www.huajiao.com' },
];

class ConfigInfo extends React.Component<Props, IState> {

  constructor(props: Props) {
    super(props);
    this.state = {
      config: null,
      parsedConfig: {},
      editMode: 'gui',
      loading: false,
      selectedPlatform: 'bilibili',
    }
  }

  componentDidMount(): void {
    this.loadConfig();
  }

  loadConfig = () => {
    this.setState({ loading: true });
    api.getConfigInfo()
      .then((rsp: any) => {
        try {
          // Try to parse YAML config
          const parsedConfig = this.parseYamlConfig(rsp.config);
          this.setState({
            config: rsp.config,
            parsedConfig: parsedConfig,
            loading: false,
          });
        } catch (e) {
          // Fallback to text mode if parsing fails
          this.setState({
            config: rsp.config,
            parsedConfig: {},
            editMode: 'text',
            loading: false,
          });
        }
      })
      .catch(err => {
        message.error("获取配置信息失败");
        this.setState({ loading: false });
      });
  }

  parseYamlConfig = (yamlText: string) => {
    // Simple YAML parser for basic config structure
    const lines = yamlText.split('\n');
    const config: any = {
      interval: 30,
      out_put_path: './',
      platform_configs: {},
      live_rooms: []
    };

    let currentSection = '';
    let currentPlatform = '';

    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith('#')) continue;

      const currentIndent = line.length - line.trimLeft().length;
      
      if (trimmed.includes(':')) {
        const [key, value] = trimmed.split(':', 2);
        const cleanKey = key.trim();
        const cleanValue = value ? value.trim() : '';

        if (currentIndent === 0) {
          if (cleanKey === 'platform_configs') {
            currentSection = 'platform_configs';
          } else if (cleanKey === 'live_rooms') {
            currentSection = 'live_rooms';
          } else if (cleanValue) {
            config[cleanKey] = this.parseValue(cleanValue);
          }
        } else if (currentSection === 'platform_configs' && currentIndent === 2) {
          currentPlatform = cleanKey;
          if (!config.platform_configs[currentPlatform]) {
            config.platform_configs[currentPlatform] = {};
          }
        } else if (currentSection === 'platform_configs' && currentIndent === 4 && currentPlatform) {
          config.platform_configs[currentPlatform][cleanKey] = this.parseValue(cleanValue);
        }
      }
    }

    return config;
  }

  parseValue = (value: string) => {
    if (value === 'true') return true;
    if (value === 'false') return false;
    if (!isNaN(Number(value))) return Number(value);
    return value;
  }

  generateYamlConfig = () => {
    const { parsedConfig } = this.state;
    let yaml = '';

    // Global settings
    yaml += `rpc:\n  enable: true\n  bind: :8080\n\n`;
    yaml += `debug: false\n`;
    yaml += `interval: ${parsedConfig.interval || 30}\n`;
    yaml += `out_put_path: ${parsedConfig.out_put_path || './'}\n`;
    yaml += `ffmpeg_path: ${parsedConfig.ffmpeg_path || ''}\n\n`;

    // Platform configs
    if (parsedConfig.platform_configs && Object.keys(parsedConfig.platform_configs).length > 0) {
      yaml += `platform_configs:\n`;
      Object.keys(parsedConfig.platform_configs).forEach(platform => {
        const config = parsedConfig.platform_configs[platform];
        yaml += `  ${platform}:\n`;
        if (config.name) yaml += `    name: "${config.name}"\n`;
        if (config.min_access_interval_sec) yaml += `    min_access_interval_sec: ${config.min_access_interval_sec}\n`;
        if (config.interval) yaml += `    interval: ${config.interval}\n`;
        if (config.out_put_path) yaml += `    out_put_path: ${config.out_put_path}\n`;
      });
      yaml += '\n';
    }

    // Live rooms (preserve existing rooms)
    yaml += `live_rooms: []\n`;
    yaml += `cookies: {}\n`;
    yaml += `timeout_in_us: 60000000\n`;

    return yaml;
  }

  onGuiConfigChange = (field: string, value: any, platform?: string) => {
    const { parsedConfig } = this.state;
    const newConfig = { ...parsedConfig };

    if (platform) {
      if (!newConfig.platform_configs) newConfig.platform_configs = {};
      if (!newConfig.platform_configs[platform]) newConfig.platform_configs[platform] = {};
      newConfig.platform_configs[platform][field] = value;
    } else {
      newConfig[field] = value;
    }

    // Regenerate YAML
    const newYaml = this.generateYamlConfig();

    this.setState({
      parsedConfig: newConfig,
      config: newYaml,
    });
  }

  onSettingSave = () => {
    this.setState({ loading: true });
    api.saveRawConfig({ config: this.state.config })
      .then((rsp: any) => {
        if (rsp.err_no === 0) {
          message.success("设置保存成功");
        } else {
          message.error(`Server Error!\n${rsp.err_msg}`);
        }
        this.setState({ loading: false });
      })
      .catch(err => {
        message.error("设置保存失败！");
        this.setState({ loading: false });
      })
  }

  renderGlobalSettings = () => {
    const { parsedConfig } = this.state;

    return (
      <Card title="全局设置" style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="检测间隔 (秒)">
              <Tooltip title="全局检测间隔，可被平台和房间级设置覆盖">
                <InputNumber
                  value={parsedConfig.interval || 30}
                  min={1}
                  max={3600}
                  onChange={(value) => this.onGuiConfigChange('interval', value)}
                />
              </Tooltip>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="输出路径">
              <Tooltip title="全局录制文件保存路径，可被平台和房间级设置覆盖">
                <Input
                  value={parsedConfig.out_put_path || './'}
                  onChange={(e) => this.onGuiConfigChange('out_put_path', e.target.value)}
                />
              </Tooltip>
            </Form.Item>
          </Col>
        </Row>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="FFmpeg 路径">
              <Tooltip title="FFmpeg 可执行文件路径，可被平台和房间级设置覆盖">
                <Input
                  value={parsedConfig.ffmpeg_path || ''}
                  placeholder="留空自动在环境变量中寻找"
                  onChange={(e) => this.onGuiConfigChange('ffmpeg_path', e.target.value)}
                />
              </Tooltip>
            </Form.Item>
          </Col>
        </Row>
      </Card>
    );
  }

  renderPlatformSettings = () => {
    const { parsedConfig, selectedPlatform } = this.state;
    const platformConfigs = parsedConfig.platform_configs || {};

    return (
      <Card 
        title="平台特定设置" 
        extra={
          <Select
            value={selectedPlatform}
            style={{ width: 120 }}
            onChange={(value) => this.setState({ selectedPlatform: value })}
          >
            {PLATFORM_OPTIONS.map(platform => (
              <Option key={platform.key} value={platform.key}>
                {platform.name}
              </Option>
            ))}
          </Select>
        }
        style={{ marginBottom: 16 }}
      >
        <div style={{ marginBottom: 16 }}>
          <Tooltip title="平台级设置将覆盖全局设置，但会被房间级设置覆盖">
            <Icon type="info-circle" /> 设置优先级：房间级 &gt; 平台级 &gt; 全局级
          </Tooltip>
        </div>
        
        {selectedPlatform && (
          <div>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="平台访问频率限制 (秒)">
                  <Tooltip title="对该平台的最小访问间隔，防止触发风控">
                    <InputNumber
                      value={platformConfigs[selectedPlatform]?.min_access_interval_sec || 0}
                      min={0}
                      max={60}
                      onChange={(value) => this.onGuiConfigChange('min_access_interval_sec', value, selectedPlatform)}
                      placeholder="0 = 无限制"
                    />
                  </Tooltip>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label="检测间隔 (秒)">
                  <Tooltip title="覆盖全局检测间隔，仅对该平台生效">
                    <InputNumber
                      value={platformConfigs[selectedPlatform]?.interval}
                      min={1}
                      max={3600}
                      placeholder="使用全局设置"
                      onChange={(value) => this.onGuiConfigChange('interval', value, selectedPlatform)}
                    />
                  </Tooltip>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item label="输出路径">
                  <Tooltip title="覆盖全局输出路径，仅对该平台生效">
                    <Input
                      value={platformConfigs[selectedPlatform]?.out_put_path || ''}
                      placeholder="使用全局设置"
                      onChange={(e) => this.onGuiConfigChange('out_put_path', e.target.value, selectedPlatform)}
                    />
                  </Tooltip>
                </Form.Item>
              </Col>
            </Row>
          </div>
        )}
      </Card>
    );
  }

  render() {
    const { config, editMode, loading } = this.state;

    if (loading || config === null) {
      return <div>loading...</div>;
    }

    return (
      <div>
        <div style={{ marginBottom: 16, textAlign: 'right' }}>
          <Button.Group>
            <Button 
              type={editMode === 'gui' ? 'primary' : 'default'}
              onClick={() => this.setState({ editMode: 'gui' })}
            >
              <Icon type="form" /> GUI 模式
            </Button>
            <Button 
              type={editMode === 'text' ? 'primary' : 'default'}
              onClick={() => this.setState({ editMode: 'text' })}
            >
              <Icon type="code" /> 文本模式
            </Button>
          </Button.Group>
        </div>

        {editMode === 'gui' ? (
          <div>
            {this.renderGlobalSettings()}
            {this.renderPlatformSettings()}
            
            <Card title="配置层次结构说明">
              <p><Icon type="info-circle" style={{ color: '#1890ff' }} /> 本程序支持三级配置覆盖：</p>
              <ul>
                <li><strong>全局级</strong>：适用于所有直播间的默认设置</li>
                <li><strong>平台级</strong>：适用于特定平台的所有直播间，覆盖全局设置</li>
                <li><strong>房间级</strong>：适用于单个直播间，覆盖平台和全局设置</li>
              </ul>
              <p>平台访问频率限制可以有效防止被直播平台风控，建议根据平台特性合理设置。</p>
              <p><strong>支持的平台：</strong>{PLATFORM_OPTIONS.map(p => p.name).join('、')}</p>
            </Card>
          </div>
        ) : (
          <Editor
            value={config}
            onValueChange={code => this.setState({ config: code })}
            highlight={code => {
              return highlight(code, languages.yaml, 'yaml');
            }}
            padding={10}
            style={{
              fontFamily: '"Fira code", "Fira Mono", monospace',
              fontSize: 12,
              border: '1px solid #d9d9d9',
              borderRadius: 4,
            }}
          />
        )}
        
        <div style={{ marginTop: 16, textAlign: 'center' }}>
          <Button
            type="primary"
            loading={loading}
            onClick={this.onSettingSave}
          >
            保存设置
          </Button>
        </div>
      </div>
    );
  }
}

export default ConfigInfo;