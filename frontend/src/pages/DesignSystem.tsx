import { useState } from 'react';
import toast from 'react-hot-toast';
import {
  Activity,
  AlertTriangle,
  Bell,
  Cpu,
  Flame,
  Grid,
  LayoutGrid,
  Network,
  Route,
  Settings,
  Share2,
  Terminal,
} from 'lucide-react';
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
} from 'recharts';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { CodeBlock } from '../components/ui/CodeBlock';
import { MetricCard } from '../components/ui/MetricCard';
import { SearchInput } from '../components/ui/SearchInput';
import { Select } from '../components/ui/Select';
import { Skeleton } from '../components/ui/Skeleton';
import { StatusDot } from '../components/ui/StatusDot';
import { PageHeader } from '../components/ui/PageStates';

const COLORS = [
  { name: 'bg-base', var: '--bg-base' },
  { name: 'bg-surface', var: '--bg-surface' },
  { name: 'accent-primary', var: '--accent-primary' },
  { name: 'severity-critical', var: '--severity-critical' },
  { name: 'status-healthy', var: '--status-healthy' },
  { name: 'text-primary', var: '--text-primary' },
  { name: 'text-secondary', var: '--text-secondary' },
];

const ICONS = [
  { Icon: LayoutGrid, name: 'LayoutGrid' },
  { Icon: Flame, name: 'Flame' },
  { Icon: Terminal, name: 'Terminal' },
  { Icon: Route, name: 'Route' },
  { Icon: Activity, name: 'Activity' },
  { Icon: Share2, name: 'Share2' },
  { Icon: Cpu, name: 'Cpu' },
  { Icon: Bell, name: 'Bell' },
  { Icon: Settings, name: 'Settings' },
  { Icon: Network, name: 'Network' },
];

const chartData = [
  { t: '00:00', v: 2 },
  { t: '04:00', v: 3 },
  { t: '08:00', v: 8 },
  { t: '12:00', v: 5 },
  { t: '16:00', v: 12 },
  { t: '20:00', v: 7 },
];

