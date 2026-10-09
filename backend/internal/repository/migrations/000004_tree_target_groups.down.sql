-- ============================================================================
-- 000004 Down: Revert Tree Target Groups Schema
-- ============================================================================

DROP INDEX IF EXISTS idx_target_groups_parent_id;

ALTER TABLE target_groups
    DROP COLUMN IF EXISTS color_tag,
    DROP COLUMN IF EXISTS parent_id;

