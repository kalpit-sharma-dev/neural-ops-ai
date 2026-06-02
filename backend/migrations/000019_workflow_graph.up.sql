-- Persist the full branching workflow graph (nodes + edges + canvas positions)
-- authored in the visual editor. The existing `steps` column continues to hold
-- the linearized execution order derived from this graph.
ALTER TABLE observability_workflows
    ADD COLUMN IF NOT EXISTS graph JSONB NOT NULL DEFAULT '{}'::jsonb;
