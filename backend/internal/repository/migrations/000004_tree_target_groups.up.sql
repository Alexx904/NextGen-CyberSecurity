-- ============================================================================
-- 000004: Hierarchical Tree Folders for Target Groups
-- ============================================================================

ALTER TABLE target_groups
    ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES target_groups(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS color_tag VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_target_groups_parent_id ON target_groups(parent_id);

