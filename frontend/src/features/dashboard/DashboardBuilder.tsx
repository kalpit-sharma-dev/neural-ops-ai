import { useCallback, useState } from 'react';
import type { Dashboard } from '../../api/observability';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';

interface DashboardBuilderProps {
  dashboard: Dashboard;
  onSave: (tiles: Dashboard['tiles']) => void;
  saving?: boolean;
}

const GRID_COLS = 12;

export function DashboardBuilder({ dashboard, onSave, saving }: DashboardBuilderProps) {
  const [tiles, setTiles] = useState(dashboard.tiles);
  const [dragId, setDragId] = useState<string | null>(null);

  const moveTile = useCallback((targetId: string) => {
    if (!dragId || dragId === targetId) return;
    setTiles((prev) => {
      const from = prev.findIndex((t) => t.id === dragId);
      const to = prev.findIndex((t) => t.id === targetId);
      if (from < 0 || to < 0) return prev;
      const next = [...prev];
      const [item] = next.splice(from, 1);
      next.splice(to, 0, item);
      return next.map((t, i) => ({
        ...t,
        position: { ...t.position, y: Math.floor(i / 2) * 4, x: (i % 2) * 6 },
      }));
    });
    setDragId(null);
  }, [dragId]);

  const addTile = () => {
    const id = `tile-${Date.now()}`;
    setTiles((prev) => [
      ...prev,
      {
        id,
        type: 'metric',
        title: 'New metric',
        metric: 'throughput',
        position: { x: 0, y: Math.ceil(prev.length / 2) * 4, w: 6, h: 4 },
      },
    ]);
  };

  return (
    <div className="dashboard-builder">
      <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
        <Button variant="secondary" size="sm" onClick={addTile}>Add tile</Button>
        <Button variant="primary" size="sm" disabled={saving} onClick={() => onSave(tiles)}>
          Save layout
        </Button>
      </div>
      <div className="dashboard-builder__grid" style={{ gridTemplateColumns: `repeat(${GRID_COLS}, 1fr)` }}>
        {tiles.map((tile) => (
          <div
            key={tile.id}
            className="dashboard-builder__tile"
            draggable
            onDragStart={() => setDragId(tile.id)}
            onDragOver={(e) => e.preventDefault()}
            onDrop={() => moveTile(tile.id)}
            style={{
              gridColumn: `span ${tile.position.w ?? 6}`,
              gridRow: `span ${tile.position.h ?? 4}`,
            }}
          >
            <Card title={tile.title}>
              <p className="muted">{tile.type} · {tile.metric ?? tile.query ?? 'custom'}</p>
              <p className="muted" style={{ fontSize: 11 }}>Drag to reorder</p>
            </Card>
          </div>
        ))}
      </div>
    </div>
  );
}
