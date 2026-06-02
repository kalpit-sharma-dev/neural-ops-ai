import { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import toast from 'react-hot-toast';
import {
  fetchIntegrations,
  fetchMarketplace,
  installExtension,
  uninstallExtension,
  type MarketplaceExtension,
} from '../api/observability';
import { getApiErrorMessage } from '../api/client';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/PageStates';
import { StitchPageShell } from '../components/stitch';

const ALL = 'all';

function categoryLabel(category: string): string {
  switch (category) {
    case 'visualization':
      return 'Visualization';
    case 'datasource':
      return 'Data source';
    case 'app':
      return 'App';
    default:
      return category;
  }
}

export default function Marketplace() {
  const queryClient = useQueryClient();
  const { data: extensions, isLoading } = useQuery({
    queryKey: ['marketplace'],
    queryFn: fetchMarketplace,
  });
  const { data: integrations } = useQuery({ queryKey: ['integrations'], queryFn: fetchIntegrations });

  const [category, setCategory] = useState<string>(ALL);
  const [search, setSearch] = useState('');
  const [active, setActive] = useState<MarketplaceExtension | null>(null);
  const [form, setForm] = useState<Record<string, string>>({});

  const categories = useMemo(() => {
    const set = new Set<string>();
    (extensions ?? []).forEach((e) => set.add(e.category));
    return [ALL, ...Array.from(set).sort()];
  }, [extensions]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return (extensions ?? []).filter((e) => {
      const matchesCategory = category === ALL || e.category === category;
      const matchesSearch =
        q === '' || e.name.toLowerCase().includes(q) || e.description.toLowerCase().includes(q);
      return matchesCategory && matchesSearch;
    });
  }, [extensions, category, search]);

  const installMut = useMutation({
    mutationFn: ({ key, config }: { key: string; config: Record<string, string> }) =>
      installExtension(key, config),
    onSuccess: (_d, vars) => {
      const wasInstalled = (extensions ?? []).find((e) => e.key === vars.key)?.installed;
      toast.success(wasInstalled ? 'Extension reconfigured' : 'Extension installed');
      closeModal();
      void queryClient.invalidateQueries({ queryKey: ['marketplace'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const uninstallMut = useMutation({
    mutationFn: uninstallExtension,
    onSuccess: () => {
      toast.success('Extension uninstalled');
      void queryClient.invalidateQueries({ queryKey: ['marketplace'] });
    },
    onError: (e) => toast.error(getApiErrorMessage(e)),
  });

  const closeModal = () => {
    setActive(null);
    setForm({});
  };

  const openInstall = (ext: MarketplaceExtension) => {
    // No configuration required: install in one click.
    if (!ext.configFields || ext.configFields.length === 0) {
      installMut.mutate({ key: ext.key, config: {} });
      return;
    }
    setActive(ext);
    // Prefill non-secret fields from the masked public config when reconfiguring.
    const prefill: Record<string, string> = {};
    ext.configFields.forEach((f) => {
      prefill[f.key] = !f.secret ? (ext.configPublic?.[f.key] ?? '') : '';
    });
    setForm(prefill);
  };

  const submit = () => {
    if (!active) return;
    installMut.mutate({ key: active.key, config: form });
  };

  return (
    <StitchPageShell
      title="Marketplace"
      subtitle="Apps, extensions and integration catalog"
      actions={<Link to="/integrations">Manage connections →</Link>}
    >
      {isLoading && <LoadingState />}

      <div className="marketplace-toolbar">
        <div className="marketplace-tabs" role="tablist" aria-label="Categories">
          {categories.map((c) => (
            <button
              key={c}
              type="button"
              role="tab"
              aria-selected={category === c}
              className={`marketplace-tab${category === c ? ' marketplace-tab--active' : ''}`}
              onClick={() => setCategory(c)}
            >
              {c === ALL ? 'All' : categoryLabel(c)}
            </button>
          ))}
        </div>
        <input
          className="marketplace-search"
          type="search"
          placeholder="Search extensions…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Search extensions"
        />
      </div>

      <div className="settings-grid">
        {filtered.map((ext) => (
          <Card key={ext.key} title={ext.name}>
            <div className="marketplace-card__badges">
              <Badge variant="info">{categoryLabel(ext.category)}</Badge>
              {ext.installed && <Badge variant="healthy">Installed</Badge>}
            </div>
            <p className="muted marketplace-card__meta">
              {ext.publisher} · v{ext.version}
            </p>
            <p className="marketplace-card__desc">{ext.description}</p>
            {ext.installed && ext.configPublic && Object.keys(ext.configPublic).length > 0 && (
              <ul className="muted marketplace-card__config">
                {Object.entries(ext.configPublic).map(([k, v]) => (
                  <li key={k}>
                    {k}: {v}
                  </li>
                ))}
              </ul>
            )}
            <div className="marketplace-card__actions">
              {ext.installed ? (
                <>
                <Button variant="secondary" size="sm" onClick={() => openInstall(ext)} data-testid={`marketplace-configure-${ext.key}`}>
                    Configure
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    disabled={uninstallMut.isPending}
                    data-testid={`marketplace-uninstall-${ext.key}`}
                    onClick={() => {
                      if (window.confirm(`Uninstall "${ext.name}"?`)) uninstallMut.mutate(ext.key);
                    }}
                  >
                    Uninstall
                  </Button>
                </>
              ) : (
                <Button
                  variant="primary"
                  size="sm"
                  disabled={installMut.isPending}
                  data-testid={`marketplace-install-${ext.key}`}
                  onClick={() => openInstall(ext)}
                >
                  Install
                </Button>
              )}
            </div>
          </Card>
        ))}
        {!isLoading && filtered.length === 0 && (
          <p className="muted">No extensions match your filters.</p>
        )}
      </div>

      <h3 className="marketplace-section-heading">Integrations</h3>
      <div className="settings-grid">
        {(integrations ?? []).map((i) => (
          <Card key={i.id} title={i.name}>
            <Badge variant={i.connected ? 'healthy' : 'info'}>
              {i.connected ? 'Connected' : 'Available'}
            </Badge>
            <p className="muted">{i.type}</p>
            <div className="marketplace-card__actions">
              <Link to="/integrations">{i.connected ? 'Manage' : 'Connect'}</Link>
            </div>
          </Card>
        ))}
      </div>

      {active && (
        <div className="modal-overlay" role="dialog" aria-modal="true">
          <Card
            title={active.installed ? `Configure ${active.name}` : `Install ${active.name}`}
            style={{ maxWidth: 480, margin: '10vh auto' }}
          >
            <p className="muted marketplace-card__meta">
              {active.publisher} · v{active.version}
            </p>
            <p className="marketplace-card__desc">{active.description}</p>
            <div className="form-stack">
              {(active.configFields ?? []).map((f) => (
                <label key={f.key} className="ui-field">
                  <span className="ui-field__label">{f.label}</span>
                  <input
                    className="ui-input"
                    type={f.secret ? 'password' : 'text'}
                    value={form[f.key] ?? ''}
                    placeholder={f.secret && active.installed ? 'Leave blank to keep current' : ''}
                    onChange={(e) => setForm((prev) => ({ ...prev, [f.key]: e.target.value }))}
                  />
                </label>
              ))}
              <div style={{ display: 'flex', gap: 8 }}>
                <Button variant="primary" disabled={installMut.isPending} onClick={submit}>
                  {active.installed ? 'Save changes' : 'Install'}
                </Button>
                <Button variant="secondary" onClick={closeModal}>
                  Cancel
                </Button>
              </div>
            </div>
          </Card>
        </div>
      )}
    </StitchPageShell>
  );
}