export default function DesignSystem() {
  const [theme, setTheme] = useState<'dark' | 'light'>('dark');

  const copyClasses = (text: string) => {
    void navigator.clipboard.writeText(text);
    toast.success('Copied to clipboard');
  };

  return (
    <div>
      <PageHeader
        title="Design System"
        subtitle="Neural Dark — living style guide for NeuralOps"
        actions={
          <Select label="Theme" value={theme} onChange={(e) => setTheme(e.target.value as 'dark' | 'light')}>
            <option value="dark">Dark (default)</option>
            <option value="light">Light (preview)</option>
          </Select>
        }
      />

      <section style={{ marginBottom: 48 }}>
        <h2 style={{ fontFamily: 'var(--font-display)' }}>Colors</h2>
        <div className="design-grid">
          {COLORS.map(({ name, var: v }) => (
            <div key={name}>
              <div className="design-swatch" style={{ background: `var(${v})` }} />
              <code style={{ fontSize: 11 }}>{name}</code>
              <Button variant="ghost" size="sm" onClick={() => copyClasses(`var(${v})`)}>
                Copy
              </Button>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: 48 }}>
        <h2 style={{ fontFamily: 'var(--font-display)' }}>Typography</h2>
        <p style={{ fontFamily: 'var(--font-display)', fontSize: 48, fontWeight: 700, margin: '8px 0' }}>
          Display XL — NeuralOps Platform
        </p>
        <p style={{ fontFamily: 'var(--font-display)', fontSize: 28, fontWeight: 600 }}>Heading 1</p>
        <p style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 600 }}>Heading 2</p>
        <p style={{ fontSize: 16 }}>Body Large — Inter Regular 16px</p>
        <p style={{ fontSize: 14 }}>Body — Inter Regular 14px</p>
        <p style={{ fontSize: 12, textTransform: 'uppercase', fontWeight: 600, letterSpacing: '0.05em' }}>Label</p>
        <code style={{ fontFamily: 'var(--font-mono)', fontSize: 13 }}>
          Mono — JetBrains Mono 13px log line
        </code>
      </section>

      <section style={{ marginBottom: 48 }}>
        <h2 style={{ fontFamily: 'var(--font-display)' }}>Icons</h2>
        <div className="design-grid">
          {ICONS.map(({ Icon, name }) => (
            <div key={name} style={{ textAlign: 'center' }}>
              <Icon size={24} />
              <p style={{ fontSize: 11, color: 'var(--text-muted)' }}>{name}</p>
            </div>
          ))}
        </div>
      </section>

      <section style={{ marginBottom: 48 }}>
        <h2 style={{ fontFamily: 'var(--font-display)' }}>Components</h2>
        <div className="dashboard-row-4">
          <Card title="Badges">
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              <Badge variant="critical">CRITICAL</Badge>
              <Badge variant="p1">P1</Badge>
              <Badge variant="p2">P2</Badge>
              <Badge variant="healthy">Healthy</Badge>
              <Badge variant="degraded">Degraded</Badge>
            </div>
            <div style={{ marginTop: 16 }}>
              <StatusDot status="healthy" pulse label="Live pulse" />
            </div>
          </Card>
          <MetricCard label="Metric Card" value="99.2%" trend={-2.1} sparkline={[1, 2, 3, 2, 4, 3, 5]} />
          <Card title="Forms">
            <SearchInput placeholder="Search…" />
            <Select label="Environment" style={{ marginTop: 12 }}>
              <option>PROD</option>
            </Select>
          </Card>
        </div>

        <div className="dashboard-row-2" style={{ marginTop: 16 }}>
          <Card title="Log severity bars">
            {['INFO', 'WARN', 'ERROR', 'FATAL'].map((sev) => (
              <div key={sev} className={`severity-bar severity-bar--${sev}`} style={{ padding: 8, marginBottom: 8 }}>
                [{sev}] Connection timeout to payment-db
              </div>
            ))}
          </Card>
          <Card title="Charts">
            <ResponsiveContainer width="100%" height={120}>
              <AreaChart data={chartData}>
                <XAxis dataKey="t" stroke="var(--text-muted)" fontSize={10} />
                <YAxis stroke="var(--text-muted)" fontSize={10} />
                <Area type="monotone" dataKey="v" stroke="var(--accent-primary)" fill="var(--accent-glow)" />
              </AreaChart>
            </ResponsiveContainer>
            <ResponsiveContainer width="100%" height={120}>
              <BarChart data={chartData}>
                <Bar dataKey="v" fill="var(--accent-secondary)" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </div>

        <Card title="Code block" style={{ marginTop: 16 }}>
          <CodeBlock code={'level=ERROR service=upi-routing msg="thread pool exhausted"'} />
        </Card>

        <Card title="Feedback" style={{ marginTop: 16 }}>
          <Skeleton height={48} />
          <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
            <Button variant="primary" onClick={() => toast.success('Success toast')}>Success</Button>
            <Button variant="danger" onClick={() => toast.error('Error toast')}>Error</Button>
            <Button variant="secondary" onClick={() => toast('Info toast')}>Info</Button>
          </div>
        </Card>
      </section>

      <section>
        <h2 style={{ fontFamily: 'var(--font-display)' }}>Layout Patterns</h2>
        <div className="dashboard-grid">
          <Card title="Dashboard grid">
            <Grid size={16} /> 12-col responsive grid with 24px gutter
          </Card>
          <Card title="Log explorer split">
            <Terminal size={16} /> 300px filters · flex main · 400px detail
          </Card>
          <Card title="Incident focus">
            <AlertTriangle size={16} /> Full-page incident detail without sidebar
          </Card>
        </div>
      </section>
    </div>
  );
}
